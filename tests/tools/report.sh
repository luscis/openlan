#!/bin/bash

REPORT_DIR=""
REPORT_FILE=""
REPORT_HTML=""
REPORT_TITLE=""
REPORT_ROWS_HTML=""
REPORT_RUN_STAMP=""
REPORT_RUN_DIR=""
REPORT_CASE_DIR=""

report_init_env() {
  local base_dir="$1"
  REPORT_DIR="$base_dir/reports"
}

init_report() {
  local stamp
  stamp=$(date +%y%m%d-%H%M%S)
  REPORT_RUN_STAMP="$stamp"
  REPORT_RUN_DIR="$REPORT_DIR/$REPORT_RUN_STAMP"
  REPORT_CASE_DIR="$REPORT_RUN_DIR/cases"
  mkdir -p "$REPORT_DIR"
  mkdir -p "$REPORT_RUN_DIR"
  mkdir -p "$REPORT_CASE_DIR"
  REPORT_FILE="$REPORT_RUN_DIR/run.log"
  REPORT_HTML="$REPORT_RUN_DIR/run.html"
  REPORT_ROWS_HTML=""
  : > "$REPORT_FILE"
}

report_case_log_file() {
  local index="$1"
  local name="$2"
  local safe
  safe=$(echo "$name" | tr -c 'A-Za-z0-9._-' '_')
  echo "$REPORT_CASE_DIR/$(printf "%02d" "$index")-${safe}.log"
}

report_line() {
  local line="$1"
  echo "$line" >> "$REPORT_FILE"
}

write_report_header() {
  local title="$1"
  REPORT_TITLE="$title"
  report_line "OpenLAN Test Report"
  report_line "Start: $(now_text)"
  report_line "Title: $title"
  report_line ""
  report_line "Details:"
}

report_case_html() {
  local status="$1"
  local name="$2"
  local cost="$3"
  local log_path="$4"
  local ts
  local cls
  local name_html
  local rel_log
  ts=$(now_text)
  cls="fail"
  if [[ "$status" == "PASS" ]]; then
    cls="pass"
  fi
  name_html="$name"
  if [[ -n "$log_path" ]]; then
    rel_log="cases/$(basename "$log_path")"
    name_html="<a href=\"${rel_log}\">${name}</a>"
  fi
  REPORT_ROWS_HTML="${REPORT_ROWS_HTML}
<tr><td>${ts}</td><td class=\"${cls}\">${status}</td><td>${name_html}</td><td>${cost}</td></tr>"
}

write_report_html() {
  local pass_count="$1"
  local fail_count="$2"
  local total="$3"
  cat > "$REPORT_HTML" <<EOF
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>OpenLAN Test Report</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; margin: 24px; color: #111827; background: #f8fafc; }
    .card { background: #fff; border: 1px solid #e5e7eb; border-radius: 12px; padding: 16px; margin-bottom: 16px; }
    h1 { margin: 0 0 8px; font-size: 22px; }
    .meta { color: #6b7280; font-size: 14px; }
    .sum { display: flex; gap: 16px; margin-top: 8px; }
    .sum span { font-weight: 600; }
    table { width: 100%; border-collapse: collapse; background: #fff; border: 1px solid #e5e7eb; border-radius: 12px; overflow: hidden; }
    th, td { text-align: left; padding: 10px 12px; border-bottom: 1px solid #e5e7eb; font-size: 14px; }
    th { background: #f3f4f6; }
    .pass { color: #065f46; font-weight: 700; }
    .fail { color: #991b1b; font-weight: 700; }
  </style>
</head>
<body>
  <div class="card">
    <h1>OpenLAN Test Report</h1>
    <div class="meta">Title: ${REPORT_TITLE}</div>
    <div class="meta">Generated: $(now_text)</div>
    <div class="sum">
      <div>Passed: <span>${pass_count}</span></div>
      <div>Failed: <span>${fail_count}</span></div>
      <div>Total: <span>${total}</span></div>
    </div>
  </div>
  <table>
    <thead>
      <tr><th>Time</th><th>Status</th><th>Scenario</th><th>Cost</th></tr>
    </thead>
    <tbody>
${REPORT_ROWS_HTML}
    </tbody>
  </table>
</body>
</html>
EOF
}
