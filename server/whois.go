package main

// PIR.org limits us:
// http://www.dnforum.com/f17/pir-limits-org-whois-thread-110375.html

// TODO: sign up for direct access to their zone files to get around it?
//   http://pir.org/help/access

// TODO: another option for zone files: http://www.premiumdrops.com/zones.html

// TODO: for taken domains, display if there is a GoDaddy.com auction

// Other whois sites:
// http://www.betterwhois.com/
// http://whois.domaintools.com/
// http://www.snapcheck.com/ ?


import (
  "io"
  "os"
  "os/exec"
  "bufio"
  "strings"
  "log"
  "errors"
)

var whoisConcurrencyLimiter chan bool

func whoisLimitAcquire() {
  <- whoisConcurrencyLimiter
}

func whoisLimitRelease() {
  whoisConcurrencyLimiter <- true
}

func whoisDomainExists(domain string) (bool, error) {

  wlog := *log.New(os.Stdout, "whois " + domain + " ", log.Flags())

  command := exec.Command("whois", "-H", domain)
  var err error

  outPipe, err := command.StdoutPipe()
  if err != nil {
    wlog.Println("Error in StdoutPipe():", err);
    return false, err
  }
  outBuf := bufio.NewReader(outPipe)

  command.Stderr = os.Stderr

  whoisLimitAcquire()
  defer whoisLimitRelease()

  err = command.Start()
  if err != nil {
    wlog.Println("Error in Cmd.start:", err);
    return false, err
  }
  wlog.Println("Started")

  noMatch := false
  match := false
  for {
    str, err := outBuf.ReadString('\n')
    if err != nil {
      if err == io.EOF {
        break
      }
      wlog.Println("Error reading line:", err);
      return false, err
    }
    
    if strings.HasPrefix(str, "No match for") {
      noMatch = true
    }

    if strings.HasPrefix(str, "   Domain Name:") {
      match = true
    }

    // This is what we get for pololu.org
    if str == "NOT FOUND\n" {
      noMatch = true
    }

    // This is what we got for foobarcrumbles.org
    if strings.HasPrefix(str, "No Match") {
      noMatch = true
    }

  }

  command.Wait();   // Clean up the defunct process.

  if (noMatch == match) {
    wlog.Println("Unrecognized result.");
    // TODO: log this confusing result from whois
    // TODO: don't lie to the user; need to return an err from this
    return false, errors.New("Unrecognized result.");
  }

  return match, nil;
}

func whoisInit() {
  whoisConcurrencyLimiter = make(chan bool, 100)
  for i := 0; i < cap(whoisConcurrencyLimiter); i++ {
    whoisConcurrencyLimiter <- true
  }
}