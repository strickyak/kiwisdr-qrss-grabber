#!/bin/bash

set -ex

ID="$1"; shift
HOST="$1"; shift
BASE_FREQ="$1"; shift
MINUTES="$1"; shift

SECONDS=$(expr 60 '*' $MINUTES)

python2 vendor/jks-prv/kiwiclient/kiwirecorder.py --server-host "$HOST" --user "$ID" --station "$ID" --log info --dir /tmp/ --time-limit "$SECONDS"   -f "$BASE_FREQ" -m usb --no_compression >&2
