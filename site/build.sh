#!/bin/bash

# Copy over the files
./pre.sh

# This script builds the Odo website for viewing using a Docker container
export JEKYLL_VERSION=3.8
docker run --rm \
  --volume="$PWD:/srv/jekyll" \
  -it jekyll/jekyll:$JEKYLL_VERSION \
  jekyll build
