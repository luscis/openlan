#!/bin/bash

export VERSION=v6
while true; do
  names=$(openlan na ls | grep -w 'name:' | sed 's/name://g')
  for name in $names; do
    openlan name add --name $name
  done
  sleep 5
done
