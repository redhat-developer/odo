package devfile

import (
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/tests/helper"
	"gopkg.in/yaml.v2"
)

var _ = Describe("odo devfile init command tests", func() {

	var commonVar helper.CommonVar

	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
	})

	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	It("should fail", func() {
		By("running odo init with incomplete flags", func() {
			helper.Cmd("odo", "init", "--name", "aname").ShouldFail()
		})

		By("keeping an empty directory when running odo init with wrong starter name", func() {
			helper.Cmd("odo", "init", "--name", "aname", "--devfile", "go", "--starter", "wrongname").ShouldFail()
			files := helper.ListFilesInDir(commonVar.Context)
			Expect(len(files)).To(Equal(0))
		})
		By("running odo init in a directory containing a devfile.yaml", func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-registry.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			err := helper.Cmd("odo", "init").ShouldFail().Err()
			Expect(err).To(ContainSubstring("a devfile already exists in the current directory"))
		})
		By("running odo init in a directory containing a .devfile.yaml", func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-registry.yaml"), filepath.Join(commonVar.Context, ".devfile.yaml"))
			err := helper.Cmd("odo", "init").ShouldFail().Err()
			Expect(err).To(ContainSubstring("a devfile already exists in the current directory"))
		})
	})

	When("running odo init with valid flags", func() {
		BeforeEach(func() {
			helper.Cmd("odo", "init", "--name", "aname", "--devfile", "go").ShouldPass()
		})

		It("should download a devfile.yaml file", func() {
			files := helper.ListFilesInDir(commonVar.Context)
			Expect(files).To(Equal([]string{"devfile.yaml"}))
		})
	})

	When("running odo init from a directory with sources", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
		})
		It("should work without --starter flag", func() {
			helper.Cmd("odo", "init", "--name", "aname", "--devfile", "nodejs").ShouldPass()
		})
		It("should not accept --starter flag", func() {
			err := helper.Cmd("odo", "init", "--name", "aname", "--devfile", "nodejs", "--starter", "nodejs-starter").ShouldFail().Err()
			Expect(err).To(ContainSubstring("--starter parameter cannot be used when the directory is not empty"))
		})
	})

	When("devfile contains parent URI", func() {
		var originalKeyList []string
		var srcDevfile string

		BeforeEach(func() {
			var err error
			srcDevfile = helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-parent.yaml")
			originalDevfileContent, err := ioutil.ReadFile(srcDevfile)
			Expect(err).To(BeNil())
			var content map[string]interface{}
			Expect(yaml.Unmarshal(originalDevfileContent, &content)).To(BeNil())
			for k := range content {
				originalKeyList = append(originalKeyList, k)
			}
		})

		It("should not replace the original devfile", func() {
			helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", srcDevfile).ShouldPass()
			devfileContent, err := ioutil.ReadFile(filepath.Join(commonVar.Context, "devfile.yaml"))
			Expect(err).To(BeNil())
			var content map[string]interface{}
			Expect(yaml.Unmarshal(devfileContent, &content)).To(BeNil())
			for k := range content {
				Expect(k).To(BeElementOf(originalKeyList))
			}
		})
	})
})
