package main

import (
  //"flag"
  "log"
  //"net/http"
  //"code.google.com/p/go.net/websocket"
  "strings"
  "fmt"
  "os"
  "io"
  "bufio"
  "net"
  "time"
  "sort"
  "regexp"
  //"runtime"
  //"time"
)

// This represents a line from tld_serv_list, which came from the
// standard unix whois utility. 
type upstreamSuffixInfo struct {
  name, server, note string
}

type serverInfo struct {
  name, note, protocol string
  suffixes []string
}

func (s *serverInfo) log(v ...interface{}) {
  fmt.Print(s.name + ": ")
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

// Opens a TCP connection to the remote server and sends a query.  The query consists
// of the provided string followed by "\r\n".  Reads data back from the server and
// returns it as a queryResult,  which is really just a slice of strings where each
// string is a line and the line-ending characters have been removed.
func (s *serverInfo) query(query string) (queryResult, error) {
  addr := s.name + ":43"
  conn, err := net.DialTimeout("tcp", addr, 40 * time.Second)
  if err != nil {
    s.log("Error dialing", err)
    return nil, err
  }
  defer conn.Close()

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
    s.log("Claims to support suffixes: " + strings.Join(claimedSuffixes, ", "))
    
    // TODO: print a warning message if we the followling line REMOVES any suffixes from s
    s.suffixes = claimedSuffixes
  } else {
    s.log("The last paragraph did not talk about the registry's scope.  It was simply: " + str);
  }
}

func (s *serverInfo) identify() {
  log := *log.New(os.Stdout, fmt.Sprintf("%s: ", s.name), log.Flags())

  s.log("Identifying.  Suffixes =", s.suffixes)

  // Can we get a help screen?
  questionMarkResult, err := s.query("?")
  if err != nil {
    s.log("Failed to get help.")
    return
  }

  switch {
  case len(questionMarkResult) > 20 && questionMarkResult[1] == "Whois Server Version 2.0":
    s.protocol = "ws20"
    s.identifyWs20()

  case len(questionMarkResult) == 1 && questionMarkResult[0] == "out of this registry":
    s.protocol = "ootr"
  }


  if (s.protocol == "") {
    s.log("Failed to determine protocol.");
  }
  return
}

//func parallelMap(input []upstreamSuffixInfo, f func(upstreamSuffixInfo) niceSuffixInfo) (output []niceSuffixInfo) {
//  output = make([]niceSuffixInfo, len(input))
//  ch := make(chan niceSuffixInfo)
//  process := func(info upstreamSuffixInfo) { ch <- f(info) }
//  for _, info := range input { go process(info) }
//  for i := range output { output[i] = <- ch }
//  return
//}

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
      server.name = suffix.server
      server.note = suffix.note
    }

    if server.note != suffix.note {
      log.Fatalf("Conflicting notes for %s: %s and %s.",
        server.name, server.note, suffix.note)
    }
    
    server.suffixes = append(server.suffixes, suffix.name)
  }

  return servers
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

func main() {
  upstreamSuffixInfos := readUpstreamSuffixInfos("tld_serv_list")
  
  //fmt.Println(upstreamSuffixInfos)

  servers := groupByServer(upstreamSuffixInfos)

  // TODO: Since servers only has info about actual whois servers, we
  // should also pull out the information about TLDs that have no server
  // or only have a web interface, so we can show it to our users should
  // they request it.  That infor is in upstreamSuffixInfos.

  //niceSuffixInfos := parallelMap(upstreamSuffixInfos, identifyServer)
  serialIdentifyAll(sortServers(servers))
  
  // TODO: sort results

  for _, server := range servers {
    fmt.Println(*server);
  }
}