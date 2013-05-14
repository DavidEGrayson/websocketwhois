package data

import (
  "encoding/json"
  "os"
  "regexp"
  "log"
)

type Server struct {
  Name, Protocol string
  Suffixes []string
  NotExistRegexp, ExistRegexp *regexp.Regexp
}

type OutputFile struct {
  Servers []*Server
}

func ServersWrite(servers []*Server) {
  output := OutputFile { }

  output.Servers = servers

  file, err := os.Create(Directory + "/servers.json")
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


