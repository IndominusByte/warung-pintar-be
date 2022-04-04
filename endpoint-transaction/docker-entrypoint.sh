#!/bin/sh
set -e

if [ "$BACKEND_STAGE" = 'production' ]; then
  make run
else
  make watch
fi
