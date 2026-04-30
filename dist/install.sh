#!/bin/bash

set -euo pipefail

installer="$0"
nosysm="no"
nodeps="no"
dry_run="no"
verbose="no"

tmp=""
sys="linux"
archive=""

function log() {
    local level="$1"
    shift
    printf '[%s] %s\n' "$level" "$*"
}

function info() {
    log "INFO" "$*"
}

function warn() {
    log "WARN" "$*" >&2
}

function die() {
    log "ERROR" "$*" >&2
    exit 1
}

function usage() {
    cat <<EOF
Usage: $0 [OPTIONS]

Options:
  --nodeps        Skip dependency installation and post initialization.
  --dry-run       Print commands without executing them.
  --verbose       Print each command before execution.
  --nosysm        Skip system detection and use default settings.
  -h, --help      Show this help message.

Legacy:
  nodeps          Same as --nodeps (kept for compatibility).
EOF
}

function parse_args() {
    while [ "$#" -gt 0 ]; do
        case "$1" in
            --nodeps|nodeps)
                nodeps="yes"
                nosysm="yes"
                ;;
            --nosysm|nosysm)
                nosysm="yes"
                ;;
            --dry-run)
                dry_run="yes"
                ;;
            --verbose)
                verbose="yes"
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                die "Unknown argument: $1. Use --help for usage."
                ;;
        esac
        shift
    done
}

function run_cmd() {
    if [ "$verbose" == "yes" ] || [ "$dry_run" == "yes" ]; then
        info "CMD: $*"
    fi
    if [ "$dry_run" == "yes" ]; then
        return 0
    fi
    "$@"
}

function cleanup() {
    if [ -n "${tmp:-}" ] && [ -d "$tmp" ]; then
        if [ "$dry_run" == "yes" ]; then
            info "DRY-RUN: rm -rf $tmp"
        else
            rm -rf "$tmp"
        fi
    fi
}

trap cleanup EXIT

function find_sys() {
    if command -v yum >/dev/null 2>&1; then
        sys="redhat"
    elif command -v apt >/dev/null 2>&1; then
        sys="debian"
    else
        die "Unsupported system: neither yum nor apt is available."
    fi
}

function download() {
    info "Uncompress files ..."
    tmp=$(mktemp -d)
    if [ "$dry_run" == "yes" ]; then
        info "DRY-RUN: extract embedded archive to $tmp"
        return 0
    fi
    archive="$(grep -a -n "__ARCHIVE_BELOW__:$" "$installer" | cut -f1 -d:)"
    [ -n "$archive" ] || die "Unable to locate embedded archive marker."
    tail -n +$((archive + 1)) "$installer" | gzip -dc - | tar -xf - -C "$tmp"
}

function requires() {
    info "Install dependencies ..."
    ## Install packages from repo.
    if [ "$sys"x == "redhat"x ]; then
        run_cmd yum update -y
        run_cmd yum install -y epel-release
        run_cmd yum install -y openssl net-tools iptables iputils iperf3 tcpdump
        run_cmd yum install -y openvpn dnsmasq bridge-utils ipset procps wget socat
    elif [ "$sys"x == "debian"x ]; then
        run_cmd apt-get update -y
        run_cmd apt-get install -y net-tools iptables iproute2 tcpdump ca-certificates iperf3 socat
        run_cmd apt-get install -y openvpn dnsmasq bridge-utils ipset procps wget iputils-ping frr
    fi
    ## Install libreswan from github.
    if [ "$sys"x == "redhat"x ]; then
        run_cmd wget -O /tmp/libreswan-4.10-1.el7.x86_64.rpm https://github.com/luscis/packages/raw/main/redhat/centos7/libreswan-4.10-1.el7.x86_64.rpm
        if ! run_cmd yum install -y /tmp/libreswan-4.10-1.el7.x86_64.rpm; then
            warn "Install local libreswan rpm failed, fallback to repo package."
            run_cmd yum install -y libreswan
        fi
        run_cmd wget -O /tmp/frr-stable-repo.el7.noarch.rpm https://rpm.frrouting.org/repo/frr-stable-repo.el7.noarch.rpm
        run_cmd yum install -y /tmp/frr-stable-repo.el7.noarch.rpm
        run_cmd yum install -y frr frr-pythontools
    elif [ "$sys"x == "debian"x ]; then
        run_cmd apt-get install -y libreswan
    fi
}

function install() {
    info "Installing files ..."
    if [ "$dry_run" == "yes" ]; then
        info "DRY-RUN: copy etc/usr/var to / and refresh /usr/share/openlan.db"
        return 0
    fi
    local source
    source="$(find "$tmp" -maxdepth 1 -name 'openlan-*' | head -n1)"
    [ -n "$source" ] || die "Unable to find extracted openlan directory."
    pushd "$source" >/dev/null
    run_cmd /usr/bin/env cp -rf ./etc ./usr ./var /
    run_cmd chmod +x /var/openlan/script/*.sh
    run_cmd /usr/bin/env sh -c 'find ./ -type f > /usr/share/openlan.db'
    popd >/dev/null
}

function post() {
    info "Initializing ..."
    if [ "$sys"x == "redhat"x ]; then
        ## Prepare openvpn.
        [ -e "/var/openlan/openvpn/dh.pem" ] || {
            run_cmd openssl dhparam -out /var/openlan/openvpn/dh.pem 1024
        }
        [ -e "/var/openlan/openvpn/ta.key" ] || {
            run_cmd openvpn --genkey --secret /var/openlan/openvpn/ta.key
        }
        ## Install CA.
        run_cmd cp -rf /var/openlan/cert/ca.crt /etc/pki/ca-trust/source/anchors/OpenLAN_CA.crt
        run_cmd update-ca-trust
    elif [ "$sys"x == "debian"x ]; then
        ## Prepare openvpn.
        [ -e "/var/openlan/openvpn/dh.pem" ] || {
            run_cmd openssl dhparam -out /var/openlan/openvpn/dh.pem 2048
        }
        [ -e "/var/openlan/openvpn/ta.key" ] || {
            if [ "$dry_run" == "yes" ]; then
                info "DRY-RUN: openvpn --genkey > /var/openlan/openvpn/ta.key"
            else
                openvpn --genkey > /var/openlan/openvpn/ta.key
            fi
        }
        ## Install CA.
        run_cmd cp -rf /var/openlan/cert/ca.crt /usr/local/share/ca-certificates/OpenLAN_CA.crt
        run_cmd update-ca-certificates
    fi
}

function finish() {
    if [ x"$nosysm" == x"no" ] || [ x"$nosysm" == x"" ]; then
        run_cmd systemctl daemon-reload
        run_cmd systemctl disable --now apparmor || true
    fi
    info "Finished."
}

parse_args "$@"

if [ "$dry_run" == "yes" ]; then
    info "Dry-run mode enabled. Commands will be printed only."
fi
if [ "$verbose" == "yes" ]; then
    info "Verbose mode enabled."
fi

if [ "$dry_run" != "yes" ] && [ "$(id -u)" -ne 0 ]; then
    die "Please run this installer as root."
fi

find_sys
download
if [ "$nodeps"x == "no"x ]; then
    requires
fi
install
if [ "$nodeps"x == "no"x ]; then
    post
fi
finish
exit 0
