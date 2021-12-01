#!/usr/bin/env bash

# This document uses: https://gist.github.com/domenic/ec8b0fc8ab45f39403dd
# which effectively enables the synchronization of any documentation created on the GitHub repo
# to synchronize with the "gh-pages" branch and thus the website.
#
# In-case the above Gist is out of date, here are the instructions:
# 
# 1. Generate a NEW SSH key that will be used to commit to the branch. This should have administrative
# privileges to modify your repo. https://help.github.com/articles/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent/
#
# 2. Encrypt the key using travis:
# $ travis encrypt-file deploy_key
# encrypting deploy_key for domenic/travis-encrypt-file-example
# storing result as deploy_key.enc
# storing secure env variables for decryption
# 
# Please add the following to your build script (before_install stage in your .travis.yml, for instance):
# 
#     openssl aes-256-cbc -K $encrypted_0a6446eb3ae3_key -iv $encrypted_0a6446eb3ae3_key -in super_secret.txt.enc -out super_secret.txt -d
# 
# Pro Tip: You can add it automatically by running with --add.
# 
# Make sure to add deploy_key.enc to the git repository.
# Make sure not to add deploy_key to the git repository.
# Commit all changes to your .travis.yml.
# 
# 3. Make note of the value.. and add it to the ENCRPYTION_LABEL environment variable below.


# Ensures that we run on Travis
if [ "$TRAVIS_BRANCH" != "master" ] || [ "$BUILD_DOCS" != "yes" ] || [ "$TRAVIS_SECURE_ENV_VARS" == "false" ] || [ "$TRAVIS_PULL_REQUEST" != "false" ] ; then
    echo "Must be: a merged pr on the master branch, BUILD_DOCS=yes, TRAVIS_SECURE_ENV_VARS=false"
    exit 0
fi


# Change the below to your credentials
DOCS_REPO_NAME="odo"
DOCS_REPO_URL="git@github.com:redhat-developer/odo.git"
DOCS_REPO_HTTP_URL="http://github.com/redhat-developer/odo"
DOCS_USER="odo-bot"
DOCS_EMAIL="cdrage+odo@redhat.com"

# Your encrypted key values as from Steps 2&3 of the above tutorial
ENCRYPTION_LABEL="0e738444b7d0"

# Things that don't "really" need to be changed
DOCS_KEY="scripts/deploy_key"
DOCS_BRANCH="gh-pages"
DOCS_FOLDER="docs"

# decrypt the private key
ENCRYPTED_KEY_VAR="encrypted_${ENCRYPTION_LABEL}_key"
ENCRYPTED_IV_VAR="encrypted_${ENCRYPTION_LABEL}_iv"
ENCRYPTED_KEY=${!ENCRYPTED_KEY_VAR}
ENCRYPTED_IV=${!ENCRYPTED_IV_VAR}
openssl aes-256-cbc -K $ENCRYPTED_KEY -iv $ENCRYPTED_IV -in "$DOCS_KEY.enc" -out "$DOCS_KEY" -d
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

# =========================
# START Modify the documentation
# =========================

# Remove README.md from docs folder as it isn't relevant
rm docs/README.md

# Copy over the original README.md in the root directory
# to use as the index page for "documentation" on the site
cp README.md docs/readme.md

# TODO: Add Slate in the future. Keep this here for reference.
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

# TODO: Add Slate in the future. Keep this here for reference.
# This builds "slate" our file reference documentation.
#slate="---
#title: Odo File Reference
#
#language_tabs:
#  - yaml
#
#toc_footers:
#  - <a href='http://openshiftdo.org'>openshiftdo.org</a>
#  - <a href='https://github.com/redhat-developer/odo'>odo on GitHub</a>
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


# =========================
# END Modify the documentation
# =========================


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
