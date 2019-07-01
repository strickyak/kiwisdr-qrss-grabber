#!/bin/bash

set -ex

ID="$1"; shift
HOST="$1"; shift
BASE_FREQ="$1"; shift
MINUTES="$1"; shift

# Be in same directory as the script.
cd "$(dirname "$0")"

while true
do
	TIMESTAMP=$(date -u +%Y-%m-%d-%H%M%Sz)

	bash ./recorder.sh "__QRSS__$ID" "$HOST" "$BASE_FREQ" "$MINUTES"

	set /tmp/*__QRSS__$ID*.wav

	go run render.go -imprint "$ID  $TIMESTAMP  ($MINUTES min)" < "$1" > "/tmp/__QRSS__$ID.png"

	bash ./publish.sh "$ID" "$TIMESTAMP" "$MINUTES" "/tmp/__QRSS__$ID.png"

	rm -f /tmp/*__QRSS__$ID*.wav "/tmp/__QRSS__$ID.png"
done
