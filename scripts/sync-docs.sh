#!/usr/bin/env bash

# This document uses: https://gist.github.com/domenic/ec8b0fc8ab45f39403dd
# which effectively enables the synchronization of any documentation created on the GitHub repo
# to synchronize with the "gh-pages" branch and thus the website.

# Ensures that we run on Travis
if [ "$TRAVIS_BRANCH" != "master" ] || [ "$BUILD_DOCS" != "yes" ] || [ "$TRAVIS_SECURE_ENV_VARS" == "false" ] || [ "$TRAVIS_PULL_REQUEST" != "false" ] ; then
    echo "Must be: a merged pr on the master branch, BUILD_DOCS=yes, TRAVIS_SECURE_ENV_VARS=false"
    exit 0
fi

DOCS_REPO_NAME="odo"
DOCS_REPO_URL="git@github.com:redhat-developer/odo.git"
DOCS_REPO_HTTP_URL="http://github.com/redhat-developer/odo"
DOCS_KEY="scripts/deploy_key"
DOCS_USER="odo-bot"
DOCS_EMAIL="cdrage+odo@redhat.com"
DOCS_BRANCH="gh-pages"
DOCS_FOLDER="docs"

# decrypt the private key
openssl aes-256-cbc -K $encrypted_0e738444b7d0_key -iv $encrypted_0e738444b7d0_iv -in "$DOCS_KEY.enc" -out "$DOCS_KEY" -d
chmod 600 "$DOCS_KEY"
eval `ssh-agent -s`
ssh-add "$DOCS_KEY"

# clone the repo
git clone "$DOCS_REPO_URL" "$DOCS_REPO_NAME"

# change to that directory (to prevent accidental pushing to master, etc.)
cd "$DOCS_REPO_NAME"

# switch to gh-pages and grab the docs folder from master
git checkout gh-pages
git checkout master docs

# Remove README.md from docs folder as it isn't relevant
rm docs/README.md

# File reference is going to be built with "Slate"
# cp docs/file-reference.md slate/source/index.html.md

# clean-up the docs and convert to jekyll-friendly docs
cd docs
for filename in *.md; do
    if cat $filename | head -n 1 | grep "\-\-\-";
    then
    echo "$filename already contains Jekyll format"
    else
    # Remove ".md" from the name
    name=${filename::-3}
    echo "Adding Jekyll file format to $filename"
    jekyll="---
layout: default
permalink: /$name/
redirect_from: 
  - /docs/$name.md/
---
"
    echo -e "$jekyll\n$(cat $filename)" > $filename
    fi
done
cd ..

# This builds "slate" our file reference documentation.
#slate="---
#title: Odo File Reference
#
#language_tabs:
#  - yaml
#
#toc_footers:
#  - <a href='http://openshiftdo.org'>openshiftdo.org</a>
#  - <a href='https://github.com/redhat-developer/odo'>Odo (OpenShift Do) on GitHub</a>
#
#search: true
#---
#"

#echo -e "$slate\n$(cat slate/source/index.html.md)" >  slate/source/index.html.md
#cd slate
#docker run --rm -v $PWD:/usr/src/app/source -w /usr/src/app/source cdrage/slate bundle exec middleman build --clean
#cd ..
# Weird file permissions when building slate (since it's in a docker container)
#sudo chown -R $USER:$USER slate
# remove the old file-reference
#rm -rf file-reference 
#mv slate/build file-reference

# add relevant user information
git config user.name "$DOCS_USER"

# email assigned
git config user.email "$DOCS_EMAIL"
git add --all

# Check if anything changed, and if it's the case, push to origin/master.
if git commit -m 'Update docs' -m "Commit: $DOCS_REPO_HTTP_URL/commit/$TRAVIS_COMMIT" ; then
  git push
fi

# cd back to the original root folder
cd ..
