#!/bin/sh
set -e

# Check if TS_AUTHKEY is provided
if [ -n "$TS_AUTHKEY" ]; then
  # If TS_AUTHKEY is provided, enable Tailscale mode
  echo "Starting in Tailscale mode with hostname: ${HOSTNAME:-chatroom}"
  exec ./chat-server --tailscale --hostname "${HOSTNAME:-chatroom}" --port "$PORT" --room-name "$ROOM_NAME" --max-users "$MAX_USERS" "$@"
else
  # Otherwise, start in regular TCP mode
  echo "Starting in regular TCP mode on port: $PORT"
  exec ./chat-server --port "$PORT" --room-name "$ROOM_NAME" --max-users "$MAX_USERS" "$@"
fi