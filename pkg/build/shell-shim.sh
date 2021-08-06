#!/bin/bash

# Params are based in as param_slug_1=value1, param_slug_2=value2
# Export as environment varaibles, PARAM_SLUG_1=value1, PARAM_SLUG_2=value2
for param in "$@"; do
    param_slug="$(cut -d '=' -f 1 <<< "${param}")"
    param_value="$(cut -d '=' -f 2 <<< "${param}")"
    # Convert to uppercase
    var_name="$(echo "PARAM_${param_slug}" | tr '[:lower:]' '[:upper:]')"
    # Export env var
    export "${var_name}"="${param_value}"
done

exec "{{ .Entrypoint }}"
