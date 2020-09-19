#!/bin/sh

# fail if some commands fails
set -e
# show commands
set -x

python scripts/prow.py
exit 1
