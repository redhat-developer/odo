#!/bin/bash
set -e

# Install pandoc and asciidoctor

#############
# VARIABLES #
#############

tmp_openshift_docs_dir=$(mktemp -d -t odo-upstream-docs-XXXXXXXXXX)
tmp_public_docs_dir=$(mktemp -d -t odo-public-docs-XXXXXXXXXX)
odo_doc_directory="/cli_reference/openshift_developer_cli"
odo_public_doc_dir="/docs/public"
output_dir=$(mktemp -d -t odo-output-docs-XXXXXXXXXX)
openshift_docs="chosen-docs/openshift-docs"
upstream_docs="chosen-docs/upstream"
docs="docs"

openshift_docs_repo="github.com/openshift/openshift-docs"
odo_repo="github.com/openshift/odo"


#########################
# PLEASE READ           #
# DOC RELATED VARIABLES #
#########################
#
# There are some *product* related variables we *MUST* override when converting documentation as to not
# conflict with any upstream documentation.
# Ex: OpenShift Container Platform => OKD

attributes='product-title=OpenShift'


#############
# FUNCTIONS #
#############
#


shout() {
  echo -e "\n!!!!!!!!!!!!!!!!!!!!\n${1}\n!!!!!!!!!!!!!!!!!!!!\n"
}

convert_to_markdown() {
    noext="${1%.adoc}"
    asciidoctor -a "${attributes}" -b docbook $1 -o $noext.xml
    pandoc -f docbook -t gfm $noext.xml -o $noext.md
    file="$(basename -- $noext)"
    iconv -t utf-8 $noext.xml | pandoc -f docbook -t gfm | iconv -f utf-8 > $output_dir/$file.md
    echo "Converted $1 to markdown"
}

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



###########################
# UPSTREAM OPENSHIFT DOCS #
###########################

shout "Converting $openshift_docs_repo documentation"

# Clone the OpenShift documentation
git clone https://$openshift_docs_repo $tmp_openshift_docs_dir

# Directory
dir=$tmp_openshift_docs_dir$odo_doc_directory

# Convert all the documentation to markdown from the OpenShift upstream repository
for f in $dir/*.adoc; do
  convert_to_markdown $f
done

########################
# PUBLIC DOCUMENTATION # 
########################

shout "Converting $odo_repo documentation"

# Clone the master odo repo
#git clone https://$odo_repo $tmp_public_docs_dir
git clone --single-branch --branch add-new-docs https://github.com/cdrage/odo $tmp_public_docs_dir

# Directory
public_doc_dir=$tmp_public_docs_dir$odo_public_doc_dir

# Convert all the documentation to markdown from the OpenShift upstream repository
for f in $public_doc_dir/*.adoc; do
  convert_to_markdown $f
done

######################################
# COPY FILES OVER TO /docs IN JEKYLL #
######################################

shout "Merging files to /docs"

# Delete everything in /docs so it's clean for copying files over
rm -rf $docs/* || true

# Go through and "merge" the documentation if we have a template file available (this means we actually want to use the docs.openshift.com documentation
# everything else, we ignore.

echo "Merging openshift docs"
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

echo ""
echo "Merging upstream docs"
for f in $upstream_docs/*.md; do
  file="$(basename -- $f)"
  if [ -f $output_dir/$file ]; then
      echo "$file exists merging files"
      cp $upstream_docs/$file $docs/$file
      cat $output_dir/$file >> $docs/$file
  else 
      echo "$file does not exist, cancelling script, there is no upstream doc available from github.com/openshift/openshift-docs"
      exit 0
  fi
done
