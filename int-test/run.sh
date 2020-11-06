#!/usr/bin/env bash

# This script runs a quick integration test using MailHog as the SMTP server and
# socat to provide a TLS wrapper for TLS integratoin tests.
# 
# 	https://github.com/mailhog/MailHog
#
# If you wish to run these tests, ensure mailhog and socat are in your path.
# You'll probably need OpenSSL too.
# 
# Results must be verified manually, either with the UI or the MailHog API:
# 
# 	curl http://127.0.0.1${MAILHOG_API_PORT}/api/v2/messages -s | \
# 		jq '.total, .items[].Content.Headers.Subject'
# 
# 

# Define the ports the services listen on
SMTP_PORT=${SMTP_PORT:="7025"}
TLS_PORT=${TLS_PORT:="7026"}

# PID files for cleanup
MAILHOG_PID="$(pwd)/mailhog.pid"
SOCAT_PID="$(pwd)/socat.pid"

INT_DIR=$(dirname "$(realpath -s "$0")")

# kill -9 a process with a pidfile at the first argument.
function stop_pidfile() {
	pid_file=$1
	if [ -f "${pid_file}" ]; then
		kill -9 "$(cat "${pid_file}")" || true
		rm "${pid_file}"
	fi
}

function cleanup() {
	stop_pidfile "${MAILHOG_PID}";
	stop_pidfile "${SOCAT_PID}";
}
trap cleanup EXIT

mailhog -smtp-bind-addr=127.0.0.1:${SMTP_PORT} & echo "$!" > "${MAILHOG_PID}"
socat -v openssl-listen:${TLS_PORT},cert="${INT_DIR}/cert.pem",verify=0,reuseaddr,fork tcp4:127.0.0.1:${SMTP_PORT} & echo "$!" > "${SOCAT_PID}"

wait "$(cat "${MAILHOG_PID}")"