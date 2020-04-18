package sync

import (
	taro "archive/tar"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/openshift/odo/pkg/util"

	"github.com/golang/glog"
)

type SyncClient interface {
	ExecCMDInContainer(string, string, []string, io.Writer, io.Writer, io.Reader, bool) error
	ExtractProjectToComponent(string, string, string, io.Reader) error
}

// CopyFile copies localPath directory or list of files in copyFiles list to the directory in running Pod.
// copyFiles is list of changed files captured during `odo watch` as well as binary file path
// During copying binary components, localPath represent base directory path to binary and copyFiles contains path of binary
// During copying local source components, localPath represent base directory path whereas copyFiles is empty
// During `odo watch`, localPath represent base directory path whereas copyFiles contains list of changed Files
func CopyFile(client SyncClient, localPath string, targetPodName string, targetContainerName string, targetPath string, copyFiles []string, globExps []string) error {

	// Destination is set to "ToSlash" as all containers being ran within OpenShift / S2I are all
	// Linux based and thus: "\opt\app-root\src" would not work correctly.
	dest := filepath.ToSlash(filepath.Join(targetPath, filepath.Base(localPath)))
	targetPath = filepath.ToSlash(targetPath)

	glog.V(4).Infof("CopyFile arguments: localPath %s, dest %s, targetPath %s, copyFiles %s, globalExps %s", localPath, dest, targetPath, copyFiles, globExps)
	reader, writer := io.Pipe()
	// inspired from https://github.com/kubernetes/kubernetes/blob/master/pkg/kubectl/cmd/cp.go#L235
	go func() {
		defer writer.Close()

		err := makeTar(localPath, dest, writer, copyFiles, globExps)
		if err != nil {
			glog.Errorf("Error while creating tar: %#v", err)
			os.Exit(1)
		}

	}()

	err := client.ExtractProjectToComponent(targetPodName, targetContainerName, targetPath, reader)
	if err != nil {
		return err
	}

	return nil
}

// checkFileExist check if given file exists or not
func checkFileExist(fileName string) bool {
	_, err := os.Stat(fileName)
	return !os.IsNotExist(err)
}

// makeTar function is copied from https://github.com/kubernetes/kubernetes/blob/master/pkg/kubectl/cmd/cp.go#L309
// srcPath is ignored if files is set
func makeTar(srcPath, destPath string, writer io.Writer, files []string, globExps []string) error {
	// TODO: use compression here?
	tarWriter := taro.NewWriter(writer)
	defer tarWriter.Close()
	srcPath = filepath.Clean(srcPath)

	// "ToSlash" is used as all containers within OpenShift are Linux based
	// and thus \opt\app-root\src would be an invalid path. Backward slashes
	// are converted to forward.
	destPath = filepath.ToSlash(filepath.Clean(destPath))

	glog.V(4).Infof("makeTar arguments: srcPath: %s, destPath: %s, files: %+v", srcPath, destPath, files)
	if len(files) != 0 {
		//watchTar
		for _, fileName := range files {
			if checkFileExist(fileName) {
				// Fetch path of source file relative to that of source base path so that it can be passed to recursiveTar
				// which uses path relative to base path for taro header to correctly identify file location when untarred

				// Yes, now that the file exists, now we need to get the absolute path.. if we don't, then when we pass in:
				// 'odo push --context foobar' instead of 'odo push --context ~/foobar' it will NOT work..
				fileAbsolutePath, err := util.GetAbsPath(fileName)
				if err != nil {
					return err
				}
				glog.V(4).Infof("Got abs path: %s", fileAbsolutePath)
				glog.V(4).Infof("Making %s relative to %s", srcPath, fileAbsolutePath)

				// We use "FromSlash" to make this OS-based (Windows uses \, Linux & macOS use /)
				// we get the relative path by joining the two
				destFile, err := filepath.Rel(filepath.FromSlash(srcPath), filepath.FromSlash(fileAbsolutePath))
				if err != nil {
					return err
				}

				// Now we get the source file and join it to the base directory.
				srcFile := filepath.Join(filepath.Base(srcPath), destFile)

				glog.V(4).Infof("makeTar srcFile: %s", srcFile)
				glog.V(4).Infof("makeTar destFile: %s", destFile)

				// The file could be a regular file or even a folder, so use recursiveTar which handles symlinks, regular files and folders
				err = recursiveTar(filepath.Dir(srcPath), srcFile, filepath.Dir(destPath), destFile, tarWriter, globExps)
				if err != nil {
					return err
				}
			}
		}
	} else {
		return recursiveTar(filepath.Dir(srcPath), filepath.Base(srcPath), filepath.Dir(destPath), "", tarWriter, globExps)
	}

	return nil
}

// recursiveTar function is copied from https://github.com/kubernetes/kubernetes/blob/master/pkg/kubectl/cmd/cp.go#L319
func recursiveTar(srcBase, srcFile, destBase, destFile string, tw *taro.Writer, globExps []string) error {
	glog.V(4).Infof("recursiveTar arguments: srcBase: %s, srcFile: %s, destBase: %s, destFile: %s", srcBase, srcFile, destBase, destFile)

	// The destination is a LINUX container and thus we *must* use ToSlash in order
	// to get the copying over done correctly..
	destBase = filepath.ToSlash(destBase)
	destFile = filepath.ToSlash(destFile)
	glog.V(4).Infof("Corrected destinations: base: %s file: %s", destBase, destFile)

	joinedPath := filepath.Join(srcBase, srcFile)
	matchedPathsDir, err := filepath.Glob(joinedPath)
	if err != nil {
		return err
	}

	matchedPaths := []string{}

	// checking the files which are allowed by glob matching
	for _, path := range matchedPathsDir {
		matched, err := util.IsGlobExpMatch(path, globExps)
		if err != nil {
			return err
		}
		if !matched {
			matchedPaths = append(matchedPaths, path)
		}
	}

	// adding the files for taring
	for _, matchedPath := range matchedPaths {
		stat, err := os.Lstat(matchedPath)
		if err != nil {
			return err
		}
		if stat.IsDir() {
			files, err := ioutil.ReadDir(matchedPath)
			if err != nil {
				return err
			}
			if len(files) == 0 {
				//case empty directory
				hdr, _ := taro.FileInfoHeader(stat, matchedPath)
				hdr.Name = destFile
				if err := tw.WriteHeader(hdr); err != nil {
					return err
				}
			}
			for _, f := range files {
				if err := recursiveTar(srcBase, filepath.Join(srcFile, f.Name()), destBase, filepath.Join(destFile, f.Name()), tw, globExps); err != nil {
					return err
				}
			}
			return nil
		} else if stat.Mode()&os.ModeSymlink != 0 {
			//case soft link
			hdr, _ := taro.FileInfoHeader(stat, joinedPath)
			target, err := os.Readlink(joinedPath)
			if err != nil {
				return err
			}

			hdr.Linkname = target
			hdr.Name = destFile
			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}
		} else {
			//case regular file or other file type like pipe
			hdr, err := taro.FileInfoHeader(stat, joinedPath)
			if err != nil {
				return err
			}
			hdr.Name = destFile

			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}

			f, err := os.Open(joinedPath)
			if err != nil {
				return err
			}
			defer f.Close() // #nosec G307

			if _, err := io.Copy(tw, f); err != nil {
				return err
			}

			return f.Close()
		}
	}

	return nil
}
