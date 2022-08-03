#!/bin/bash

set -e

for i in FIt FDescribe FContext FWhen FEntry FDescribeTable
do
  if grep -nr "$i(" tests/; then
    echo -e "$i: Not OK. $i exists somewhere in the testing code. Please remove it.\n"
    exitCode=1
  else
    echo -e "$i: OK\n"
  fi
done

exit $exitCode
