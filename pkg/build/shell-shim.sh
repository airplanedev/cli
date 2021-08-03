#!/bin/bash

echo "DEBUG: RUNNING SHIM"
# TODO: set up env vars - for each argument, split by first =, set AP_FIRST_PART=SECOND_PART

exec "{{ .Entrypoint }}"
