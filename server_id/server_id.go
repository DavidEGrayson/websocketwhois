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

var serverMap map[string]*Server

func loadData() {
  responseAnalysisInit()

  var err error
  suffixes, err = data.SuffixesRead()
  if (err != nil) { log.Fatal(err) }

  debianSuffixInfos, err := data.DebianSuffixInfosRead()
  if (err != nil) { log.Fatal(err) }

  serverHints, err := data.ServerHintsRead()
  if (err != nil) { log.Fatal(err) }

  // Our first little bit of intelligent processing: If the hint says that
  // the server has a rate limit, that implies we should not use it.
  for _, server := range serverHints {
    if server.RateLimit {
      server.DoNotUse = true
    }
  }

  serverMap = groupByServer(debianSuffixInfos)

  for name, hint := range serverHints {
    if server, exists := serverMap[name]; exists {
      server.Hint = hint
    } else {
      log.Fatalf("Hint for server %s does not match any entry from Debian.  " +
        "Should we create a new Server object for it?", name)
    }
  }

}

func main() {
  log.SetFlags(0)  // don't put the date in the output
 
  loadData()

  singleServer := flag.String("s", "", "Only identify a single server.")
  skipToServer := flag.String("t", "", "Skip to a single server and start there (serial mode only).")
  parallelMode := flag.Bool("p", false, "Identify in parallel (much faster)")
  flag.Parse()

  if flag.NArg() != 0 {
    fmt.Fprintf(os.Stderr, "Valid arguments to %s are:\n", os.Args[0])
    flag.PrintDefaults()
    os.Exit(1)
  }

  if *singleServer != "" {

    var server *Server
    server, exists := serverMap[*singleServer]
    if !exists {
      log.Fatalf("Server %s not found.", *singleServer)
    }

    log.Printf("Only identifying a single server: %s.", *singleServer);
    server.identify()

  } else {

    servers := sortServers(serverMap)

    if *parallelMode {
      parallelIdentifyAll(servers)
    } else {
      serialIdentifyAll(servers, *skipToServer)
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
  r.NotExistRegexp = (*data.JsonRegexp)(s.NotExistRegexp)
  r.ExistRegexp = (*data.JsonRegexp)(s.ExistRegexp)

  // These fields we just copy directly from the hint; we didn't do anything to them here.
  if (s.Hint != nil) {
    r.RateLimit = s.Hint.RateLimit
    r.DoNotUse = s.Hint.DoNotUse
  }

  return &r
}

func serialIdentifyAll(servers []*Server, skipToServer string) {
  for i, server := range servers {
    if skipToServer != "" {
      if server.Name != skipToServer {
        continue
      }

      skipToServer = ""
    }

    success := server.identify()
    fmt.Println() // put space between servers in the log

    if !success {
      log.Fatalf("Aborting after first failure.  Server %d out of %d.",
        i+1, len(servers))
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