#!/bin/bash

function check_csvs () {
    local csv_name="$1"

    for i in  {1..120} ; do
        if ( kubectl get csvs -n default | grep ${csv_name} 2>&1 > /dev/null ) ; then
            return 0
        fi

        sleep 10
    done

    return 1
}

CSV_NAME="service-binding-operator"

echo "# Searching for '${CSV_NAME}'..."

if ! check_csvs ${CSV_NAME} ; then
    echo "csv doesn't exist: ${CSV_NAME}"
    exit 1
fi

echo "csv is found: ${CSV_NAME}"
