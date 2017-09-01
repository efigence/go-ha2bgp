# HA2BGP

Distribute IPs into (exa-)BGP only if backend serving them is up

WiP (please do not deploy current version of it anywhere)

Goals:

* Multiple backends for healthchecks:
   * HAProxy stats socket
   * "check if socket is actually listening"
   * nagios plugins interface
   * API (socket/http)
* ExaBGP support
* Builtin GoBGP?
