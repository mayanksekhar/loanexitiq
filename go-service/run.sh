#!/usr/bin/env bash
# Rebuild and restart the dev server on :8080, killing any existing listener.
set -e
cd "$(dirname "$0")"

PID=$(lsof -t -i :8080 2>/dev/null || true)
if [ -n "$PID" ]; then
  echo "Killing existing listener (PID $PID)"
  kill "$PID" 2>/dev/null || true
  sleep 1
fi

templ generate
go build -o /tmp/loanexitiq-dev .
/tmp/loanexitiq-dev > /tmp/loanexitiq.log 2>&1 &
sleep 1

CODE=$(curl -s http://localhost:8080/ -o /dev/null -w "%{http_code}")
if [ "$CODE" = "200" ]; then
  echo "Server up on http://localhost:8080 (logs: /tmp/loanexitiq.log)"
else
  echo "Server failed to start (HTTP $CODE). Last log lines:"
  tail -20 /tmp/loanexitiq.log
  exit 1
fi
