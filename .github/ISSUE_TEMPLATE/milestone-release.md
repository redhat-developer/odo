---
name: Milestone Release
about: Create issue to release a new version of odo
title: "\U0001F389 [VERSION] Milestone Release \U0001F389 "
labels: area/release-eng, kind/task
assignees: ''

---

/kind task
/area release-eng

Issue to track work for publishing a new release of `odo`:

- Target release process start date: **[TBD]**
- Errata date (when we expect the binaries to be published): **[TBD]**
- Version: **[VERSION]**
- OCP Version: **[TBD]**
- Once the release is done, make sure the link to download `latest` points to this new version: https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/latest/

Notes:
- Make sure this issue is linked to a corresponding milestone, which should contain all necessary issues and/or PRs.
- A draft PR will be generated by the @github-actions bot when a new GitHub Release is created (it can be a pre-release). This PR will contain all the necessary changes (`build/VERSION`, blog post, release notes, ...).
  - Feel free to review and adjust the content
  - Merge this PR once the release binaries are available on the Content Gateway
- ~The docs team wants to start testing a process for optional docs approval on advisories. A pilot will start in early July. Check if this release can leverage this process to validate it.~