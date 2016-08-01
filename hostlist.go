package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

type Host struct {
	access   sync.Mutex
	Domains  []string
	IPs      []net.IP
	Glue     []net.IP
	isSEhost bool
}

var hostlist = make(map[string]*Host)
var hm = &sync.Mutex{}

func GetHost(hostname string) *Host {
	hm.Lock()
	defer hm.Unlock()
	host, ok := hostlist[hostname]
	if !ok {
		return nil
	}
	return host
}

func GetAllHostnames() []string {
	hosts := make([]string, 0)
	hm.Lock()
	defer hm.Unlock()
	for host := range hostlist {
		hosts = append(hosts, host)
	}
	return hosts
}

func AddHost(hostname string) {
	hm.Lock()
	defer hm.Unlock()
	if _, ok := hostlist[hostname]; !ok {
		hostlist[hostname] = &Host{Domains: nil, IPs: nil, Glue: nil, isSEhost: strings.HasSuffix(hostname, ".se.")}
	}
}

func AddDomain(hostname string, domain string) {
	hm.Lock()
	defer hm.Unlock()
	hostlist[hostname].access.Lock()
	hostlist[hostname].Domains = append(hostlist[hostname].Domains, domain)
	hostlist[hostname].access.Unlock()
}

func AddIP(hostname string, ip net.IP) {
	hm.Lock()
	defer hm.Unlock()
	hostlist[hostname].access.Lock()
	hostlist[hostname].IPs = append(hostlist[hostname].IPs, ip)
	hostlist[hostname].access.Unlock()
}

func AddGlue(hostname string, glue net.IP) {
	hm.Lock()
	defer hm.Unlock()
	hostlist[hostname].access.Lock()
	hostlist[hostname].Glue = append(hostlist[hostname].Glue, glue)
	hostlist[hostname].access.Unlock()
}

func Stats() {
	fmt.Println("--------------------------------")
	fmt.Println("Stats")
	fmt.Println("--------------------------------")

	var SEhosts = 0
	var SEhostsNoGlue = 0
	var SEhostsGlue = 0
	var SEhostsNoGlueNoIP = 0
	var SEhostsGlueNoIP = 0
	var SEhostsGlueNotInIP = 0

	var EXhosts = 0
	var EXhostsNoIP = 0

	for host := range hostlist {
		if hostlist[host].isSEhost {
			SEhosts++
			if len(hostlist[host].Glue) > 0 {
				// host has glue
				SEhostsGlue++
				if len(hostlist[host].IPs) == 0 {
					// hosts has no ips resolved
					SEhostsGlueNoIP++
				} else {
					// host has ips resolved
					GlueIpMissmatch := false

					// see if all glue records are found in resolved ip
					for _, glue := range hostlist[host].Glue {
						found := false
						for _, ip := range hostlist[host].IPs {
							if glue.Equal(ip) {
								found = true
							}
						}
						if !found {
							GlueIpMissmatch = true
						}
					}

					// see if all glue records are found in resolved ip
					for _, ip := range hostlist[host].IPs {
						found := false
						for _, glue := range hostlist[host].Glue {
							if ip.Equal(glue) {
								found = true
							}
						}
						if !found {
							GlueIpMissmatch = true
						}
					}

					// count missmatch
					if GlueIpMissmatch {
						SEhostsGlueNotInIP++
					}
				}
			} else {
				// host has no glue
				SEhostsNoGlue++
				if len(hostlist[host].IPs) == 0 {
					// host did not resolv, no ip
					SEhostsNoGlueNoIP++
				}
			}
		} else {
			// host out of zone
			EXhosts++
			if len(hostlist[host].IPs) == 0 {
				//host could not be resolved
				EXhostsNoIP++
			}
		}
	}
	fmt.Printf("SE Hosts no glue        / no ip :         %5d / %5d  (%5.1f)\n", SEhostsNoGlue, SEhostsNoGlueNoIP, (100.0 * float64(SEhostsNoGlueNoIP) / float64(SEhostsNoGlue)))
	fmt.Printf("SE Hosts    glue        / no ip :         %5d / %5d  (%5.1f)\n", SEhostsGlue, SEhostsGlueNoIP, (100.0 * float64(SEhostsGlueNoIP) / float64(SEhostsGlue)))
	fmt.Printf("SE Hosts    glue and ip / not matching :  %5d / %5d  (%5.1f)\n", (SEhostsGlue - SEhostsGlueNoIP), SEhostsGlueNotInIP, (100.0 * float64(SEhostsGlueNotInIP) / float64(SEhostsGlue-SEhostsGlueNoIP)))
	fmt.Printf("EX Hosts                / no ip :         %5d / %5d  (%5.1f)\n", EXhosts, EXhostsNoIP, (100.0 * float64(EXhostsNoIP) / float64(EXhosts)))
	fmt.Println("")
	fmt.Printf("Querries: %d  Average: %7.3f ms\n", queries, resolvtime*1000.0/float64(queries))

}
