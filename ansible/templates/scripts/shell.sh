#!/bin/bash
#
# This script wraps wjcli into shell-like environment

unset "$PATH"

# Main executable
PROGRAM="{{ wirejump.basedir }}/bin/wjcli"

# Since this is still technically bash,
# disallow everything but param string:
# it should allow -some --params, maybe
# something "quoted" or 'not', as well
# as base64 symbols of a public key
BASE64CHARS="\+\=\/"
PARAMCHARS="\"\'\-"
PATTERN="^[[:alnum:][:blank:]${BASE64CHARS}${PARAMCHARS}]{0,128}$"

# Proceed if there's no fancy stuff
if [[ "$SSH_ORIGINAL_COMMAND" =~ $PATTERN ]]; then
    read -r -a PARAMS <<< "$SSH_ORIGINAL_COMMAND"
	exec $PROGRAM "${PARAMS[@]}"
else
	echo "Invalid command: $SSH_ORIGINAL_COMMAND"
	exit 1
fi

exit 0