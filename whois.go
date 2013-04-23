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

type whoisRequest struct {
  Domain string
  ResultChannel chan whoisResult  
}

type whoisResult struct {
  WhoisRequest whoisRequest
  Exists bool
}

var whoisRequestChannel chan whoisRequest = make(chan whoisRequest, 2000)

func whoisHandleRequests() {
  for request := range whoisRequestChannel {
		command := exec.Command("whois", "-H", request.Domain)
		var err error

		outPipe, err := command.StdoutPipe()
		if err != nil {
			log.Fatal("Error in StdoutPipe():", err);
			return
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

		if (noMatch != match) {
      var result whoisResult
      result.WhoisRequest = request
      result.Exists = match
      request.ResultChannel <- result
		} else {
			log.Println("Unrecognized result from whois.");
			// TODO: log this confusing result from whois
		}
  }
}

func whoisInit() {
  for i := 0; i < 100; i++ {
    go whoisHandleRequests()
  }
}