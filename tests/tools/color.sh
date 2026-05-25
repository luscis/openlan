#!/bin/bash

if [[ -t 1 ]]; then
  C_RESET=$'\033[0m'
  C_RED=$'\033[31m'
  C_GREEN=$'\033[32m'
  C_CYAN=$'\033[36m'
else
  C_RESET=""
  C_RED=""
  C_GREEN=""
  C_CYAN=""
fi

now_text() {
  date "+%y-%m-%d %H:%M:%S"
}

now_ms() {
  if [[ -n "${EPOCHREALTIME:-}" ]]; then
    awk -v t="$EPOCHREALTIME" 'BEGIN { printf "%d\n", t * 1000 }'
    return
  fi
  echo $(( $(date +%s) * 1000 ))
}

cost_s() {
  local start_ms=$1
  local end_ms
  local diff_ms
  end_ms=$(now_ms)
  diff_ms=$((end_ms - start_ms))
  printf "%d.%03ds" $((diff_ms / 1000)) $((diff_ms % 1000))
}
