# HA2BGP

Distribute IPs into (exa-)BGP only if backend serving them is up

WiP (please do not deploy current version of it anywhere)

What it does now:

* Searches for any listening socket matching filter (`--listen-filter` option, `ss` compatible filter that defaults to 80/443
* Adds any listening IP that matches one of networks (`--network` option, specify more than once for more nets)
* Checks if that listening IP actually exists in system (in case `net.ipv4.ip_nonlocal_bind=1` and IP is not [yet] up)
* Announces it (with some flapping protection)
* When listening socket or interface goes down it withdraws it after a delay (to allow for app restart).
* When interface/socket starts to flap it is also withdrawn



Goals:

* Multiple backends for healthchecks:
   * HAProxy stats socket (WiP) -  no easy haproxy side to map frontend to IP via stats socket, config parsing would be required
   * socket is actually listening - done
   * interface and IP is up - done
   * nagios plugins interface (WiP)
   * API (socket/http) (WiP)
* ExaBGP support (basics working)
* Status interface (unix pipe and/or textfile with status dump)
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

If you want to debug it just run it from commandline.
