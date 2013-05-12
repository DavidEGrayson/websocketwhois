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

func (s *serverInfo) query(query string) {
  addr := s.name + ":43"
  log := *log.New(os.Stdout, fmt.Sprintf("%s: ", addr), log.Flags())
  conn, err := net.DialTimeout("tcp", addr, 40 * time.Second)
  if err != nil {
    log.Println("Error dialing", err)
    return
  }
  defer conn.Close()

  _, err = fmt.Fprint(conn, query + "\r\n")
  if err != nil {
    log.Println("Error sending", err)
    return
  }

}

func (s *serverInfo) identify() {
  log := *log.New(os.Stdout, fmt.Sprintf("%s: ", s.name), log.Flags())

  log.Print("Identifying.")

  // Can we get a help screen?
  s.query("?")

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

func serialIdentifyAll(servers map[string] *serverInfo) {
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
  serialIdentifyAll(servers)
  
  // TODO: sort results

  for _, server := range servers {
    fmt.Println(*server);
  }
}