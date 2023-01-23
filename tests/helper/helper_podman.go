package helper

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/gomega"
)

func GenerateAndSetContainersConf(dir string) {
	ns := GetProjectName()
	containersConfPath := filepath.Join(dir, "containers.conf")
	err := CreateFileWithContent(containersConfPath, fmt.Sprintf(`
[engine]
namespace=%q
`, ns))
	Expect(err).ShouldNot(HaveOccurred())
	os.Setenv("CONTAINERS_CONF", containersConfPath)
}
