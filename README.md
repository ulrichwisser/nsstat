# nsstat
Statistics on hosts in a zone file
The following numbers will be computed

- Number of in zone hosts without glue
- Number of in zone hosts without glue and which do not resolv
- Number of in zone hosts with glue
- Number of in zone hosts with glue and which do not resolv
- Number of in zone hosts with glue where glue and resolv ip addresses do not match
- Number of ex zone hosts
- Number of ex zone hosts and which do not resolv

The script does make use of go routines to resolv hostnames to ip addresses.

The script depends heavily on the performance of your resolver(s).

All .go files are covered by the GNU GPL 3.0 (see file [LICENSE](https://github.com/ulrichwisser/nsstat/blob/master/LICENSE))

All .zone files are derived from .SE zone data which
was aquired through [IIS Zonedata](https://zonedata.iis.se) and is
covered by the [Creative Commons Attribution 4.0 International](https://creativecommons.org/licenses/by/4.0/)  license.
