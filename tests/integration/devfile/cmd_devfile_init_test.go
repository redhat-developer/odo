package devfile

import (
	"io/ioutil"
	"os"
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
		Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeFalse())
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
			defer os.Remove(filepath.Join(commonVar.Context, "devfile.yaml"))
			err := helper.Cmd("odo", "init").ShouldFail().Err()
			Expect(err).To(ContainSubstring("a devfile already exists in the current directory"))
		})

		By("running odo init in a directory containing a .devfile.yaml", func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-registry.yaml"), filepath.Join(commonVar.Context, ".devfile.yaml"))
			defer os.Remove(filepath.Join(commonVar.Context, ".devfile.yaml"))
			err := helper.Cmd("odo", "init").ShouldFail().Err()
			Expect(err).To(ContainSubstring("a devfile already exists in the current directory"))
		})

		By("running odo init with wrong local file path given to --devfile-path", func() {
			err := helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", "/some/path/devfile.yaml").ShouldFail().Err()
			Expect(err).To(ContainSubstring("unable to download devfile"))
		})
		By("running odo init with wrong URL path given to --devfile-path", func() {
			err := helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", "https://github.com/path/to/devfile.yaml").ShouldFail().Err()
			Expect(err).To(ContainSubstring("unable to download devfile"))
		})
	})

	Context("running odo init with valid flags", func() {
		When("using --devfile flag", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "init", "--name", "aname", "--devfile", "go").ShouldPass().Out()
			})

			It("should download a devfile.yaml file", func() {
				files := helper.ListFilesInDir(commonVar.Context)
				Expect(files).To(Equal([]string{"devfile.yaml"}))
			})
		})
		When("using --devfile-path flag with a local devfile", func() {
			var newContext string
			BeforeEach(func() {
				newContext = helper.CreateNewContext()
				newDevfilePath := filepath.Join(newContext, "devfile.yaml")
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-registry.yaml"), newDevfilePath)
				helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", newDevfilePath).ShouldPass()
			})
			AfterEach(func() {
				helper.DeleteDir(newContext)
			})
			It("should copy the devfile.yaml file", func() {
				files := helper.ListFilesInDir(commonVar.Context)
				Expect(files).To(Equal([]string{"devfile.yaml"}))
			})
		})
		When("using --devfile-path flag with a URL", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", "https://raw.githubusercontent.com/odo-devfiles/registry/master/devfiles/nodejs/devfile.yaml").ShouldPass()
			})
			It("should copy the devfile.yaml file", func() {
				files := helper.ListFilesInDir(commonVar.Context)
				Expect(files).To(Equal([]string{"devfile.yaml"}))
			})
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
	Context("checking odo init final output message", func() {
		var newContext, devfilePath string
		BeforeEach(func() {
			newContext = helper.CreateNewContext()
			devfilePath = filepath.Join(newContext, "devfile.yaml")
		})
		AfterEach(func() {
			helper.DeleteDir(newContext)
		})
		When("the devfile used by `odo init` does not contain a deploy command", func() {
			var out string
			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), devfilePath)
				out = helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", devfilePath).ShouldPass().Out()
			})
			It("should only show information about `odo dev`, and not `odo deploy`", func() {
				Expect(out).To(ContainSubstring("odo dev"))
				Expect(out).ToNot(ContainSubstring("odo deploy"))
			})
		})
		When("the devfile used by `odo init` contains a deploy command", func() {
			var out string
			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-deploy.yaml"), devfilePath)
				out = helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", devfilePath).ShouldPass().Out()
			})
			It("should show information about both `odo dev`, and `odo deploy`", func() {
				Expect(out).To(ContainSubstring("odo dev"))
				Expect(out).To(ContainSubstring("odo deploy"))
			})
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

	When("source directory is empty", func() {
		BeforeEach(func() {
			Expect(helper.ListFilesInDir(commonVar.Context)).To(HaveLen(0))
		})

		It("name in devfile is personalized in non-interactive mode", func() {
			helper.Cmd("odo", "init", "--name", "aname", "--devfile-path",
				filepath.Join(helper.GetExamplePath(), "source", "devfiles", "nodejs",
					"devfile-with-starter-with-devfile.yaml")).ShouldPass()

			metadata := helper.GetMetadataFromDevfile(filepath.Join(commonVar.Context, "devfile.yaml"))
			Expect(metadata.Name).To(BeEquivalentTo("aname"))
			Expect(metadata.Language).To(BeEquivalentTo("nodejs"))
		})
	})
})
