# Notes on documentation

This is the GitHub branch for [odo.dev](https://odo.dev).

## Files and file directories

- `/chosen-docs`: List of docs which have been "chosen" to be included on the site. These are all **TEMPLATE** files.
  - `/chosen-docs/openshift-docs` - Documentation from [docs.openshift.com](https://docs.openshift.com)
  - `/chosen-docs/upstream` - Documentation from the `master` branch of [github.com/openshift/odo](https://github.com/openshift/odo)
- `/docs`: Where all the converted documentation goes
- `/default.md`: "Default" template when using 
- `/index.md`: Where we list all front-page documentation


## How to update documentation.

1. Check the `/chosen-docs` directory for which file you want to update.
2. Find out where the file is located.
  - If the file is located in the `/upstream` folder, update it here: [github.com/openshift/odo/tree/master/docs/public](https://github.com/openshift/odo/tree/master/docs/public)
  - If the file is located in the `/openshift-docs` folder, update it here: [github.com/openshift/openshift-docs/tree/master/cli_reference/openshift_developer_cli](https://github.com/openshift/openshift-docs/tree/master/cli_reference/openshift_developer_cli)
3. Once you have updated the files, run the `./script/sync.sh` file
4. Commit and push the `/docs` folder

## How to add new documentation

1. Make sure that the document exists upstream or on the `/openshift-docs` GitHub repository.
2. Use the `default.md` template.
3. Create a markdown file with the *exact* same name in the `/chosen-docs` folder.
4. Update the `index.md` file with the new file
