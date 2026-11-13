#!/bin/sh
# Records asciiverse/<id>.cast, the catalog previews, at 100x30
# (hub/RECORDING.md): each game started with --demo, the computer playing
# every seat, dealt two seconds in, for PLAY seconds (default 50), then
# stopped. Hands are shuffled from the clock, so every take differs; record
# again and keep the best.
#
# Uses hub/tools/castrec.py (a sibling checkout at ../../hub), which runs
# the game as user "player", so the human seat shows "Player". The session
# log goes to a throwaway state folder.
#
#   sh asciiverse/record-preview.sh [id ...]   (default: all five)
set -eu
root=$(cd "$(dirname "$0")/.." && pwd)
castrec=${CASTREC:-$root/../../hub/tools/castrec.py}
play=${PLAY:-50}
bin=$(mktemp -d)
state=$(mktemp -d)
trap 'rm -rf "$bin" "$state"' EXIT
ids=${*:-crazy-eights holdem colour-clash hearts gin-rummy}
for id in $ids; do
  (cd "$root" && CGO_ENABLED=0 go build -trimpath -o "$bin/$id" "./cmd/$id")
  XDG_STATE_HOME=$state python3 "$castrec" "$root/asciiverse/$id.cast" "$bin" "w$play" -- "$bin/$id" --demo
done
