package data

import (
  "encoding/json"
  "os"
  "regexp"
  "log"
  "../errorwrap"
  "io/ioutil"
)

type Server struct {
  Name, Protocol string
  Suffixes []string
  NotExistRegexp, ExistRegexp *regexp.Regexp
}

type OutputFile struct {
  Servers []*Server
}

func ServerHintsRead() (map[string]*Server, error) {
  servers := make(map[string]*Server)
  
  hint_bytes, err := ioutil.ReadFile(Directory + "/server-hints.json")
  if err != nil { return nil, err }

  err = json.Unmarshal(hint_bytes, &servers)
  if err != nil { return nil, errorwrap.New("Error decoding server-hints.json", err) }

  return servers, nil
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


