package main

import (
	"fmt"
	"github.com/miekg/dns"
	"math/rand"
	"sync"
	"time"
)

const (
	TIMEOUT   time.Duration = 5 // seconds
	RATELIMIT uint          = 200
)

var ratelimiter = make(chan string, RATELIMIT)

var resolvers Strings = make(Strings, 0)
var resolvsync = &sync.Mutex{}

var resolvtime float64 = 0.0
var resolvtimem = sync.Mutex{}
var queries uint64 = 0

// translate rcode to human readable string
var rcode2string = map[int]string{
	0:  "Success",
	1:  "Format Error",
	2:  "Server Failure",
	3:  "Name Error",
	4:  "Not Implementd",
	5:  "Refused",
	6:  "YXDomain",
	7:  "YXRrset",
	8:  "NXRrset",
	9:  "Not Auth",
	10: "Not Zone",
	16: "Bad Signature / Bad Version",
	17: "Bad Key",
	18: "Bad Time",
	19: "Bad Mode",
	20: "Bad Name",
	21: "Bad Algorithm",
	22: "Bad Trunc",
	23: "Bad Cookie",
}

func GetIPs(hostname string, wg *sync.WaitGroup) {
	resolvsync.Lock()
	resolver := resolvers[rand.Intn(len(resolvers))]
	resolvsync.Unlock()
	wg.Add(2)
	ratelimiter <- "x"
	go resolv(hostname, dns.TypeA, resolver, wg)
	ratelimiter <- "x"
	go resolv(hostname, dns.TypeAAAA, resolver, wg)
}

// resolv will send a query and return the result
func resolv(qname string, qtype uint16, server string, wg *sync.WaitGroup) {
	start := time.Now()
	defer func() { _ = <-ratelimiter }()
	defer wg.Done()

	// Setting up query
	query := new(dns.Msg)
	query.RecursionDesired = true
	query.Question = make([]dns.Question, 1)
	query.SetQuestion(qname, qtype)

	// Setting up resolver
	client := new(dns.Client)
	client.ReadTimeout = TIMEOUT * 1e9

	// make the query and wait for answer
	r, _, err := client.Exchange(query, server)

	// check for errors
	if err != nil {
		//fmt.Printf("%-30s: Error resolving %s (server %s)\n", domain, err, server)
		return
	}
	if r == nil {
		//fmt.Printf("%-30s: No answer (Server %s)\n", domain, server)
		return
	}
	if r.Rcode != dns.RcodeSuccess {
		//fmt.Printf("%-30s: %s (Rcode %d, Server %s)\n", domain, rcode2string[r.Rcode], r.Rcode, server)
		return
	}

	for _, answer := range r.Answer {
		if answer.Header().Rrtype == dns.TypeA {
			AddIP(qname, answer.(*dns.A).A)
		}
		if answer.Header().Rrtype == dns.TypeAAAA {
			AddIP(qname, answer.(*dns.AAAA).AAAA)
		}
	}

	runtime := time.Now().Sub(start)
	if verbose {
		fmt.Printf("Resolving %s (qtype %d) from %s took %7.3fms\n", qname, qtype, server, runtime.Seconds()*1000.0)
	}
	resolvtimem.Lock()
	queries++
	resolvtime = resolvtime + runtime.Seconds()
	if verbose {
		fmt.Printf("Average: %7.3fms\n", resolvtime*1000.0/float64(queries))
	}
	resolvtimem.Unlock()
}
