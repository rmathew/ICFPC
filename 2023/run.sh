#!/usr/bin/env bash
if [[ -z "${1}" ]]; then
    echo "ERROR: Missing argument(s)."
    exit 1
fi
set -euo pipefail

MY_DIR="$(dirname "${0}")"
APP_DIR="concert"
APP="placer"
cd "${MY_DIR}/${APP_DIR}" && make "${APP}" && \
    ( ./"${APP}" "${@}" || echo "Failed with error-code: ${?}" )
