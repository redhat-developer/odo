#!/usr/bin/env bash

# This script uses github_changelog_generator (https://github.com/github-changelog-generator/github-changelog-generator/) to generate a changelog
# by using the kind labels that we use..
#
# Then outputs it to STDOUT. This helps generate a changelog when doing a release.

# Get the variables we're going to use..
if [ -z "$GITHUB_TOKEN" ]
then
  echo -e "GITHUB_TOKEN environment variable is blank..\nGet your GitHub token and then:\nexport GITHUB_TOKEN=yourtoken"
  exit 1
fi

if [ -z "$1" ]  || [ -z "$2" ]
then
  echo -e "Must provide first and next release numbers..\nex: ./changelog-script.sh v1.0.0 v1.0.1"
  exit 1
fi

MIRROR="https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/$2/"
INSTALLATION_GUIDE="https://odo.dev/docs/overview/installation"

echo -e "# Installation of $2
To install odo, follow our installation guide at [odo.dev]($INSTALLATION_GUIDE)
After each release, binaries are synced to [developers.redhat.com]($MIRROR)" > /tmp/base



RUN_GITHUB_CHANGELOG_GENERATOR="github_changelog_generator"

CONTAINER_ARGS="run -it --rm -v $(pwd):/usr/local/src/your-app docker.io/githubchangeloggenerator/github-changelog-generator"

if [ -x "$(command -v podman)" ]; then
  echo "Podman detected."
  RUN_GITHUB_CHANGELOG_GENERATOR="podman $CONTAINER_ARGS"
elif [ -x "$(command -v docker)" ]; then
  echo "Docker detected."
  RUN_GITHUB_CHANGELOG_GENERATOR="docker $CONTAINER_ARGS"
else
  echo "No container runtime detected."
fi


echo "Executing github_changelog_generator using '$RUN_GITHUB_CHANGELOG_GENERATOR'"
echo ""


$RUN_GITHUB_CHANGELOG_GENERATOR \
--max-issues 500 \
--user redhat-developer \
--project odo \
--no-issues \
-t $GITHUB_TOKEN \
--since-tag $1 \
--future-release $2 \
--base /tmp/base \
--output release-changelog.md \
--exclude-labels "lifecycle/rotten,duplicate,question,invalid,wontfix" \
--header-label "# Release of $2" \
--enhancement-label "**Features/Enhancements:**" \
--enhancement-labels "kind/feature" \
--bugs-label "**Bugs:**" \
--bug-labels "kind/bug" \
--add-sections '{"documentation":{"prefix":"**Documentation:**","labels":["kind/documentation"]}, "tests": {"prefix": "**Testing/CI:**", "labels": ["kind/tests"]}, "cleanup": {"prefix": "**Cleanup/Refactor:**", "labels": ["kind/code-refactoring"]}}'

echo "---------------------------------------"
cat release-changelog.md
echo "---------------------------------------"

echo ""
echo "The changelog is located at: ./release-changelog.md"
echo ""
