package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/devfile/library/pkg/devfile/generator"
	dfutil "github.com/devfile/library/pkg/util"

	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/remotecmd"
	"github.com/redhat-developer/odo/pkg/util"

	"k8s.io/klog"
)

// Adapter is a component adapter implementation for sync
type Adapter struct {
	kubeClient    kclient.ClientInterface
	SyncExtracter SyncExtracter
}

// New instantiates a component adapter
func New(syncClient SyncExtracter, kubeClient kclient.ClientInterface) Adapter {
	return Adapter{
		kubeClient:    kubeClient,
		SyncExtracter: syncClient,
	}
}

// ComponentInfo is a struct that holds information about a component i.e.; pod name, container name, and source mount (if applicable)
type ComponentInfo struct {
	ComponentName string
	PodName       string
	ContainerName string
	SyncFolder    string
}

// SyncParameters is a struct containing the parameters to be used when syncing a devfile component
type SyncParameters struct {
	Path                     string   // Path refers to the parent folder containing the source code to push up to a component
	WatchFiles               []string // Optional: WatchFiles is the list of changed files detected by odo watch. If empty or nil, odo will check .odo/odo-file-index.json to determine changed files
	WatchDeletedFiles        []string // Optional: WatchDeletedFiles is the list of deleted files detected by odo watch. If empty or nil, odo will check .odo/odo-file-index.json to determine deleted files
	IgnoredFiles             []string // IgnoredFiles is the list of files to not push up to a component
	ForceBuild               bool     // ForceBuild determines whether or not to push all of the files up to a component or just files that have changed, added or removed.
	DevfileScanIndexForWatch bool     // DevfileScanIndexForWatch is true if watch's push should regenerate the index file during SyncFiles, false otherwise. See 'pkg/sync/adapter.go' for details

	CompInfo        ComponentInfo
	PodChanged      bool
	ComponentExists bool
	Files           map[string]string
}

// SyncFiles does a couple of things:
// if files changed/deleted are passed in from watch, it syncs them to the component
// otherwise, it checks which files have changed and syncs the delta
// it returns a boolean execRequired and an error. execRequired tells us if files have
// changed and devfile execution is required
func (a Adapter) SyncFiles(syncParameters SyncParameters) (bool, error) {

	// Whether to write the indexer content to the index file path (resolvePath)
	forceWrite := false

	// Ret from Indexer function
	var ret util.IndexerRet

	var deletedFiles []string
	var changedFiles []string
	isForcePush := syncParameters.ForceBuild || !syncParameters.ComponentExists || syncParameters.PodChanged
	isWatch := len(syncParameters.WatchFiles) > 0 || len(syncParameters.WatchDeletedFiles) > 0

	// When this function is invoked by watch, the logic is:
	// 1) If this is the first time that watch has called Push (in this OS process), then generate the file index
	//    using the file indexer, and use that to sync files (eg don't use changed/deleted files list from watch at
	//    this stage; these will be found by the indexer run).
	//    - In the watch scenario, we need to first run the indexer for two reasons:
	// 	    - In cases where the index doesn't initially exist, we need to create it (so we can ADD to it in
	//        later calls to SyncFiles(...) )
	// 	    - Even if it does initially exist, there is no guarantee that the remote pod is consistent with it; so
	//        on our first invocation we need to compare the index with the remote pod (by regenerating the index
	//        and using the changes files list from that to sync the results.)
	//
	// 2) For every other push/sync call after the first, don't run the file indexer, instead we use
	//    the watch events to determine what changed, and ensure that the index is then updated based
	//    on the watch events (to ensure future calls are correct)

	// True if the index was updated based on the deleted/changed files values from the watch (and
	// thus the indexer doesn't need to run), false otherwise
	indexRegeneratedByWatch := false

	// If watch files are specified _and_ this is not the first call (by this process) to SyncFiles by the watch command, then insert the
	// changed files into the existing file index, and delete removed files from the index
	if isWatch && !syncParameters.DevfileScanIndexForWatch {

		err := updateIndexWithWatchChanges(syncParameters)

		if err != nil {
			return false, err
		}

		changedFiles = syncParameters.WatchFiles
		deletedFiles = syncParameters.WatchDeletedFiles
		deletedFiles, err = dfutil.RemoveRelativePathFromFiles(deletedFiles, syncParameters.Path)
		if err != nil {
			return false, fmt.Errorf("unable to remove relative path from list of changed/deleted files: %w", err)
		}
		indexRegeneratedByWatch = true

	}

	if !indexRegeneratedByWatch {
		// Calculate the files to sync
		// Tries to sync the deltas unless it is a forced push
		// if it is a forced push (isForcePush) reset the index to do a full sync
		absIgnoreRules := dfutil.GetAbsGlobExps(syncParameters.Path, syncParameters.IgnoredFiles)

		// Before running the indexer, make sure the .odo folder exists (or else the index file will not get created)
		odoFolder := filepath.Join(syncParameters.Path, ".odo")
		if _, err := os.Stat(odoFolder); os.IsNotExist(err) {
			err = os.Mkdir(odoFolder, 0750)
			if err != nil {
				return false, fmt.Errorf("unable to create directory: %w", err)
			}
		}

		// If the pod changed, reset the index, which will cause the indexer to walk the directory
		// tree and resync all local files.
		// If it is a new component, reset index to make sure any previously existing file is cleaned up
		if syncParameters.PodChanged || !syncParameters.ComponentExists {
			err := util.DeleteIndexFile(syncParameters.Path)
			if err != nil {
				return false, fmt.Errorf("unable to reset the index file: %w", err)
			}
		}

		// Run the indexer and find the modified/added/deleted/renamed files
		var err error
		ret, err = util.RunIndexerWithRemote(syncParameters.Path, syncParameters.IgnoredFiles, syncParameters.Files)

		if err != nil {
			return false, fmt.Errorf("unable to run indexer: %w", err)
		}

		if len(ret.FilesChanged) > 0 || len(ret.FilesDeleted) > 0 {
			forceWrite = true
		}

		// apply the glob rules from the .gitignore/.odoignore file
		// and ignore the files on which the rules apply and filter them out
		filesChangedFiltered, filesDeletedFiltered := dfutil.FilterIgnores(ret.FilesChanged, ret.FilesDeleted, absIgnoreRules)

		deletedFiles = append(filesDeletedFiltered, ret.RemoteDeleted...)
		deletedFiles = append(deletedFiles, ret.RemoteDeleted...)
		klog.V(4).Infof("List of files to be deleted: +%v", deletedFiles)
		changedFiles = filesChangedFiltered
		klog.V(4).Infof("List of files changed: +%v", changedFiles)

		if len(filesChangedFiltered) == 0 && len(filesDeletedFiltered) == 0 && !isForcePush {
			return false, nil
		}

		if isForcePush {
			deletedFiles = append(deletedFiles, "*")
		}
	}

	err := a.pushLocal(syncParameters.Path,
		changedFiles,
		deletedFiles,
		isForcePush,
		syncParameters.IgnoredFiles,
		syncParameters.CompInfo,
		ret,
	)
	if err != nil {
		return false, fmt.Errorf("failed to sync to component with name %s: %w", syncParameters.CompInfo.ComponentName, err)
	}
	if forceWrite {
		err = util.WriteFile(ret.NewFileMap, ret.ResolvedPath)
		if err != nil {
			return false, fmt.Errorf("failed to write file: %w", err)
		}
	}

	return true, nil
}

// pushLocal syncs source code from the user's disk to the component
func (a Adapter) pushLocal(path string, files []string, delFiles []string, isForcePush bool, globExps []string, compInfo ComponentInfo, ret util.IndexerRet) error {
	klog.V(4).Infof("Push: componentName: %s, path: %s, files: %s, delFiles: %s, isForcePush: %+v", compInfo.ComponentName, path, files, delFiles, isForcePush)

	// Edge case: check to see that the path is NOT empty.
	emptyDir, err := dfutil.IsEmpty(path)
	if err != nil {
		return fmt.Errorf("unable to check directory: %s: %w", path, err)
	} else if emptyDir {
		return fmt.Errorf("directory/file %s is empty", path)
	}

	// Sync the files to the pod
	syncFolder := compInfo.SyncFolder

	if syncFolder != generator.DevfileSourceVolumeMount {
		// Need to make sure the folder already exists on the component or else sync will fail
		klog.V(4).Infof("Creating %s on the remote container if it doesn't already exist", syncFolder)
		cmdArr := getCmdToCreateSyncFolder(syncFolder)

		_, _, err = remotecmd.ExecuteCommand(cmdArr, a.kubeClient, compInfo.PodName, compInfo.ContainerName, false, nil, nil)
		if err != nil {
			return err
		}
	}
	// If there were any files deleted locally, delete them remotely too.
	if len(delFiles) > 0 {
		cmdArr := getCmdToDeleteFiles(delFiles, syncFolder)

		_, _, err = remotecmd.ExecuteCommand(cmdArr, a.kubeClient, compInfo.PodName, compInfo.ContainerName, false, nil, nil)
		if err != nil {
			return err
		}
	}

	if !isForcePush {
		if len(files) == 0 && len(delFiles) == 0 {
			return nil
		}
	}

	if isForcePush || len(files) > 0 {
		klog.V(4).Infof("Copying files %s to pod", strings.Join(files, " "))
		err = CopyFile(a.SyncExtracter, path, compInfo, syncFolder, files, globExps, ret)
		if err != nil {
			return fmt.Errorf("unable push files to pod: %w", err)
		}
	}

	return nil
}

// updateIndexWithWatchChanges uses the pushParameters.WatchDeletedFiles and pushParamters.WatchFiles to update
// the existing index file; the index file is required to exist when this function is called.
func updateIndexWithWatchChanges(syncParameters SyncParameters) error {
	indexFilePath, err := util.ResolveIndexFilePath(syncParameters.Path)

	if err != nil {
		return fmt.Errorf("unable to resolve path: %s: %w", syncParameters.Path, err)
	}

	// Check that the path exists
	_, err = os.Stat(indexFilePath)
	if err != nil {
		// This shouldn't happen: in the watch case, SyncFiles should first be called with 'DevfileScanIndexForWatch' set to true, which
		// will generate the index. Then, all subsequent invocations of SyncFiles will run with 'DevfileScanIndexForWatch' set to false,
		// which will not regenerate the index (merely updating it based on changed files.)
		//
		// If you see this error it means somehow watch's SyncFiles was called without the index being first generated (likely because the
		// above mentioned pushParam wasn't set). See SyncFiles(...) for details.
		return fmt.Errorf("resolved path doesn't exist: %s: %w", indexFilePath, err)
	}

	// Parse the existing index
	fileIndex, err := util.ReadFileIndex(indexFilePath)
	if err != nil {
		return fmt.Errorf("unable to read index from path: %s: %w", indexFilePath, err)
	}

	rootDir := syncParameters.Path

	// Remove deleted files from the existing index
	for _, deletedFile := range syncParameters.WatchDeletedFiles {

		relativePath, err := util.CalculateFileDataKeyFromPath(deletedFile, rootDir)

		if err != nil {
			klog.V(4).Infof("Error occurred for %s: %v", deletedFile, err)
			continue
		}
		delete(fileIndex.Files, relativePath)
		klog.V(4).Infof("Removing watch deleted file from index: %s", relativePath)
	}

	// Add changed files to the existing index
	for _, addedOrModifiedFile := range syncParameters.WatchFiles {
		relativePath, fileData, err := util.GenerateNewFileDataEntry(addedOrModifiedFile, rootDir)

		if err != nil {
			klog.V(4).Infof("Error occurred for %s: %v", addedOrModifiedFile, err)
			continue
		}
		fileIndex.Files[relativePath] = *fileData
		klog.V(4).Infof("Added/updated watched file in index: %s", relativePath)
	}

	// Write the result
	return util.WriteFile(fileIndex.Files, indexFilePath)

}

// getCmdToCreateSyncFolder returns the command used to create the remote sync folder on the running container
func getCmdToCreateSyncFolder(syncFolder string) []string {
	return []string{"mkdir", "-p", syncFolder}
}

// getCmdToDeleteFiles returns the command used to delete the remote files on the container that are marked for deletion
func getCmdToDeleteFiles(delFiles []string, syncFolder string) []string {
	rmPaths := dfutil.GetRemoteFilesMarkedForDeletion(delFiles, syncFolder)
	klog.V(4).Infof("remote files marked for deletion are %+v", rmPaths)
	cmdArr := []string{"rm", "-rf"}

	for _, remote := range rmPaths {
		cmdArr = append(cmdArr, filepath.ToSlash(remote))
	}
	return cmdArr
}
