#!/bin/bash

echo "Checking for nested vendor dirs"

# count all vendor directories inside Odo vendor
NO_NESTED_VENDORS=$(find vendor/ -type d | sed 's/^[^/]*.//g' | grep -E "vendor$" | grep -v _vendor | wc -l)

if [ $NO_NESTED_VENDORS -ne 0 ]; then
    echo "ERROR"
    echo "  There are $NO_NESTED_VENDORS nested vendors in Odo vendor directory"
    echo "  Please run 'glide update --strip-vendor'"
    exit 1
else
    echo "OK"
    exit 0
fi
