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
  suffix, server, note string
}

type niceSuffixInfo struct {
  suffix, server, note, protocol string
}

type server string

func (s *server) query(query string) {
  addr := string(*s) + ":43"
  l := *log.New(os.Stdout, fmt.Sprintf("%s: ", addr), log.Flags())
  l.Println("Dialing")
  conn, err := net.DialTimeout("tcp", addr, 40 * time.Second)
  if err != nil {
    l.Println("Error dialing", err)
    return
  }
  defer conn.Close()

  _, err = fmt.Fprint(conn, query + "\r\n")
  if err != nil {
    l.Println("Error sending", err)
    return
  }

  
}

func identifyServer(info upstreamSuffixInfo) (output niceSuffixInfo) {
  output.suffix = info.suffix
  output.server = info.server
  output.note = info.note

  var s server = server(info.server)

  if output.note == "WEB" {
    // This one is only accessible from the web.
    return
  }

  if output.note == "NONE" || output.server == "" {
    return
  }

  s.query("?")

  return
}

func parallelMap(input []upstreamSuffixInfo, f func(upstreamSuffixInfo) niceSuffixInfo) (output []niceSuffixInfo) {
  output = make([]niceSuffixInfo, len(input))
  ch := make(chan niceSuffixInfo)
  process := func(info upstreamSuffixInfo) { ch <- f(info) }
  for _, info := range input { go process(info) }
  for i := range output { output[i] = <- ch }
  return
}

// For easier debugging...
func normalMap(input []upstreamSuffixInfo, f func(upstreamSuffixInfo) niceSuffixInfo) (output []niceSuffixInfo) {
  output = make([]niceSuffixInfo, len(input))
  for i, info := range input { output[i] = f(info) }
  return
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
    
    var info upstreamSuffixInfo

    if len(fields) == 0 {
      continue   // Empty line.
    }

    info.suffix = fields[0]
    attrs := fields[1:]
    for _, attr := range attrs {
      if attr[0] >= 'A' && attr[0] <= 'Z' {
        info.note = attr
      } else {
        info.server = attr
      }
    }

    upstreamSuffixInfos = append(upstreamSuffixInfos, info)
  }

  return upstreamSuffixInfos
}

func main() {
  fmt.Println("Identifying the servers...")

  upstreamSuffixInfos := readUpstreamSuffixInfos("tld_serv_list")

  niceSuffixInfos := parallelMap(upstreamSuffixInfos, identifyServer)
  //niceSuffixInfos := normalMap(upstreamSuffixInfos, identifyServer)
  
  // TODO: sort results

  fmt.Println(niceSuffixInfos);
}