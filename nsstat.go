package main

import (
	"flag"
	"fmt"
	"github.com/miekg/dns"
	"io"
	"os"
	"sync"
)

var verbose bool = false

func main() {

	// define and parse command line arguments
	flag.BoolVar(&verbose, "verbose", false, "print more information while running")
	flag.BoolVar(&verbose, "v", false, "print more information while running")
	flag.Var(&resolvers, "resolver", "give ip addresses to resolvers")
	flag.Var(&resolvers, "r", "give ip addresses to resolvers")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Printf("Usage: %s [-v] <filename>\n", os.Args[0])
		os.Exit(1)
	}

	// set the list of resolvers to use
	GetResolvers()

	f, err := os.Open(flag.Arg(0))
	if err != nil {
		panic(err)
	}
	nsstat(f)
	f.Close()

	// print statistics
	Stats()
}

func nsstat(zonefile io.Reader) {

	rrlist := filter_ns(dns.ParseZone(zonefile, "", ""))

	// syncing of ip resolvers
	var wg sync.WaitGroup

	for rr := range rrlist {
		switch rr.(type) {
		case *dns.NS:
			hostname := rr.(*dns.NS).Ns
			domain := rr.Header().Name
			if GetHost(hostname) == nil {
				AddHost(hostname)
				GetIPs(hostname, &wg)
			}
			AddDomain(hostname, domain)
		case *dns.A:
			hostname := rr.Header().Name
			if GetHost(hostname) == nil {
				AddHost(hostname)
				GetIPs(hostname, &wg)
			}
			glue := rr.(*dns.A).A
			AddGlue(hostname, glue)
		case *dns.AAAA:
			hostname := rr.Header().Name
			if GetHost(hostname) == nil {
				AddHost(hostname)
				GetIPs(hostname, &wg)
			}
			glue := rr.(*dns.AAAA).AAAA
			AddGlue(hostname, glue)
		}
	}

	// wait for all hostnames to be resolved
	wg.Wait()
}

func filter_ns(tokens <-chan *dns.Token) <-chan dns.RR {

	out := make(chan dns.RR)
	go func() {
		for token := range tokens {
			if token.Error != nil {
				fmt.Println("Error: ", token.Error)
				os.Exit(1)
			}
			if token.RR.Header().Rrtype == dns.TypeNS {
				out <- token.RR
			}
			if token.RR.Header().Rrtype == dns.TypeA {
				out <- token.RR
			}
			if token.RR.Header().Rrtype == dns.TypeAAAA {
				out <- token.RR
			}
		}
		close(out)
	}()
	return out
}
