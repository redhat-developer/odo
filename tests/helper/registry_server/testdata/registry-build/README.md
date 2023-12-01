This folder is used to generate and serve artifacts by a Devfile Registry started only for testing.

To update, simply copy the relevant files from an existing folder from an existing registry stacks folder (see https://github.com/devfile/registry/tree/main/stacks)
to this folder.
Then, and anytime a change is made to the source `registry/stacks` folder, you need to run 'make generate-test-registry-build'
to regenerate the `registry-build` folder with the artifacts that will be served for the tests.