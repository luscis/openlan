#!/bin/bash

REPORT_DIR=""
REPORT_HTML=""
REPORT_MD=""
REPORT_TITLE=""
REPORT_ROWS_HTML=""
REPORT_ROWS_MD=""
REPORT_RUN_STAMP=""
REPORT_RUN_DIR=""
REPORT_CASE_DIR=""
REPORT_TAR=""

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
  REPORT_HTML="$REPORT_RUN_DIR/run.html"
  REPORT_MD="$REPORT_RUN_DIR/run.md"
  REPORT_TAR="$REPORT_DIR/$REPORT_RUN_STAMP.tar"
  REPORT_ROWS_HTML=""
  REPORT_ROWS_MD=""
}

pack_report_tar() {
  tar -cf "$REPORT_TAR" -C "$REPORT_DIR" "$REPORT_RUN_STAMP" && \
    rm -rf "$REPORT_RUN_DIR"
}

report_case_log_file() {
  local index="$1"
  local name="$2"
  local safe
  safe=$(echo "$name" | tr -c 'A-Za-z0-9._-' '_')
  echo "$REPORT_CASE_DIR/$(printf "%02d" "$index")-${safe}.log"
}


write_report_header() {
  local title="$1"
  REPORT_TITLE="$title"
}

report_case_html() {
  local status="$1"
  local name="$2"
  local cost="$3"
  local topo="$4"
  local log_path="$5"
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
<tr><td>${ts}</td><td class=\"${cls}\">${status}</td><td>${name_html}</td><td>${topo}</td><td>${cost}</td></tr>"
}

report_case_md() {
  local status="$1"
  local name="$2"
  local cost="$3"
  local topo="$4"
  local log_path="$5"
  local ts
  local name_md
  local rel_log
  local topo_md
  ts=$(now_text)
  name_md="\`$name\`"
  if [[ -n "$log_path" ]]; then
    rel_log="cases/$(basename "$log_path")"
    name_md="[\`$name\`](${rel_log})"
  fi
  topo_md=${topo//|/\\|}
  if [[ -n "$REPORT_ROWS_MD" ]]; then
    REPORT_ROWS_MD="${REPORT_ROWS_MD}
| ${ts} | ${status} | ${name_md} | ${topo_md} | ${cost} |"
  else
    REPORT_ROWS_MD="| ${ts} | ${status} | ${name_md} | ${topo_md} | ${cost} |"
  fi
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
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      margin: 24px;
      color: #e5e7eb;
      background: #0b1220;
    }
    .card {
      background: #111827;
      border: 1px solid #1f2937;
      border-radius: 12px;
      padding: 16px;
      margin-bottom: 16px;
      box-shadow: 0 10px 24px rgba(0, 0, 0, 0.28);
    }
    h1 { margin: 0 0 8px; font-size: 22px; }
    .meta { color: #9ca3af; font-size: 14px; }
    .sum { display: flex; gap: 16px; margin-top: 8px; }
    .sum span { font-weight: 700; color: #f9fafb; }
    table {
      width: 100%;
      border-collapse: collapse;
      background: #111827;
      border: 1px solid #1f2937;
      border-radius: 12px;
      overflow: hidden;
      box-shadow: 0 10px 24px rgba(0, 0, 0, 0.28);
    }
    th, td {
      text-align: left;
      padding: 10px 12px;
      border-bottom: 1px solid #1f2937;
      font-size: 14px;
    }
    th {
      background: #0f172a;
      color: #cbd5e1;
      letter-spacing: 0.01em;
    }
    tbody tr:nth-child(even) { background: #0f1a2f; }
    tbody tr:hover { background: #1a2740; }
    a { color: #93c5fd; text-decoration: none; }
    a:hover { color: #bfdbfe; text-decoration: underline; }
    .pass { color: #34d399; font-weight: 700; }
    .fail { color: #f87171; font-weight: 700; }
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
      <tr><th>Time</th><th>Status</th><th>Scenario</th><th>Topo</th><th>Cost</th></tr>
    </thead>
    <tbody>
${REPORT_ROWS_HTML}
    </tbody>
  </table>
</body>
</html>
EOF
}

write_report_md() {
  local pass_count="$1"
  local fail_count="$2"
  local total="$3"
  cat > "$REPORT_MD" <<EOF
# OpenLAN Test Report

- Title: ${REPORT_TITLE}
- Generated: $(now_text)
- Passed: ${pass_count}
- Failed: ${fail_count}
- Total: ${total}

## Scenarios

| Time | Status | Scenario | Topology | Cost |
|---|---|---|---|---|
${REPORT_ROWS_MD}
EOF
}
