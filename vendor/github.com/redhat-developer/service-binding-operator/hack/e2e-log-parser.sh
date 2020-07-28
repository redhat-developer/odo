#!/bin/bash
#
# Based on the input extracted from operator-sdk running end-to-end tests, parse and print out
# regular log lines.
#

set -e

grep 'msg="Local operator stderr' $1 \
    |sed 's/^time.*err: //g;s/\\n\"/\\n/g' \
    |eval 'in=$(cat); echo -en "$in"'
