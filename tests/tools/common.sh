
## variables

ASSERT_SEQ=0
ASSERT_ID=""

_assert_next_id() {
  ASSERT_SEQ=$((ASSERT_SEQ + 1))
  ASSERT_ID=$(printf "%04d" "$ASSERT_SEQ")
}

_flush() {
  local out=$1

  while IFS= read -r line; do
    echo "$line" >> $out
  done
}

# wait for a command to output expected string, with retry count.
# and count is max retry timeout, so it will be more stable in CI with variable performance.
_wait() {
  local count=$1; local cmd=$2; local match=$3; 
  local code=1
  local out=/tmp/_wait.1

  $cmd 2>&1 | _flush $out & disown
  local pid=$!

  for i in $(seq 1 $count); do
    if [ ! -e $out ]; then
      sleep 1
      continue
    fi
    if cat $out | grep "$match" -C 3 --color; then
      code=0; break
    fi
    if ! ps -p $pid > /dev/null; then
      break
    fi
    sleep 1
  done

  if ps -p $pid > /dev/null; then
    kill $pid;
  fi
  rm -f $out
  return $code
}

# the count is total retry times, not total timeout, so it will be more stable in CI with variable performance.
_check() {
  local count=$1; local cmd=$2; local match=$3; 
  local code=1
  local out=/tmp/_check.1

  for i in $(seq 1 $count); do
    rm -f $out
    if $cmd > $out 2>&1; then
      :
    fi
    if cat $out | grep "$match" -C 3 --color; then
      code=0
      break
    fi
    sleep 1
  done

  rm -f $out
  
  return $code
}

# the count is total retry times, not total timeout, so it will be more stable in CI with variable performance.
_check_fuzzy() {
  local count=$1; local cmd=$2; local pattern=$3
  local code=1
  local out=/tmp/_check_fuzzy.1

  for i in $(seq 1 $count); do
    rm -f $out
    if $cmd > $out 2>&1; then
      :
    fi
    if cat $out | grep -Ei "$pattern" -C 3 --color; then
      code=0
      break
    fi
    sleep 1
  done

  rm -f $out
  return $code
}

assert_match() {
  local count="$1"; local cmd="$2"; local match="$3"
  local script="${BASH_SOURCE[1]}"
  local line="${BASH_LINENO[0]}"
  local func="${FUNCNAME[1]}"

  _assert_next_id
  local aid="$ASSERT_ID"
  local start_ms
  local cost
  local now_text
  local end_text
  start_ms=$(now_ms)
  now_text=$(now_text)
  echo "${C_CYAN}[${now_text}][ASSERT#${aid}][match] at ${script}:${line} fn=${func} retry=${count} cmd=\"${cmd}\" expect=\"${match}\"${C_RESET}"
  if ! _check "$count" "$cmd" "$match"; then
    cost=$(cost_s "$start_ms")
    end_text=$(now_text)
    echo "${C_RED}[${end_text}][ASSERT#${aid}][FAIL] cost=${cost}${C_RESET}"
    exit 1
  fi
  cost=$(cost_s "$start_ms")
  end_text=$(now_text)
  echo "${C_GREEN}[${end_text}][ASSERT#${aid}][OK] cost=${cost}${C_RESET}"
}

assert_fuzzy() {
  local count="$1"; local cmd="$2"; local pattern="$3"
  local script="${BASH_SOURCE[1]}"
  local line="${BASH_LINENO[0]}"
  local func="${FUNCNAME[1]}"

  _assert_next_id
  local aid="$ASSERT_ID"
  local start_ms
  local cost
  local now_text
  local end_text
  start_ms=$(now_ms)
  now_text=$(now_text)
  echo "${C_CYAN}[${now_text}][ASSERT#${aid}][fuzzy] at ${script}:${line} fn=${func} retry=${count} cmd=\"${cmd}\" pattern=\"${pattern}\"${C_RESET}"
  if ! _check_fuzzy "$count" "$cmd" "$pattern"; then
    cost=$(cost_s "$start_ms")
    end_text=$(now_text)
    echo "${C_RED}[${end_text}][ASSERT#${aid}][FAIL] cost=${cost}${C_RESET}"
    exit 1
  fi
  cost=$(cost_s "$start_ms")
  end_text=$(now_text)
  echo "${C_GREEN}[${end_text}][ASSERT#${aid}][OK] cost=${cost}${C_RESET}"
}

assert_unmatch() {
  local count="$1"; local cmd="$2"; local match="$3"
  local script="${BASH_SOURCE[1]}"
  local line="${BASH_LINENO[0]}"
  local func="${FUNCNAME[1]}"

  _assert_next_id
  local aid="$ASSERT_ID"
  local start_ms
  local cost
  local now_text
  local end_text
  start_ms=$(now_ms)
  now_text=$(now_text)
  echo "${C_CYAN}[${now_text}][ASSERT#${aid}][unmatch] at ${script}:${line} fn=${func} retry=${count} cmd=\"${cmd}\" unexpected=\"${match}\"${C_RESET}"
  if _check "$count" "$cmd" "$match"; then
    cost=$(cost_s "$start_ms")
    end_text=$(now_text)
    echo "${C_RED}[${end_text}][ASSERT#${aid}][FAIL] cost=${cost}${C_RESET}"
    exit 1
  fi
  cost=$(cost_s "$start_ms")
  end_text=$(now_text)
  echo "${C_GREEN}[${end_text}][ASSERT#${aid}][OK] cost=${cost}${C_RESET}"
}

assert_expect() {
  local count="$1"; local cmd="$2"; local match="$3"
  local script="${BASH_SOURCE[1]}"
  local line="${BASH_LINENO[0]}"
  local func="${FUNCNAME[1]}"

  _assert_next_id
  local aid="$ASSERT_ID"
  local start_ms
  local cost
  local now_text
  local end_text
  start_ms=$(now_ms)
  now_text=$(now_text)
  echo "${C_CYAN}[${now_text}][ASSERT#${aid}][expect] at ${script}:${line} fn=${func} retry=${count} cmd=\"${cmd}\" expect=\"${match}\"${C_RESET}"
  if ! _wait "$count" "$cmd" "$match"; then
    cost=$(cost_s "$start_ms")
    end_text=$(now_text)
    echo "${C_RED}[${end_text}][ASSERT#${aid}][FAIL] cost=${cost}${C_RESET}"
    exit 1
  fi
  cost=$(cost_s "$start_ms")
  end_text=$(now_text)
  echo "${C_GREEN}[${end_text}][ASSERT#${aid}][OK] cost=${cost}${C_RESET}"
}

assert_unexpect() {
  local count="$1"; local cmd="$2"; local match="$3"
  local script="${BASH_SOURCE[1]}"
  local line="${BASH_LINENO[0]}"
  local func="${FUNCNAME[1]}"

  _assert_next_id
  local aid="$ASSERT_ID"
  local start_ms
  local cost
  local now_text
  local end_text
  start_ms=$(now_ms)
  now_text=$(now_text)
  echo "${C_CYAN}[${now_text}][ASSERT#${aid}][unexpect] at ${script}:${line} fn=${func} retry=${count} cmd=\"${cmd}\" unexpected=\"${match}\"${C_RESET}"
  if _wait "$count" "$cmd" "$match"; then
    cost=$(cost_s "$start_ms")
    end_text=$(now_text)
    echo "${C_RED}[${end_text}][ASSERT#${aid}][FAIL] cost=${cost}${C_RESET}"
    exit 1
  fi
  cost=$(cost_s "$start_ms")
  end_text=$(now_text)
  echo "${C_GREEN}[${end_text}][ASSERT#${aid}][OK] cost=${cost}${C_RESET}"
}

assert_cmd() {
  local script="${BASH_SOURCE[1]}"
  local line="${BASH_LINENO[0]}"
  local func="${FUNCNAME[1]}"

  _assert_next_id
  local aid="$ASSERT_ID"
  local start_ms
  local cost
  local now_text
  local end_text
  start_ms=$(now_ms)
  now_text=$(now_text)
  echo "${C_CYAN}[${now_text}][ASSERT#${aid}][cmd] at ${script}:${line} fn=${func} cmd=\"$*\"${C_RESET}"
  if ! "$@"; then
    cost=$(cost_s "$start_ms")
    end_text=$(now_text)
    echo "${C_RED}[${end_text}][ASSERT#${aid}][FAIL] cost=${cost}${C_RESET}"
    exit 1
  fi
  cost=$(cost_s "$start_ms")
  end_text=$(now_text)
  echo "${C_GREEN}[${end_text}][ASSERT#${aid}][OK] cost=${cost}${C_RESET}"
}

assert_fail_cmd() {
  local script="${BASH_SOURCE[1]}"
  local line="${BASH_LINENO[0]}"
  local func="${FUNCNAME[1]}"

  _assert_next_id
  local aid="$ASSERT_ID"
  local start_ms
  local cost
  local now_text
  local end_text
  start_ms=$(now_ms)
  now_text=$(now_text)
  echo "${C_CYAN}[${now_text}][ASSERT#${aid}][cmd_fail] at ${script}:${line} fn=${func} cmd=\"$*\"${C_RESET}"
  if "$@"; then
    cost=$(cost_s "$start_ms")
    end_text=$(now_text)
    echo "${C_RED}[${end_text}][ASSERT#${aid}][FAIL] cost=${cost}${C_RESET}"
    exit 1
  fi
  cost=$(cost_s "$start_ms")
  end_text=$(now_text)
  echo "${C_GREEN}[${end_text}][ASSERT#${aid}][OK] cost=${cost}${C_RESET}"
}

_pause() {
  echo "Press ENTER to continue: "
  read
}

main() {
  _cleanup
  setup
  if [[ $PAUSE == true ]]; then
    _pause
  fi
  _cleanup
}

start_switch() {
  local name=$1; local network_name=$2; local address=$3
  local volume_opts="--volume /opt/openlan/$name/etc/openlan:/etc/openlan --volume /opt/openlan/$name/etc/ipsec.d:/etc/ipsec.d --volume /opt/openlan/$name/run/pluto:/run/pluto --volume /opt/openlan/$name/var/openlan/frr:/var/openlan/frr --volume /opt/openlan/$name/etc/frr:/etc/frr"

  # Start a paused switch container with openlan and ipsec.
  if docker run -d --rm --privileged --network $network_name --ip $address $volume_opts --name $name-pause $IMAGE /bin/bash -c "trap : TERM; sleep infinity & wait" >/dev/null; then
    echo "Started switch pause container: $name-pause"
  fi
  # Start frr.
  if docker run -d --rm --privileged --network container:$name-pause $volume_opts --name $name-frr $IMAGE /var/openlan/script/frr.sh >/dev/null; then
    echo "Started switch frr container: $name-frr"
  fi
  # Start ipsec.
  if docker run -d --rm --privileged --network container:$name-pause $volume_opts --name $name-ipsec $IMAGE /var/openlan/script/ipsec.sh >/dev/null; then
    echo "Started switch ipsec container: $name-ipsec"
  fi
  # Start switch:
  if docker run -d --rm --privileged --network container:$name-pause $volume_opts --name $name $IMAGE /var/openlan/script/switch.sh >/dev/null; then
    echo "Started switch container: $name"
  fi
}

start_access() {
  local name=$1; local network_name=$2

  # Start access point.
  if docker run -d --rm --privileged --network $network_name --volume /opt/openlan/$name/etc/openlan:/etc/openlan --name $name $IMAGE /usr/bin/openlan-access -conf /etc/openlan/access.yaml >/dev/null; then
    echo "Started access container: $name"
  fi
}

start_openvpn() {
  local name=$1; local network_name=$2

  # Start OpenVPN client.
  if docker run -d --rm --cap-add=NET_ADMIN --device /dev/net/tun --network $network_name --volume /opt/openlan/$name/ovpn:/ovpn --name $name $IMAGE /usr/sbin/openvpn --config /ovpn/client.ovpn --auth-user-pass /ovpn/auth.txt --verb 3 >/dev/null; then
    echo "Started OpenVPN client container: $name"
  fi
}

stop_switch() {
  local name=$1
  docker exec $name openlan config save
  if docker rm -f $name $name-pause $name-ipsec $name-frr >/dev/null || true; then
    echo "Stopped switch containers: $name, $name-pause, $name-ipsec, $name-frr"
  fi
}

_cleanup() {
  local containers
  local networks

  containers=$(docker ps -aq --filter "name=^tests-" 2>/dev/null)
  if [[ -n "$containers" ]]; then
    docker rm -f $containers >/dev/null || true
  fi
  networks=$(docker network ls -q --filter "name=^tests-" 2>/dev/null)
  if [[ -n "$networks" ]]; then
    docker network rm $networks >/dev/null || true
  fi

  rm -rf /opt/openlan/tests-*
}
