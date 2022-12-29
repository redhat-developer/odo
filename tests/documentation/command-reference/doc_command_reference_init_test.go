package docautomation

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("doc command reference odo init", Label(helper.LabelNoCluster), func() {
	var commonVar helper.CommonVar
	var commonPath = filepath.Join("command-reference", "docs-mdx", "init")
	var outputStringFormat = "```console\n$ odo %s\n%s```\n"

	BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
		Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeFalse())
	})

	AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})
	Context("Interactive Mode", func() {
		It("Empty directory", func() {
			args := []string{"odo", "init"}
			out, err := helper.RunInteractive(args, []string{"ODO_LOG_LEVEL=0"}, func(ctx helper.InteractiveContext) {
				helper.ExpectString(ctx, "Select language")
				helper.SendLine(ctx, "Java")

				helper.ExpectString(ctx, "Select project type")
				helper.SendLine(ctx, "")

				helper.ExpectString(ctx, "Select container for which you want to change configuration?")
				helper.SendLine(ctx, "")

				helper.ExpectString(ctx, "Which starter project do you want to use")
				helper.SendLine(ctx, "")

				helper.ExpectString(ctx, "Enter component name")
				helper.SendLine(ctx, "my-java-maven-app")

				helper.ExpectString(ctx, "Your new component 'my-java-maven-app' is ready in the current directory")
			})
			Expect(err).To(BeNil())
			got := fmt.Sprintf(outputStringFormat, " init", helper.StripSpinner(out))
			got = helper.ReplaceTimeInString(got, "[1s]")
			want := helper.GetMDXContent(filepath.Join(commonPath, "interactive_mode_empty_directory_output.mdx"))
			want = helper.ReplaceTimeInString(want, "[1s]")
			diff := cmp.Diff(want, got)
			Expect(diff).To(BeEmpty())
		})

		When("the directory is not empty", func() {
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			})

			It("Directory with sources", func() {
				args := []string{"odo", "init"}
				out, err := helper.RunInteractive(args, []string{"ODO_LOG_LEVEL=0"}, func(ctx helper.InteractiveContext) {
					helper.ExpectString(ctx, "Is this correct?")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "✓  Downloading devfile \"nodejs:2.1.1\" from registry \"DefaultDevfileRegistry\"")

					helper.ExpectString(ctx, "Select container for which you want to change configuration?")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "Enter component name")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "Changes will be directly reflected on the cluster.")
				})
				Expect(err).To(BeNil())
				got := helper.StripAnsi(out)
				got = fmt.Sprintf(outputStringFormat, "init", helper.StripSpinner(got))
				got = strings.Replace(got, "? Is this correct? (Y/n) ", "", 1)
				got = strings.Replace(got, "? Is this correct? Yes", "? Is this correct? Yes\n✓  Downloading devfile \"nodejs:2.1.1\" from registry \"DefaultDevfileRegistry\" [3s]", 1)
				got = strings.Replace(got, "\n? Select container for which you want to change configuration?  [Use arrows to move, type to filter]\nruntime\n> NONE - configuration is correct\n", "", 1)
				got = strings.Replace(got, "? Enter component name: (node-echo) ", "", 1)
				want := helper.GetMDXContent(filepath.Join(commonPath, "interactive_mode_directory_with_sources_output.mdx"))
				diff := cmp.Diff(want, got)
				Expect(diff).To(BeEmpty())
			})
		})
	})
	Context("Non Interactive Mode", func() {

		It("Fetch Devfile of a specific version", func() {
			args := []string{"init", "--devfile", "go", "--name", "my-go-app", "--devfile-version", "2.0.0"}
			out := helper.Cmd("odo", args...).ShouldPass().Out()
			got := fmt.Sprintf(outputStringFormat, strings.Join(args, " "), helper.StripSpinner(out))
			got = helper.ReplaceTimeInString(got, "[3s]")
			want := helper.GetMDXContent(filepath.Join(commonPath, "versioned_devfile_output.mdx"))
			diff := cmp.Diff(want, got)
			Expect(diff).To(BeEmpty())
		})

		It("Fetch Devfile of the latest version", func() {
			args := []string{"init", "--devfile", "go", "--name", "my-go-app", "--devfile-version", "latest"}
			out := helper.Cmd("odo", args...).ShouldPass().Out()
			got := fmt.Sprintf(outputStringFormat, strings.Join(args, " "), helper.StripSpinner(out))
			got = helper.ReplaceTimeInString(got, "[4s]")
			want := helper.GetMDXContent(filepath.Join(commonPath, "latest_versioned_devfile_output.mdx"))
			diff := cmp.Diff(want, got)
			Expect(diff).To(BeEmpty())
		})

		It("Fetch Devfile from a URL", func() {
			args := []string{"init", "--devfile-path", "https://registry.devfile.io/devfiles/nodejs-angular", "--name", "my-nodejs-app", "--starter", "nodejs-angular-starter"}
			out := helper.Cmd("odo", args...).ShouldPass().Out()
			got := fmt.Sprintf(outputStringFormat, strings.Join(args, " "), helper.StripSpinner(out))
			got = helper.ReplaceTimeInString(got, "[4s]")
			want := helper.GetMDXContent(filepath.Join(commonPath, "devfile_from_url_output.mdx"))
			want = helper.ReplaceTimeInString(want, "[4s]")
			diff := cmp.Diff(want, got)
			Expect(diff).To(BeEmpty())
		})

		Context("fetching devfile from a registry", func() {
			When("setting up the registry", func() {
				const (
					defaultReg    = "DefaultDevfileRegistry"
					defaultRegURL = "https://registry.devfile.io"
					stagingReg    = "StagingRegistry"
					stagingRegURL = "https://registry.stage.devfile.io"
				)
				BeforeEach(func() {
					helper.Cmd("odo", "preference", "remove", "registry", defaultReg, "-f").ShouldPass()
					helper.Cmd("odo", "preference", "add", "registry", defaultReg, defaultRegURL).ShouldPass()

					helper.Cmd("odo", "preference", "add", "registry", stagingReg, stagingRegURL).ShouldPass()
				})

				AfterEach(func() {
					helper.Cmd("odo", "preference", "remove", "registry", stagingReg, "-f").ShouldPass()
					helper.SetDefaultDevfileRegistryAsStaging()
				})

				removePreferenceKeys := func(docString string) (returnString string) {
					command := strings.SplitAfter(docString, "odo preference view\n")
					devfileRegistries := strings.Split(command[1], "\n\n")
					returnString = command[0] + "[...]\n\n" + devfileRegistries[1] + "\n\n" + devfileRegistries[2]
					return
				}
				checkRegistriesOutput := func() {
					args := []string{"preference", "view"}
					out := helper.Cmd("odo", args...).ShouldPass().Out()
					got := helper.StripAnsi(out)
					got = fmt.Sprintf(outputStringFormat, strings.Join(args, " "), helper.StripSpinner(got))
					got = removePreferenceKeys(got)
					want := helper.GetMDXContent(filepath.Join(commonPath, "registry_output.mdx"))
					diff := cmp.Diff(want, got)
					Expect(diff).To(BeEmpty())
				}

				It("Fetch Devfile from a specific registry of the list", func() {
					By("checking for required registries", func() {
						checkRegistriesOutput()
					})

					By("checking for the init output", func() {
						args := []string{"init", "--name", "my-spring-app", "--devfile", "java-springboot", "--devfile-registry", "DefaultDevfileRegistry", "--starter", "springbootproject"}
						out := helper.Cmd("odo", args...).ShouldPass().Out()
						got := fmt.Sprintf(outputStringFormat, strings.Join(args, " "), helper.StripSpinner(out))
						got = helper.ReplaceTimeInString(got, "[4s]")
						want := helper.GetMDXContent(filepath.Join(commonPath, "devfile_from_specific_registry_output.mdx"))
						want = helper.ReplaceTimeInString(want, "[4s]")
						diff := cmp.Diff(want, got)
						Expect(diff).To(BeEmpty())
					})
				})
				It("Fetch Devfile from any registry of the list", func() {
					By("checking for required registries", func() {
						checkRegistriesOutput()
					})

					By("checking for the registry list output", func() {
						args := []string{"registry", "--devfile", "nodejs-react"}
						out := helper.Cmd("odo", args...).ShouldPass().Out()
						got := helper.StripAnsi(out)
						got = fmt.Sprintf(outputStringFormat, strings.Join(args, " "), helper.StripSpinner(got))
						want := helper.GetMDXContent(filepath.Join(commonPath, "registry_list_output.mdx"))
						diff := cmp.Diff(want, got)
						Expect(diff).To(BeEmpty())
					})

					By("checking for the init output", func() {
						args := []string{"init", "--devfile", "nodejs-react", "--name", "my-nr-app"}
						out := helper.Cmd("odo", args...).ShouldPass().Out()
						got := fmt.Sprintf(outputStringFormat, strings.Join(args, " "), helper.StripSpinner(out))
						got = helper.ReplaceTimeInString(got, "[3s]")
						want := helper.GetMDXContent(filepath.Join(commonPath, "devfile_from_any_registry_output.mdx"))
						diff := cmp.Diff(want, got)
						Expect(diff).To(BeEmpty())
					})
				})

			})
		})

		It("Fetch Devfile from a URL", func() {
			args := []string{"init", "--devfile-path", "https://registry.devfile.io/devfiles/nodejs-angular", "--name", "my-nodejs-app", "--starter", "nodejs-angular-starter"}
			out := helper.Cmd("odo", args...).ShouldPass().Out()
			got := fmt.Sprintf(outputStringFormat, strings.Join(args, " "), helper.StripSpinner(out))
			got = helper.ReplaceTimeInString(got, "[1s]")
			want := helper.GetMDXContent(filepath.Join(commonPath, "devfile_from_url_output.mdx"))
			want = helper.ReplaceTimeInString(want, "[1s]")
			diff := cmp.Diff(want, got)
			Expect(diff).To(BeEmpty())
		})
	})

})
