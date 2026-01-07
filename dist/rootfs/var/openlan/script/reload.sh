#!/bin/bash

set -ex

backup=$(pidof openlan-switch | awk '{print $1}')
if [ $backup -ne 1 ]; then
	kill $backup
fi