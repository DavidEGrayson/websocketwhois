package main

import (
  "net"
  "bufio"
  "strings"
)

type queryResult []string

func (r *queryResult) String() string {
  return strings.Join(queryResult, "\\n")
}

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
  conn, err := net.DialTimeout("tcp", s.Name + ":43", 40 * time.Second)
  if err != nil {
    s.log.Println("Error dialing", err)
    return nil, err
  }
  defer conn.Close()
  conn.SetDeadline(time.Now().Add(40 * time.Second))

  _, err = fmt.Fprint(conn, query + "\r\n")
  if err != nil {
    s.log.Println("Error sending", err)
    return nil, err
  }

  scanner := bufio.NewScanner(conn);
  result := queryResult([]string{})
  for scanner.Scan() {
    result = append(result, scanner.Text())
  }
  if scanner.Err() != nil {
    s.log.Println("Error scanning response: ", scanner.Err())
    return nil, scanner.Err()
  }
  
  return result, nil
}

