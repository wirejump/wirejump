#!/bin/bash
#
# This scripts updates upstream servers list so it's always fresh
#

# query upstream servers and silence output
exec "{{ wirejump.basedir }}/bin/wjcli servers" > /dev/null 2>&1

# failure here is not important, as it means
# that provider has not been selected yet
exit 0
