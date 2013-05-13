package main

// TODO: Instead of using that sketchy tld_serv_list from the whois utility,
// get the root info from the IANA root zone file:
// http://www.iana.org/domains/root/files

import (
  "log"
  "strings"
  "fmt"
  "os"
  "io"
  "bufio"
  "net"
  "time"
  "sort"
  "regexp"
  "encoding/json"
  //"runtime"
  //"time"
)

// This represents a line from tld_serv_list, which came from the
// standard unix whois utility. 
type upstreamSuffixInfo struct {
  name, server, note string
}

type serverInfo struct {
  Name, note, Protocol string
  Suffixes []string
}

func (s *serverInfo) log(v ...interface{}) {
  v = append([]interface{}{ s.Name + ": " }, v...)
  fmt.Println(v...)  
}

type queryResult []string

func (r *queryResult) lastParagraphJoin() string {
  lines := *r
  paragraph := ""
  i := len(lines) - 1
  for ; lines[i] == ""; i -= 1 { }

  for ; i >= 0; i -= 1 {
    line := lines[i]
    if (line == "") {
      break
    }

    paragraph = line + " " + paragraph
  }

  return paragraph
}

func (r *queryResult) isOneLiner(line string) bool {
  lines := *r
  return len(lines) == 1 && lines[0] == line
}

// Opens a TCP connection to the remote server and sends a query.  The query consists
// of the provided string followed by "\r\n".  Reads data back from the server and
// returns it as a queryResult,  which is really just a slice of strings where each
// string is a line and the line-ending characters have been removed.
func (s *serverInfo) query(query string) (queryResult, error) {
  addr := s.Name + ":43"
  conn, err := net.DialTimeout("tcp", addr, 40 * time.Second)
  if err != nil {
    s.log("Error dialing", err)
    return nil, err
  }
  defer conn.Close()
  conn.SetDeadline(time.Now().Add(40 * time.Second))

  _, err = fmt.Fprint(conn, query + "\r\n")
  if err != nil {
    s.log("Error sending", err)
    return nil, err
  }

  reader := bufio.NewReader(conn)
  result := queryResult([]string {})
  for {
    str, err := reader.ReadString('\n')
    if err == io.EOF {
      break
    } else if err != nil {
      s.log("Error reading line:", err);
      return nil, err
    }
    str = strings.TrimRight(str, "\r\n")

    result = append(result, str)
  }

  return result, nil
}

// Whois Server Version 2.0
// This is a very important protocol and the servers that use will tell you what TLDs they have.
func (s *serverInfo) identifyWs20() {
  // Do an invalid query just so we can see the notes at the end of the
  // query.
  result, err := s.query("sum domain -")
  if err != nil {
    s.log("Failed to get Whois Server Version 2.0 help screen.");
    return
  }

  str := result.lastParagraphJoin()
  str = strings.ToLower(str);

  if strings.HasPrefix(str, "the registry database contains only") {
    re := regexp.MustCompile("\\.[\\.a-z]+")
    claimedSuffixes := re.FindAllString(str, -1)
    //s.log("Claims to support suffixes: " + strings.Join(claimedSuffixes, ", "))
    
    // TODO: print a warning message if we the followling line REMOVES any suffixes from s
    s.Suffixes = claimedSuffixes
  } else {
    s.log("The last paragraph did not talk about the registry's scope.  It was simply: " + str);
  }
}

func (s *serverInfo) detectAfilias() bool {
  result, err := s.query("help")
  if err != nil {
    s.log("Failed to get afilias help screen.");
    return false
  }

  str := strings.Join(result, " ");
  if !strings.Contains(str, "afilias") {
    s.log("Looked like an afilias server but did not return 'afilias' anywhere in the help screen.")
    return false
  }

  return true
}

func (s *serverInfo) identify() {
  // Can we get a help screen?
  questionMarkResult, err := s.query("?")
  if err != nil {
    s.log("Failed to get help.")
    return
  }
  resultJoined := strings.Join(questionMarkResult, " ")

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

  case questionMarkResult.isOneLiner("out of this registry"):
    s.Protocol = "ootr"

  case strings.HasPrefix(resultJoined, "Incorrect domain name: "):
    s.Protocol = "idn"

  case questionMarkResult.isOneLiner("Incorrect Query or request for domain not managed by this registry."):
    s.Protocol = "iqor"

  case len(questionMarkResult) > 20 && questionMarkResult[1] == "% This is ARNES whois database":
    s.Protocol = "arnes"

  case questionMarkResult[0] == "No entries found.":
    s.Protocol = "nef"

  case strings.HasPrefix(questionMarkResult[0], "% puntCAT Whois Server"):
    s.Protocol = "puntcat"
  }

  if (s.Protocol == "") {
    s.log("Failed to determine protocol.");
  } else {
    s.log("protocol=", s.Protocol, " suffixes=", s.Suffixes);
  }
}

func readUpstreamSuffixInfos(filename string) []upstreamSuffixInfo {
  upstreamSuffixInfos := make([]upstreamSuffixInfo, 0)

  file, err := os.Open("tld_serv_list")
  if err != nil {
	  log.Fatal("Error opening file.", err)
  }
  defer file.Close()
  reader := bufio.NewReader(file)

  for {
    line, err := reader.ReadString('\n');
    if (err == io.EOF) {
      break
    }
    if (err != nil) {
      log.Fatal("Error reading line.", err);
    }

    line = strings.Split(line, "#")[0]     // Remove comments
    fields := strings.Fields(line)         // Split by whitespace.
    
    var suffix upstreamSuffixInfo

    if len(fields) == 0 {
      continue   // Empty line.
    }

    suffix.name = fields[0]
    attrs := fields[1:]
    for _, attr := range attrs {
      if attr[0] >= 'A' && attr[0] <= 'Z' {
        suffix.note = attr
      } else {
        suffix.server = attr
      }
    }

    upstreamSuffixInfos = append(upstreamSuffixInfos, suffix)
  }

  return upstreamSuffixInfos
}

func groupByServer(suffixes []upstreamSuffixInfo) map[string] *serverInfo {
  servers := map[string] *serverInfo { }
  for _, suffix := range(suffixes) {

    if suffix.server == "" || suffix.note == "WEB" || suffix.note == "NONE" {
      // This entry does not have an actual whois server; ignore it.
      continue
    }

    // The upstream file has dashes instead of dots for some weird TLDs
    // at the bottom, like -tel.  I am not sure why.  Some of them
    // seem to be duplicate entries.  Skip them for now.
    // TODO: figure out what the - means.
    if strings.HasPrefix(suffix.name, "-") {
      continue
    }

    server := servers[suffix.server]

    if server == nil {
      server = &serverInfo{}
      servers[suffix.server] = server
      server.Name = suffix.server
      server.note = suffix.note
    }

    if server.note != suffix.note {
      log.Fatalf("Conflicting notes for %s: %s and %s.",
        server.Name, server.note, suffix.note)
    }
    
    server.Suffixes = append(server.Suffixes, suffix.name)
  }

  return servers
}

func removeUnusableServers(serverMap map[string] *serverInfo) {
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


// In Ruby this would just be serverMap.values.sort_by(&:name).
func sortServers(serverMap map[string] *serverInfo) []*serverInfo {
  serverNames := make([]string, 0)
  for name, _ := range serverMap {
    serverNames = append(serverNames, name)
  }
  sort.Strings(serverNames)
  
  serverSlice := make([]*serverInfo, len(serverNames))
  for i, name := range serverNames {
    serverSlice[i] = serverMap[name]
  }

  return serverSlice
}

func serialIdentifyAll(servers []*serverInfo) {
  for _, server := range servers {
    (*server).identify()
  }
}

func parallelIdentifyAll(servers []*serverInfo) {
  ch := make(chan bool)

  // Define the function we want to run in parallel.
  process := func(server *serverInfo) {
    (*server).identify()
    ch <- true
  }

  // Start it running in parallel (massively).
  for _, info := range servers { go process(info) }

  // Wait for all goroutines to finish.
  for _, _ = range servers { <- ch }
}

type OutputFile struct {
  Servers []*serverInfo
}

func writeOutput(servers []*serverInfo) {
  output := OutputFile { }

  output.Servers = servers

  file, err := os.Create("whois_database.json")
  if err != nil {
	  log.Fatal("Error opening output file.", err)
  }
  defer file.Close()

  //data, err := json.Marshal(output)  // For production.
  data, err := json.MarshalIndent(output, "", "  ")  // For development.
  if err != nil {
    log.Fatal("Error marshalling json data.", err)
  }

  n, err := file.Write(data)
  if err != nil || n != len(data) {
    log.Fatal("Error writing to output file.", err)
  }
}

func main() {
  upstreamSuffixInfos := readUpstreamSuffixInfos("tld_serv_list")
  
  //fmt.Println(upstreamSuffixInfos)

  serverMap := groupByServer(upstreamSuffixInfos)

  removeUnusableServers(serverMap)

  servers := sortServers(serverMap)

  // TODO: Since servers only has info about actual whois servers, we
  // should also pull out the information about TLDs that have no server
  // or only have a web interface, so we can show it to our users should
  // they request it.  That infor is in upstreamSuffixInfos.

  //serialIdentifyAll(sortServers(servers)) // For debugging.
  parallelIdentifyAll(servers);             // For production.

  writeOutput(servers)
}