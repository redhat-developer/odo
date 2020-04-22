#!/bin/bash
set -e

# Install pandoc and asciidoctor

#############
# VARIABLES #
#############

tmp_dir=$(mktemp -d -t odo-upstream-docs-XXXXXXXXXX)
odo_doc_directory="/cli_reference/openshift_developer_cli"
output_dir=$(mktemp -d -t odo-output-docs-XXXXXXXXXX)
openshift_docs="chosen-docs/openshift-docs"
upstream_docs="chosen-docs/upstream"
docs="docs"



#########################
# PLEASE READ           #
# DOC RELATED VARIABLES #
#########################
#
# There are some *product* related variables we *MUST* override when converting documentation as to not
# conflict with any upstream documentation.
# Ex: OpenShift Container Platform => OKD

attributes='product-title=OpenShift'

###########
# SYNCING # 
###########

# Here we go
# First of all, we have two sources of documentation. 
# 1. Upstream on: https://docs.openshift.com/container-platform/latest/cli_reference/openshift_developer_cli/
# 2. GitHub documentation
# 
# We first synchronize the documentation that we have for OpenShift from docs.openshift.com from https://github.com/openshift/openshift-docs
# Second, we have Devfile and Kubernetes documentation which is located at https://github.com/openshift/odo-docs which we keep SEPARATE from OpenShift documentation.

# Clone the OpenShift documentation
git clone https://github.com/openshift/openshift-docs $tmp_dir

# Directory
dir=$tmp_dir$odo_doc_directory

# Delete everything in /docs
rm -rf $docs/* || true

# Convert all the documentation to markdown from the OpenShift upstream repository
for f in $dir/*.adoc; do
    noext="${f%.adoc}"
    asciidoctor -a "${attributes}" -b docbook $f -o $noext.xml
    pandoc -f docbook -t gfm $noext.xml -o $noext.md
    file="$(basename -- $noext)"
    iconv -t utf-8 $noext.xml | pandoc -f docbook -t gfm | iconv -f utf-8 > $output_dir/$file.md
done

# Go through and "merge" the documentation if we have a template file available (this means we actually want to use the docs.openshift.com documentation
# everything else, we ignore.
for f in $openshift_docs/*.md; do
  file="$(basename -- $f)"
  if [ -f $output_dir/$file ]; then
      echo "$file exists merging files"
      cp $openshift_docs/$file $docs/$file
      cat $output_dir/$file >> $docs/$file
  else 
      echo "$file does not exist, cancelling script, there is no upstream doc available from github.com/openshift/openshift-docs"
      exit 0
  fi
done

# Copy upstream docs to folder
cp -r $upstream_docs/*.md $docs/
