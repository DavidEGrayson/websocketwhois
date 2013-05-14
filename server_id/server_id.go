package main

// TODO: Since servers only has info about actual whois servers, we
// should also pull out the information about TLDs that have no server
// or only have a web interface, so we can show it to our users should
// they request it.  That info is in upstreamSuffixInfos.

import (
  "strings"
  "os"
  "log"
  "fmt"
  "sort"
  "../data"
  "flag"
)

var servers []*Server

func loadData() {
  responseAnalysisInit()

  var err error
  suffixes, err = data.SuffixesRead()
  if (err != nil) { log.Fatal(err) }

  debianSuffixInfos, err := data.DebianSuffixInfosRead()
  if (err != nil) { log.Fatal(err) }

  serverMap := groupByServer(debianSuffixInfos)

  removeUnusableServers(serverMap)

  servers = sortServers(serverMap)
}

func main() {
  log.SetFlags(0)  // don't put the date in the output
 
  loadData()

  singleServer := flag.String("s", "", "Only identify a single server.")
  parallelMode := flag.Bool("p", false, "Identify in parallel (much faster)")
  flag.Parse()

  if flag.NArg() != 0 {
    fmt.Fprintf(os.Stderr, "Valid arguments to %s are:\n", os.Args[0])
    flag.PrintDefaults()
    os.Exit(1)
  }

  if *singleServer != "" {

    var server *Server
    for _, s := range servers {
      if s.Name == *singleServer {
        server = s
      }
    }

    if server == nil {
      log.Fatalf("Server %s not found." , singleServer)
    }

    log.Printf("Only identifying a single server: %s.", *singleServer);
    server.identify()

  } else {

    if *parallelMode {
      parallelIdentifyAll(servers)
    } else {
      serialIdentifyAll(servers)
    }

    output := extractOutput(servers)
    data.ServersWrite(output)
  }
}

func groupByServer(suffixes []data.DebianSuffixInfo) map[string] *Server {
  servers := map[string] *Server { }
  for _, suffix := range(suffixes) {

    if suffix.Server == "" || suffix.Note == "WEB" || suffix.Note == "NONE" {
      // This entry does not have an actual whois server; ignore it.
      continue
    }

    // The upstream file has dashes instead of dots for some weird TLDs
    // at the bottom, like -tel.  I am not sure why.  Some of them
    // seem to be duplicate entries.  Skip them for now.
    // TODO: figure out what the - means.
    if strings.HasPrefix(suffix.Name, "-") {
      continue
    }

    server := servers[suffix.Server]

    if server == nil {
      server = &Server{}
      servers[suffix.Server] = server
      server.Name = suffix.Server
      server.DebianNote = suffix.Note
      server.log = log.New(os.Stdout, fmt.Sprintf("%s: ", server.Name), log.Flags())
    }

    if server.DebianNote != suffix.Note {
      log.Fatalf("Conflicting notes for %s: %s and %s.",
        server.Name, server.DebianNote, suffix.Note)
    }
    
    server.Suffixes = append(server.Suffixes, suffix.Name)
  }

  return servers
}


// In Ruby this would just be serverMap.values.sort_by(&:name).
func sortServers(serverMap map[string] *Server) []*Server {
  serverNames := make([]string, 0)
  for name, _ := range serverMap {
    serverNames = append(serverNames, name)
  }
  sort.Strings(serverNames)
  
  serverSlice := make([]*Server, len(serverNames))
  for i, name := range serverNames {
    serverSlice[i] = serverMap[name]
  }

  return serverSlice
}

func extractOutput(servers []*Server) []*data.Server {
  r := make([]*data.Server, len(servers))
  for i, server := range servers {
    r[i] = server.extractOutput()
  }
  return r
}

func (s *Server) extractOutput() *data.Server {
  r := data.Server{}
  r.Name = s.Name
  r.Suffixes = s.Suffixes
  r.Protocol = s.Protocol
  r.NotExistRegexp = s.NotExistRegexp
  r.ExistRegexp = s.ExistRegexp
  return &r
}

func serialIdentifyAll(servers []*Server) {
  for _, server := range servers {
    server.identify()
    fmt.Println() // put space between servers in the log

    if server.Protocol == "" {
      log.Fatal("Aborting after first failure.")
    }
  }
}

func parallelIdentifyAll(servers []*Server) {
  ch := make(chan bool)

  // Define the function we want to run in parallel.
  process := func(server *Server) {
    (*server).identify()
    ch <- true
  }

  // Start it running in parallel (massively).
  for _, info := range servers { go process(info) }

  // Wait for all goroutines to finish.
  for _, _ = range servers { <- ch }
}