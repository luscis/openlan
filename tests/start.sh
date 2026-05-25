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
  case "$file" in
    access_success.sh) echo "two access clients authenticate and can communicate" ;;
    access_fail.sh) echo "reject client authentication with wrong password" ;;
    access_admin_multi_login.sh) echo "admin user can login concurrently from multiple access clients" ;;
    access_same_user_mutex.sh) echo "same user multiple access logins are mutually exclusive" ;;
    access_openvpn.sh) echo "add/remove OpenVPN and validate cipher negotiation" ;;
    access_openvpn_redirect.sh) echo "redirect openvpn source route to sw2 and switch vip reachability" ;;
    access_openvpn_client_ping.sh) echo "two OpenVPN clients with static addresses can ping each other" ;;
    access_openvpn_snat_vip.sh) echo "openvpn client reaches sw2 vip through sw1 snat" ;;
    access_snat_scope_matrix.sh) echo "verify snat scope matrix for openvpn, network a access, and network b access" ;;
    switch_acl.sh) echo "verify acl add-list-save-reload-remove with vip tcp/80 and icmp" ;;
    switch_acl_default_action.sh) echo "verify acl default action switch between drop and accept" ;;
    switch_bgp.sh) echo "verify bgp peer establishment and prefix filter persistence" ;;
    switch_dnat.sh) echo "verify dnat add-list-remove and nat table rule updates" ;;
    switch_findhop.sh) echo "verify findhop route binding, remove guard, and reload state" ;;
    switch_ztrust.sh) echo "verify ztrust enable/disable with guest and knock controls" ;;
    switch_tcp.sh) echo "build two switches and verify tcp output connectivity" ;;
    switch_udp.sh) echo "build two switches and verify udp output connectivity" ;;
    switch_ipsec_vxlan.sh) echo "build two switches and verify ipsec vxlan output connectivity" ;;
    switch_ipsec_gre.sh) echo "build two switches and verify ipsec gre output connectivity" ;;
    switch_ratelimit.sh) echo "verify ratelimit add-update-remove and tc state" ;;
    switch_route3.sh) echo "validate forwarding and route reachability via sw2" ;;
    *) echo "custom scenario" ;;
  esac
}

print_scenario_header() {
  local file=$1
  local name
  name=$(basename "$file")
  name=${name%.sh}
  local desc
  desc=$(scenario_description "$(basename "$file")")

  printf "${C_CYAN}[%s][RUN]${C_RESET} %-17s : %s\n" \
    "$(now_text)" "$name" "$desc"
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
        if [[ "$REPORT_ENABLED" == "true" ]]; then
          case_log=$(report_case_log_file "$index" "$name")
          echo "[$(now_text)] case log: $case_log"
          {
            echo "[$(now_text)] START $name"
            echo "scenario: $file"
            echo "capture: stdout+stderr"
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
          report_line "[$status] $name cost=$cost"
          report_line "  log: $case_log"
          report_case_html "$status" "$name" "$cost" "$case_log"
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
        } > "$case_log"
        report_line "[FAIL] unknown scenario: $key"
        report_line "  log: $case_log"
        report_case_html "FAIL" "$key (unknown)" "0.000s" "$case_log"
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
    report_line ""
    report_line "Summary:"
    report_line "  Passed: $pass_count"
    report_line "  Failed: $fail_count"
    report_line "  Total:  $total"
    report_line "End: $(now_text)"
    report_line "Report: $REPORT_FILE"
    write_report_html "$pass_count" "$fail_count" "$total"
    printf "[%s] Report generated: %s\n" "$(now_text)" "$REPORT_FILE"
    printf "[%s] HTML report: %s\n" "$(now_text)" "$REPORT_HTML"
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
