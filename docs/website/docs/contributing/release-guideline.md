---
title: Release Guideline
sidebar_position: 8
---

### Releasing a new version

Making artifacts for a new release is automated within RedHat's internal CI servers.

1. Create a PR to update the version. Using the helper script [scripts/bump-version.sh](https://github.com/redhat-developer/odo/blob/main/scripts/bump-version.sh) update the version in the following files:
  - [pkg/version/version.go](https://github.com/redhat-developer/odo/blob/main/pkg/version/version.go)
  - [Dockerfile.rhel](https://github.com/redhat-developer/odo/blob/main/Dockerfile.rhel)
  - [scripts/rpm-prepare.sh](https://github.com/redhat-developer/odo/blob/main/scripts/rpm-prepare.sh)
2. Merge the above PR.
3. Once the PR is merged create and push the new git tag for the version.
4. Create a new release using the GitHub site (this must be a proper release and NOT a draft).
5. Update the release description (changelog) on GitHub.
   Run the script below to get the baseline template of release changelog and then copy and paste the content from the [Changelog.md](https://github.com/redhat-developer/odo/blob/main/Changelog.md) into appropriately marked location in generated `/tmp/changelog.md`.
  ```shell
  export GITHUB_TOKEN=yoursupersecretgithubtoken
  ./scripts/changelog-script.sh ${PREVIOUS_VERSION} ${NEW_VERSION}
  ```
6. Update the Homebrew package:
  1. Check commit id for the released tag `git show-ref v0.0.1`
  2. Create a PR to update `:tag` and `:revision` in the [odo.rb](https://github.com/kadel/homebrew-odo/blob/master/Formula/odo.rb) file in [kadel/homebrew-odo](https://github.com/kadel/homebrew-odo).
  3. Create a PR and update the file `build/VERSION` with the  latest version number.
  4. Create a blog post! Follow the [template.md](https://github.com/redhat-developer/odo/blob/main/RELEASE_TEMPLATE.md) file and push it to the website; learn how to push to the website [here](docs.md).
  5. After the blog post, ideally the CHANGELOG in the release should be the same as the blog post. This is an example of a good release changelog: <https://github.com/redhat-developer/odo/releases/tag/v2.0.0>
  6. Add the built site (including the blog post) to the release with `site.tar.gz` using the [bundling the site for releases](https://github.com/redhat-developer/odo/tree/gh-pages#bundling-the-site-for-releases) guide.

