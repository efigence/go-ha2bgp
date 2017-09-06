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

Usage:

In exabgp.conf:

    group bgpproxy {
        router-id 10.1.1.1;
        neighbor 127.0.1.1 {
          local-address 127.0.1.2;
          local-as 65000;
          peer-as 65000;
        }
        process do-bgp-stuff {
            run /usr/local/bin/ha2bgp -network 100.64.0.0/24 ;
        }
    }


in Bird:

    protocol bgp exabgp {
      local as 65000;
      import all;
      export none;
      local 127.0.1.1 as 65000;
      neighbor 127.0.1.2 as 65000;

    };

Then just redistribute it to OSPF or wherever you need:

    filter core_export {
      if net ~ [ 100.64.0.0/24+ ] then accept;
      else reject;
    };
    protocol ospf core {
      router id 1.2.3.4;
      ecmp;
      import all;
      export filter core_export;
      area 0 {
          interface "eth2.110" {};
          interface "eth3.111" {};
      };
    };


Note that you should add import filter on bird to only allow *your* networks (to avoid any mistakes).
