
## variables

export ARCH=amd64
export VERSION=unknown

## functions
version() {
  $PWD/../dist/version.sh
}

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
    if cat $out | grep "$match"; then
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

pause() {
  echo "Press ENTER to continue: "
  read
}

export VERSION=$(version)
export IMAGE="luscis/openlan:$VERSION.$ARCH.deb"

main() {
  setup
  if [[ $PAUSE == true ]]; then
    pause
  fi
  cleanup
}