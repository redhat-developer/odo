#!/bin/bash

# Copy over the docs files from the root directory
cp -r ../docs .

# Add the jekyll format to each doc
cd docs
for filename in *.adoc; do
    if cat $filename | head -n 1 | grep "\-\-\-";
    then
    echo "$filename already contains Jekyll format"
    else
    # Remove ".md" from the name
    name=${filename::-5}
    echo "Adding Jekyll file format to $filename"
    jekyll="---
layout: default
permalink: /$name/
redirect_from: 
  - /docs/$name.adoc/
---
"
    echo -e "$jekyll\n$(cat $filename)" > $filename
    fi
done
cd ..
