package main

import (
	"fmt"
	"github.com/miekg/dns"
	"os"
	"strings"
)

func ip2resolver(server string) string {
	if strings.ContainsAny(":", server) {
		// IPv6 address
		server = "[" + server + "]:53"
	} else {
		server = server + ":53"
	}
	return server
}

// getResolvers will read the list of resolvers from /etc/resolv.conf
func GetResolvers() []string {
	//resolvers = append(resolvers, "172.21.36.10:53")
	if len(resolvers) == 0 {
		conf, err := dns.ClientConfigFromFile("/etc/resolv.conf")
		if conf == nil {
			fmt.Printf("Cannot initialize the local resolver: %s\n", err)
			os.Exit(1)
		}
		for i := range conf.Servers {
			resolvers = append(resolvers, ip2resolver(conf.Servers[i]))
		}
		if len(resolvers) == 0 {
			fmt.Println("No resolvers found.")
			os.Exit(5)
		}
	}
	if verbose {
		for _, server := range resolvers {
			fmt.Println("Found resolver " + server)
		}
	}
	return resolvers
}
