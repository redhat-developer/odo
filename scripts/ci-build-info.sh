#!/usr/bin/env bash

if [[ -z $PR_NO ]]; then
	echo "Please set PR_NO"
	exit 1
fi

if [[ -z $GITHUB_TOKEN ]]; then
	echo "Please set GITHUB_TOKEN"
	exit 1
fi

REPO=openshift/odo

hist=`mktemp`

# Travis login with token here
echo "Logging in to travis with provided token..."
travis login --github-token $GITHUB_TOKEN --com

echo "Querying travis history..."
travis history -p $PR_NO --limit 4 -r $REPO --com | tr -s ' ' > $hist
echo "Iterating over travis history to find all builds for pr $PR_NO"
while IFS= read -r line
do
        build_no=`echo $line | cut -d ' ' -f1`
        build_no="${build_no//#}"
        build_status=`echo $line | cut -d ' ' -f2`
        build_status="${build_status//:}"
	if [[ ! -z $out ]];then
		out="$out "
	fi
	out="$out$build_no:$build_status"
done < "$hist"

echo $out
