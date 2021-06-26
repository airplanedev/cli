#!/bin/bash

# Checks if `go mod tidy` produces any changes and prints the diff.
# 
# Based on https://github.com/google/cadvisor/blob/master/build/check_gotidy.sh
# Could eventually be replaced by `go mod tidy --check`:
# https://github.com/golang/go/issues/27005
function check_gotidy() {
    TMP_GOMOD=$(mktemp)
    TMP_GOSUM=$(mktemp)

    # Snapshot current files
    cp go.mod "${TMP_GOMOD}"
    cp go.sum "${TMP_GOSUM}"

    go mod tidy

    # Diff against snapshot
    DIFF_MOD=$(diff -u "${TMP_GOMOD}" go.mod)
    DIFF_SUM=$(diff -u "${TMP_GOSUM}" go.sum)

    # Restore snapshot
    cp "${TMP_GOMOD}" go.mod
    cp "${TMP_GOSUM}" go.sum

    if [[ -n "${DIFF_MOD}" || -n "${DIFF_SUM}" ]]; then
        echo "go tidy changes are needed; please run `go mod tidy`"
        echo "go.mod diff:"
        echo "${DIFF_MOD}"
        echo "go.sum diff:"
        echo "${DIFF_SUM}"
        exit 1
    fi
}

check_gotidy
