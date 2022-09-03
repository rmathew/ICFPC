#!/bin/bash
if [[ -z "${1}" ]]; then
    echo "ERROR: Missing argument(s)."
    exit 1
fi
set -euo pipefail

DIR="robovinci"
APP="lenny"
cd "${DIR}" && make "${APP}" && \
    ( ./"${APP}" "${@}" || echo "Failed with error-code: ${?}" )
