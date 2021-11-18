---
title: Pull Request Review
sidebar_position: 5
---

Each PR requires *one* code review (lgtm label) and *one* approve (approved label).
After a code reviewer is satisfied with the changes they will add `/lgtm` (looks good to me) as a comment to the PR that applies the *lgtm* label. Optionally, approving the PR via GitHub review also adds *lgtm* label to the PR.

The approver checks the functional part of the PR and if satisfied, adds `/approve` as a comment to the PR that applies the *approve* label.

Once the PR has *lgtm* and *approve* labels and the required tests pass, the bot automatically merges the PR.

Following are some tips you can take into account while reviewing a pull request:
- Ensure that necessary tests have been added.
- Ensure that the feature or fix works locally.
- Check if the code is understandable, and has comments been added to it. 
- Check if the PR passes all the pre-submit tests, and all the requested changes are resolved.
- As a code reviewer, if you apply the `/lgtm` label before it meets all the necessary criteria, put it on hold with the `/hold` label immediately. You can use `/lgtm cancel` to cancel your `/lgtm` and use `/hold cancel` once you are ready to approve it. This especially applies to draft PRs.
- As an approver, you can use `/approve` and `/approve cancel` to approve or hold their approval respectively.
- Avoid merging the PR manually unless it is an emergency, and you have the required permissions. Prow’s tide component automatically merges the PR once all the conditions are met.

### Prow
odo uses the [Prow](https://github.com/kubernetes/test-infra/tree/master/prow) infrastructure for CI testing. Use the [command-help](https://prow.ci.openshift.org/command-help) to see the list of possible bot commands.

Prow uses [OWNERS](https://github.com/kubernetes/community/blob/master/contributors/guide/owners.md) files to determine who can approve and lgtm a PR.
It also ensures that post-submit tests (tests that run before merge) validate the PR. 

Prow has two levels of OWNERS:
   1. **Approvers** look for holistic acceptance criteria, including dependencies with other features, forward and backward compatibility, API and flag definitions, etc. In essence, the high levels of design.
   2. **Reviewers** look for general code quality, correctness, sane software engineering, style, etc. In essence, the quality of the actual code itself.

