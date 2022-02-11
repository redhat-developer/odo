#!/bin/bash

# Why we are wrapping gofmt?
# - ignore files in vendor directory
# - gofmt doesn't exit with error code when there are errors

GO_FILES=$(find . -path ./vendor -prune -o -name '*.go' -print)

for file in $GO_FILES; do
	gofmtOutput=$(gofmt -l "$file")
	if [ "$gofmtOutput" ]; then
		errors+=("$gofmtOutput")
	fi
done


if [ ${#errors[@]} -eq 0 ]; then
	echo "gofmt OK"
else
	echo "gofmt ERROR - These files are not formated by gofmt:"
	for err in "${errors[@]}"; do
		echo "$err"
	done
	exit 1
fi
