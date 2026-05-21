
## variables

export ARCH=amd64

## functions
version() {
  $PWD/../dist/version.sh
}

export VERSION=$(version)
export IMAGE="luscis/openlan:$VERSION.$ARCH.deb"

flush() {
  local out=$1

  while IFS= read -r line; do
    echo "$line" >> $out
  done
}

wait() {
  set +x
  local cmd=$1; local match=$2
  local count=$3; local code=1
  local out=/tmp/wait.1

  if [[ $count == "" ]]; then
    count=3
  fi

  $cmd 2>&1 | flush $out & disown
  local pid=$!

  for i in $(seq 1 $count); do
    if [ ! -e $out ]; then
      sleep 1
      continue
    fi
    if cat $out | grep "$match" -C 3; then
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
  set -x
  return $code
}

check() {
  set +x
  local cmd=$1; local match=$2
  local count=$3; local code=1
  local out=/tmp/check.1

  if [[ $count == "" ]]; then
    count=3
  fi

  for i in $(seq 1 $count); do
    rm -f $out
    if $cmd > $out 2>&1; then
      :
    fi
    if cat $out | grep "$match" -C 3; then
      code=0
      break
    fi
    sleep 1
  done

  rm -f $out
  set -x
  return $code
}

pause() {
  echo "Press ENTER to continue: "
  read
}

main() {
  setup
  if [[ $PAUSE == true ]]; then
    pause
  fi
  cleanup
}

start_switch() {
  local name=$1
  local network_name=$2
  local address=$3

  # Start a paused switch container with openlan and ipsec.
  docker run -d --rm --privileged \
    --network $network_name --ip $address \
    --volume /opt/openlan/$name/etc/openlan:/etc/openlan \
    --volume /opt/openlan/$name/etc/ipsec.d:/etc/ipsec.d \
    --volume /opt/openlan/$name/run/pluto:/run/pluto \
    --name $name-pause $IMAGE \
    /bin/bash -c "trap : TERM; sleep infinity & wait"
  # Start switch:
  docker run -d --rm --privileged \
    --network container:$name-pause \
    --volume /opt/openlan/$name/etc/openlan:/etc/openlan \
    --volume /opt/openlan/$name/etc/ipsec.d:/etc/ipsec.d \
    --volume /opt/openlan/$name/run/pluto:/run/pluto \
    --name $name $IMAGE \
    /var/openlan/script/switch.sh
  # Start ipsec.
  docker run -d --rm --privileged \
    --network container:$name-pause \
    --volume /opt/openlan/$name/etc/openlan:/etc/openlan \
    --volume /opt/openlan/$name/etc/ipsec.d:/etc/ipsec.d \
    --volume /opt/openlan/$name/run/pluto:/run/pluto \
    --name $name-ipsec $IMAGE \
    /var/openlan/script/ipsec.sh
}

start_access() {
  local name=$1
  local network_name=$2

  # Start access point.
  docker run -d --rm --privileged \
    --network $network_name \
    --volume /opt/openlan/$name/etc/openlan:/etc/openlan \
    --name $name $IMAGE \
    /usr/bin/openlan-access -conf /etc/openlan/access.yaml
}

start_openvpn() {
  local name=$1
  local network_name=$2

  # Start OpenVPN client.
  docker run -d --rm --cap-add=NET_ADMIN \
    --device /dev/net/tun --network $network_name \
    --volume /opt/openlan/$name/ovpn:/ovpn \
    --name $name $IMAGE \
    /usr/sbin/openvpn --config /ovpn/client.ovpn --auth-user-pass /ovpn/auth.txt --verb 3
}

cleanup() {
  set +x
  local containers
  local networks

  containers=$(docker ps -aq --filter "name=^tests-" 2>/dev/null)
  if [[ -n "$containers" ]]; then
    docker rm -f $containers >/dev/null 2>&1 || true
  fi
  networks=$(docker network ls -q --filter "name=^tests-" 2>/dev/null)
  if [[ -n "$networks" ]]; then
    docker network rm $networks >/dev/null 2>&1 || true
  fi

  rm -rf /opt/openlan/tests-*
  set -x
}
