# This script refreshes (reconnects) your upstream connection.
# Add this script to the scheduler and execute regularly, depending on your needs.

:do {
    :local result [/system/ssh-exec address=172.16.1.1 user=manager command="connect" as-value]

    :if (([$result]->"exit-code") = 0) do={
        :log info "WireJump: reconnected"
    } else={
        :log error "WireJump: failed to reconnect"
        :log error ([$result]->"output")
    }
} on-error={
    :log info "WireJump: failed to exec SSH command"
}
