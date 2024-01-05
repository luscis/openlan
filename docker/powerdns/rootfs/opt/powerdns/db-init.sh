#!/bin/bash

set -ex

[ -e "/var/lib/powerdns/pdns.sqlite3" ] || {
	sqlite3 /var/lib/powerdns/pdns.sqlite3 < /var/lib/powerdns/schema.sqlite3.sql
}