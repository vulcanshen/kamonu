#!/bin/sh
# kamonu uninstaller for macOS / Linux.
# Usage: curl -fsSL https://raw.githubusercontent.com/vulcanshen/kamonu/main/uninstall.sh | sh

set -e

CANDIDATES="$HOME/.local/bin/kamonu /usr/local/bin/kamonu"

FOUND=""
for path in $CANDIDATES; do
  if [ -f "$path" ]; then
    FOUND="$path"
    break
  fi
done

if [ -z "$FOUND" ]; then
  echo "kamonu not found in expected locations."
  echo "Checked: $CANDIDATES"
  exit 1
fi

rm "$FOUND"
echo "removed $FOUND"

# ask offers to remove one directory: the answer is read from the terminal,
# not stdin — under `curl ... | sh` stdin is the script itself, so a plain
# `read` would consume script text (or hit EOF) instead of the keypress. No
# controlling terminal (cron, nohup) -> keep it, the safe default.
ask() {
  dir="$1"; what="$2"
  [ -d "$dir" ] || return 0
  printf "Remove %s in %s? [y/N]: " "$what" "$dir"
  if read -r answer < /dev/tty 2>/dev/null; then
    case "$answer" in
      y|Y|yes|YES) rm -rf "$dir"; echo "removed $dir" ;;
      *) echo "kept $dir" ;;
    esac
  else
    echo ""
    echo "kept $dir (no terminal to confirm on)"
  fi
}

# What the user drew: the marks (conf/) and the svg exported of them (svg/).
if [ -n "$KAMONU__DATA" ]; then
  DATA_DIR="$KAMONU__DATA"
else
  DATA_DIR="$HOME/.kamonu"
fi
ask "$DATA_DIR" "your marks and their exported svg"

echo ""
echo "kamonu uninstalled."
