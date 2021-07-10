#!/bin/sh
set -euo pipefail
cd poses && make poses && ( ./poses "$@" || echo "Failed with error-code: $?" )
