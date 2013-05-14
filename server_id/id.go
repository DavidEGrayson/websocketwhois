package main

// TODO: check all uses of io.EOF in this whole project and make sure they are OK.
//   Often a Read function will return io.EOF along with some data!

// TODO: Instead of using that sketchy tld_serv_list from the whois utility,
// get the root info from the IANA root zone file:
// http://www.iana.org/domains/root/files

import (
  "strings"
  "regexp"
  "math/rand"
  "../data"
  "log"
)

type Server struct {
  // Identity
  Name string
  Suffixes []string

  Hint *data.Server   // Hints from the maintainers of this program.
  DebianNote string   // Info from the Debian whois utility.

  // The fields we want to compute.
  Protocol string
  NotExistRegexp, ExistRegexp *regexp.Regexp

  log *log.Logger
}

var suffixes map[string] data.Suffix;


// Whois Server Version 2.0
// This is a very important protocol and the servers that use will tell you what TLDs they have.
func (s *Server) identifyWs20() {
  // Do an invalid query just so we can see the notes at the end of the
  // query.
  result, err := s.query("sum domain -")
  if err != nil {
    s.log.Println("Failed to get Whois Server Version 2.0 help screen.");
    return
  }

  str := result.lastParagraphJoin()
  str = strings.ToLower(str)

  if strings.HasPrefix(str, "the registry database contains only") {
    re := regexp.MustCompile("\\.[\\.a-z]+")
    claimedSuffixes := re.FindAllString(str, -1)
    //s.log("Claims to support suffixes: " + strings.Join(claimedSuffixes, ", "))
    
    // TODO: print a warning message if we the followling line REMOVES any suffixes from s
    s.Suffixes = claimedSuffixes
  } else {
    s.log.Println("The last paragraph did not talk about the registry's scope.  It was simply: " + str);
  }
}

func (s *Server) detectAfilias() (success bool) {
  result, err := s.query("help")
  if err != nil {
    s.log.Println("Failed to get afilias help screen.");
    return false
  }

  str := strings.Join(result, " ");
  if !strings.Contains(str, "afilias") {
    s.log.Println("Looked like an afilias server but did not return 'afilias' anywhere in the help screen.")
    return false
  }

  return true
}

var bytelist = []byte {
  'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
  'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
  // Do NOT include '-' in this list because we are tryin to generate a domain
  // name that probably does not exist but would be valid, and a hyphen
  // in certain spots is not allowed.
}

func randomDomain(suffix string) string {
  length := 50
  str := ""
  for i := 0; i < length; i++ {
    str += string( bytelist[ rand.Intn(len(bytelist)) ] )
  }
  return str + suffix
}

// Returns a domain name that is likely to exist.
func likelyDomain(suffix string) string {
  if suffixData, ok := suffixes[suffix]; ok {
    if ex := suffixData.ExampleExistingDomain; ex != "" {
      return ex
    }
  }

  return "aa" + suffix
}

func patternsMatchCounts(strings []string, patterns []*regexp.Regexp) map[*regexp.Regexp] int {
  patternMap := make(map[*regexp.Regexp] int)
  for _, line := range strings {
    for _, pattern := range patterns {
      if pattern.MatchString(line) {
        count, ok := patternMap[pattern]
        if !ok { count = 0 }
        count += 1
        patternMap[pattern] = count
      }
    }
  }
  return patternMap
}

func (s *Server) identifyGenericProtocol() (err error) {
  
  suffix := s.Suffixes[0]

  domainNameProbablyNotExist := randomDomain(suffix)
  s.log.Println("Asking about " + domainNameProbablyNotExist)
  queryResult, err := s.query(domainNameProbablyNotExist)
  if err != nil { return err }

  s.NotExistRegexp, err = analyzeNotExistResponse(queryResult)
  if (err != nil) { return err }
  s.log.Println("Not-exist response matches ", s.NotExistRegexp)

  domainNameProbablyExist := likelyDomain(suffix)
  queryResult, err = s.query(domainNameProbablyExist)
  if err != nil { return err }

  s.ExistRegexp, err = analyzeExistResponse(queryResult, domainNameProbablyExist)
  if (err != nil) { return err }
  s.log.Println("Exist response matches ", s.ExistRegexp)
  
  s.Protocol = "generic"


  return nil
}

func (s *Server) identify() {

  if (s.Hint != nil && s.Hint.Protocol != "") {

    if s.Hint.Protocol == "generic" && s.Hint.NotExistRegexp == nil || s.Hint.ExistRegexp == nil {
      s.log.Printf("Error: Hint says protocol=generic but did not specify regexps.")
      return
    }

    s.Protocol = s.Hint.Protocol
    s.NotExistRegexp = (*regexp.Regexp)(s.Hint.NotExistRegexp)
    s.ExistRegexp = (*regexp.Regexp)(s.Hint.ExistRegexp)    
    s.log.Printf("Protocol (from hint) is %s, %s, %s",
      s.Protocol, s.NotExistRegexp, s.ExistRegexp)
    return
  }

  // Can we get a help screen?
  questionMarkResult, err := s.query("?")
  if err != nil {
    s.log.Println("Error with question mark query.")
    return
  }
  //resultJoined := strings.Join(questionMarkResult, " ")

  switch {

  case len(questionMarkResult) == 0:

  case len(questionMarkResult) > 20 && questionMarkResult[1] == "Whois Server Version 2.0":
    s.Protocol = "ws20"
    s.identifyWs20()

  case questionMarkResult.isOneLiner("Not a valid domain search pattern"):
    if (s.detectAfilias()) {
      s.Protocol = "afilias"
    }

  case strings.HasPrefix(questionMarkResult[0], "swhoisd"):
    s.Protocol = "swhoisd"

  default:
    err = s.identifyGenericProtocol()
  }

  if (err != nil) {
    s.log.Println(err)
  }

  if (s.Protocol == "") {
    s.log.Println("Failed to identify protocol.")
  } else {
    s.log.Println("Protocol is", s.Protocol)
  }
}


// TODO: store this info as a JSON file instead, like the
//   "human input" JSON file.
func removeUnusableServers(serverMap map[string] *Server) {
  // TODO: get zone file access for all these weird servers, or at least the important ones.

  weirdServers := []string {
    // These servers have a pretty extreme rate limit so we do not plan on contacting
    // them in production.  We do not need to identify their protocol.
    "whois.pir.org",            // .org
    "kero.yachay.pe",           // .ae
    "whois.adamsnames.tc",      // .gd .tc, .vg, 
    "whois.aeda.net.ae",        // .ae
    "whois.ausregistry.net.au", // .au

    // I tried but could not figure out how to get a meaningul response from these:
    "whois.ac.za",              // .ac.za
    // TODO: tell our users that the entire .ac.za list is here: http://protea.tenet.ac.za/cgi/cgi_domainquery.exe?list

    // These TLDs are not available even though whois might work.
    "whois.alt.za",
    // TODO: tell people that .alt.za allows no new registrations according to http://www.internet.org.za/slds.html
  }

  for _, serverName := range weirdServers {
    delete(serverMap, serverName)
  }
}



