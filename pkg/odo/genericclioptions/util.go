package genericclioptions

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/log"
	pkgUtil "github.com/redhat-developer/odo/pkg/util"
	v1 "k8s.io/api/core/v1"

	dfutil "github.com/devfile/library/v2/pkg/util"
)

const (
	gitDirName = ".git"
)

// ApplyIgnore will take the current ignores []string and append the mandatory odo-file-index.json and
// .git ignores; or find the .odoignore/.gitignore file in the directory and use that instead.
func ApplyIgnore(ignores *[]string, sourcePath string) (err error) {
	if len(*ignores) == 0 {
		rules, err := dfutil.GetIgnoreRulesFromDirectory(sourcePath)
		if err != nil {
			return err
		}
		*ignores = append(*ignores, rules...)
	}

	indexFile := pkgUtil.GetIndexFileRelativeToContext()
	// check if the ignores flag has the index file
	if !dfutil.In(*ignores, indexFile) {
		*ignores = append(*ignores, indexFile)
	}

	// check if the ignores flag has the git dir
	if !dfutil.In(*ignores, gitDirName) {
		*ignores = append(*ignores, gitDirName)
	}

	return nil
}

// WarnIfDefaultNamespace warns when user tries to run `odo dev` or `odo deploy` in the default namespace
func WarnIfDefaultNamespace(namespace string, kubeClient kclient.ClientInterface) {
	if namespace == v1.NamespaceDefault {
		noun := "namespace"
		if isOC, _ := kubeClient.IsProjectSupported(); isOC {
			noun = "project"
		}
		fmt.Println()
		log.Warningf("You are using \"default\" %[1]s, odo may not work as expected in the default %[1]s.", noun)
		log.Warningf("You may set a new %[1]s by running `odo create %[1]s <name>`, or set an existing one by running `odo set %[1]s <name>`", noun)
	}
}
