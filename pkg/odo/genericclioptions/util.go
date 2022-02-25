package genericclioptions

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	pkgUtil "github.com/redhat-developer/odo/pkg/util"

	dfutil "github.com/devfile/library/pkg/util"
)

// checkProjectCreateOrDeleteOnlyOnInvalidNamespace errors out if user is trying to create or delete something other than project
// errFormatForCommand must contain one %s
func checkProjectCreateOrDeleteOnlyOnInvalidNamespace(cmdline cmdline.Cmdline, errFormatForCommand string) error {
	// do not error out when its odo delete -a, so that we let users delete the local config on missing namespace
	if cmdline.GetParentName() != "project" && (cmdline.GetName() == "create" || (cmdline.GetName() == "v2delete" && !cmdline.IsFlagSet("all"))) {
		return fmt.Errorf(errFormatForCommand, cmdline.GetRootName())
	}
	return nil
}

// checkProjectCreateOrDeleteOnlyOnInvalidNamespaceNoFmt errors out if user is trying to create or delete something other than project
// compare to checkProjectCreateOrDeleteOnlyOnInvalidNamespace, no %s is needed
func checkProjectCreateOrDeleteOnlyOnInvalidNamespaceNoFmt(cmdline cmdline.Cmdline, errFormatForCommand string) error {
	// do not error out when its odo delete -a, so that we let users delete the local config on missing namespace
	if cmdline.GetParentName() != "project" && (cmdline.GetName() == "create" || cmdline.GetName() == "push" || (cmdline.GetName() == "v2delete" && !cmdline.IsFlagSet("all"))) {
		return fmt.Errorf(errFormatForCommand)
	}
	return nil
}

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
