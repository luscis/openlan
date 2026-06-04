#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
pushd "$SCRIPT_DIR" > /dev/null

source ./tools/auto.sh

shopt -s nullglob
scenarios=( cases/*.sh )
shopt -u nullglob

report_init_env "$SCRIPT_DIR"
REPORT_ENABLED=false

scenario_description() {
  local file=$1
  local path="$file"
  local desc
  if [[ ! -f "$path" ]]; then
    path="cases/$file"
  fi
  desc=$(bash "$path" --description 2>/dev/null)
  if [[ -z "$desc" ]]; then
    desc="custom scenario"
  fi
  printf "%s\n" "$desc"
}

scenario_topology_detail() {
  local file=$1
  local detail
  detail=$(bash "$file" --topology 2>/dev/null)
  if [[ -z "$detail" ]]; then
    cat <<'EOF'
# Topology:
# - custom topology
EOF
    return
  fi
  printf "%s\n" "$detail"
}

scenario_topology_summary() {
  local file=$1
  local path="$file"
  local summary
  if [[ ! -f "$path" ]]; then
    path="cases/$file"
  fi
  summary=$(bash "$path" --summary 2>/dev/null)
  if [[ -z "$summary" ]]; then
    summary="custom topology"
  fi
  printf "%s\n" "$summary"
}

print_scenario_header() {
  local file=$1
  local name
  name=$(basename "$file")
  name=${name%.sh}
  local desc
  desc=$(scenario_description "$(basename "$file")")
  local topo_detail
  topo_detail=$(scenario_topology_detail "$file")
  local topo
  topo=$(scenario_topology_summary "$file")

  printf "${C_CYAN}[%s][RUN]${C_RESET} %-17s : %s\n" \
    "$(now_text)" "$name" "$desc"
  printf "%s\n" "$topo_detail"
}

run_all() {
  run_batch "OpenLAN test run started" "${scenarios[@]}"
}

run_batch() {
  local title=$1
  shift
  local total=$#
  local pass_count=0
  local fail_count=0
  local index=0
  local key
  local file
  local found
  local name
  local passed=""
  local failed=""
  local start_ms
  local cost
  local status
  local case_log
  local topo
  local topo_detail

  if [[ "$REPORT_ENABLED" == "true" ]]; then
    init_report
    write_report_header "$title"
  fi
  echo "========================================"
  printf "[%s] %s: %d scenario(s)\n" "$(now_text)" "$title" "$total"
  echo "========================================"

  for key in "$@"; do
    found=0
    for file in "${scenarios[@]}"; do
      name=$(basename "$file")
      name=${name%.sh}
      if [[ "$key" == "$name" || "$key" == "$file" || "$key" == "${name}.sh" ]]; then
        found=1
        index=$((index + 1))
        printf "\n[%s][%02d/%02d] %s\n" "$(now_text)" "$index" "$total" "$name"
        echo "----------------------------------------"
        print_scenario_header "$file"
        start_ms=$(now_ms)
        topo_detail=$(scenario_topology_detail "$file")
        topo=$(scenario_topology_summary "$file")
        if [[ "$REPORT_ENABLED" == "true" ]]; then
          case_log=$(report_case_log_file "$index" "$name")
          echo "[$(now_text)] case log: $case_log"
          {
            echo "[$(now_text)] START $name"
            echo "scenario: $file"
            echo "header  : $(scenario_description "$(basename "$file")")"
            echo "topology: $topo"
            printf "%s\n" "$topo_detail" | sed 's/^/topology: /'
            echo ""
          } > "$case_log"
          bash "$file" >> "$case_log" 2>&1
          local rc=$?
        else
          bash "$file"
          local rc=$?
        fi
        if [[ $rc -eq 0 ]]; then
          cost=$(cost_s "$start_ms")
          printf "${C_GREEN}[%s][PASS]${C_RESET} %s cost=%s\n" \
            "$(now_text)" "$name" "$cost"
          status="PASS"
          pass_count=$((pass_count + 1))
          passed="$passed $name"
        else
          cost=$(cost_s "$start_ms")
          printf "${C_RED}[%s][FAIL]${C_RESET} %s cost=%s\n" \
            "$(now_text)" "$name" "$cost"
          status="FAIL"
          fail_count=$((fail_count + 1))
          failed="$failed $name"
        fi
        if [[ "$REPORT_ENABLED" == "true" ]]; then
          echo "[$(now_text)] END $name status=$status cost=$cost" >> "$case_log"
          report_case_html "$status" "$name" "$cost" "$topo" "$case_log"
          report_case_md "$status" "$name" "$cost" "$topo" "$case_log"
        fi
        break
      fi
    done

    if [[ $found -eq 0 ]]; then
      index=$((index + 1))
      printf "\n[%s][%02d/%02d] %s\n" "$(now_text)" "$index" "$total" "$key"
      echo "----------------------------------------"
      printf "${C_RED}[%s][FAIL]${C_RESET} unknown scenario: %s\n" \
        "$(now_text)" "$key"
      if [[ "$REPORT_ENABLED" == "true" ]]; then
        case_log=$(report_case_log_file "$index" "$key")
        {
          echo "[$(now_text)] START $key"
          echo "status: FAIL"
          echo "reason: unknown scenario"
          echo "topology: custom topology"
        } > "$case_log"
        report_case_html "FAIL" "$key (unknown)" "0.000s" "custom topology" "$case_log"
        report_case_md "FAIL" "$key (unknown)" "0.000s" "custom topology" "$case_log"
      fi
      echo "Available scenarios:"
      for file in "${scenarios[@]}"; do
        name=$(basename "$file")
        echo "  - ${name%.sh}"
      done
      echo "========================================"
      return 1
    fi
  done

  echo
  echo "========================================"
  printf "[%s] Summary: %d passed, %d failed, %d total\n" \
    "$(now_text)" "$pass_count" "$fail_count" "$total"
  if [[ -n "$passed" ]]; then
    echo "Passed scenarios:"
    echo "$passed" | xargs printf "  - %s\n"
  fi
  if [[ -n "$failed" ]]; then
    echo "Failed scenarios:"
    echo "$failed" | xargs printf "  - %s\n"
  fi
  echo "========================================"
  if [[ "$REPORT_ENABLED" == "true" ]]; then
    write_report_html "$pass_count" "$fail_count" "$total"
    write_report_md "$pass_count" "$fail_count" "$total"
    pack_report_tar
    printf "[%s] Markdown report: %s\n" "$(now_text)" "$REPORT_MD"
    printf "[%s] HTML report: %s\n" "$(now_text)" "$REPORT_HTML"
    printf "[%s] TAR report: %s\n" "$(now_text)" "$REPORT_TAR"
  fi
  if [[ $fail_count -gt 0 ]]; then
    return 1
  fi
}

run_selected() {
  run_batch "OpenLAN selected run started" "$@"
}


## Running senarios:
## - run_all: run all scenarios
## - run_batch "title" scenario1 scenario2 ...: run a batch of scenarios with a title
## - run_selected scenario1 scenario2 ...: run selected scenarios by name or file 
## Usage: 
## - bash tests/start.sh
if [[ $# -eq 0 ]]; then
  run_all
elif [[ "$1" == "list" || "$1" == "--list" || "$1" == "-l" ]]; then
  for file in "${scenarios[@]}"; do
    name=$(basename "$file")
    printf "%-17s : %s\n" "${name%.sh}" "$(scenario_description "$name")"
  done
elif [[ "$1" == "help" || "$1" == "--help" || "$1" == "-h" ]]; then
  echo "usage: bash tests/start.sh [scenario ...]"
  echo "examples:"
  echo "  bash tests/start.sh"
  echo "  bash tests/start.sh switch_tcp"
  echo "  bash tests/start.sh access_openvpn access_success"
  echo "  bash tests/start.sh --list"
else
  if [[ "$1" == "--report" ]]; then
    REPORT_ENABLED=true
    shift
  fi
  if [[ $# -eq 1 ]]; then
    run_batch "OpenLAN selected run started" "$1"
  elif [[ $# -eq 0 ]]; then
    run_all
  else
    run_selected "$@"
  fi
fi

popd > /dev/null
