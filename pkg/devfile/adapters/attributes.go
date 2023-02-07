package adapters

import (
	"path/filepath"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

const _devPushPathAttributePrefix = "dev.odo.push.path:"

// GetSyncFilesFromAttributes gets the target files and folders along with their respective remote destination from the devfile.
// It uses the "dev.odo.push.path:" attribute prefix, if any, in the specified command.
func GetSyncFilesFromAttributes(command v1alpha2.Command) map[string]string {
	syncMap := make(map[string]string)
	for key, value := range command.Attributes.Strings(nil) {
		if strings.HasPrefix(key, _devPushPathAttributePrefix) {
			localValue := strings.ReplaceAll(key, _devPushPathAttributePrefix, "")
			syncMap[filepath.Clean(localValue)] = filepath.ToSlash(filepath.Clean(value))
		}
	}
	return syncMap
}
