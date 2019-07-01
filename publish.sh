#!/bin/bash

set -ex

ID="$1"; shift
TIMESTAMP="$1"; shift
MINUTES="$1"; shift
PNG="$1"; shift

DATE=$(date -u +%Y-%m-%d)
mkdir -p "$HOME/pub.qrss/$ID/$DATE"
cp -av "$PNG" "$HOME/pub.qrss/$ID/$DATE/$ID-$TIMESTAMP.png"
