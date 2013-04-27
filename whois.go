package main

// TODO: better handling of concurrency.  We don't want our 100 goroutines
// handling the whois requests to be slowed down by having to wait around
// for the websockets to accept their data, and that could be a way that
// someone brings down the system

import (
  "io"
  "os"
	"os/exec"
  "bufio"
  "strings"
  "log"
)

var whoisConcurrencyLimiter chan bool

func whoisLimitAcquire() {
  <- whoisConcurrencyLimiter
}

func whoisLimitRelease() {
  whoisConcurrencyLimiter <- true
}

func whoisDomainExists(domain string) bool {
  whoisLimitAcquire()
  defer whoisLimitRelease()

	command := exec.Command("whois", "-H", domain)
	var err error

	outPipe, err := command.StdoutPipe()
	if err != nil {
		log.Fatal("Error in StdoutPipe():", err);
		return false   // TODO: return an error
	}
	outBuf := bufio.NewReader(outPipe)

	command.Stderr = os.Stderr

	err = command.Start()
	if err != nil {
		log.Fatal("Error in Cmd.start:", err);
	}

	noMatch := false
	match := false
	for {
		str, err := outBuf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal("Error reading line:", err);
		}
		
		if strings.HasPrefix(str, "No match for") {
			noMatch = true
		}

		if strings.HasPrefix(str, "   Domain Name:") {
			match = true
		}

	}

	if (noMatch == match) {
		log.Println("Unrecognized result from whois.");
		// TODO: log this confusing result from whois
    // TODO: don't lie to the user; need to return an err from this
    return false;
	}  

  return match;
}

func whoisInit() {
  whoisConcurrencyLimiter = make(chan bool, 2000)
  for i := 0; i < cap(whoisConcurrencyLimiter); i++ {
    whoisConcurrencyLimiter <- true
  }
}