package main

import (
  "strings"
  "bufio"
  "os"
  "log"
  "fmt"
  "io"
  "sort"
  "encoding/json"
)

// This represents a line from tld_serv_list, which came from the
// standard unix whois utility. 
type upstreamSuffixInfo struct {
  name, server, note string
}


func main() {
  log.SetFlags(0)  // don't put the date in the output

  initData()

  upstreamSuffixInfos := readUpstreamSuffixInfos("tld_serv_list")
  
  //fmt.Println(upstreamSuffixInfos)

  serverMap := groupByServer(upstreamSuffixInfos)

  removeUnusableServers(serverMap)

  servers := sortServers(serverMap)

  // TODO: Since servers only has info about actual whois servers, we
  // should also pull out the information about TLDs that have no server
  // or only have a web interface, so we can show it to our users should
  // they request it.  That infor is in upstreamSuffixInfos.

  serialIdentifyAll(servers) // For debugging.
  //parallelIdentifyAll(servers);             // For production.

  writeOutput(servers)
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
      server.log = log.New(os.Stdout, fmt.Sprintf("%s: ", server.Name), log.Flags())
    }

    if server.note != suffix.note {
      log.Fatalf("Conflicting notes for %s: %s and %s.",
        server.Name, server.note, suffix.note)
    }
    
    server.Suffixes = append(server.Suffixes, suffix.name)
  }

  return servers
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
    server.identify()
    fmt.Println() // put space between servers in the log

    if server.Protocol == "" {
      log.Fatal("Aborting after first failure.")
    }
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

