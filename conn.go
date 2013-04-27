package main

// TODO: configure emacs to not put tab characters in go code

import (
  "code.google.com/p/go.net/websocket"
  //"os/exec"
  "log"
  "fmt"
  "net"
  "io"
  //"bufio"
  "os"
  //"net/http"
  "reflect"
  "syscall"
  "strings"
)

// The connection between a websocket and a process.
type connection struct {
  ws *websocket.Conn
  Id int
  log log.Logger

  closeRequest chan bool
  whoisRequests chan string
  Closed bool
}

var idChannel chan int = make(chan int)

// Called once when the program starts.
func connInit() {
  go assignIds()
}

// Adds 0, 1, 2, 3, ... to the idChannel.
func assignIds() {
  i := 0
  for {
    idChannel <- i
    i++
  }
  // TODO: prevent integer overflow at some point?
}

func websocketClosedLocally(err error) bool {
  if operror, ok := err.(*net.OpError); ok {
    if errno, ok := operror.Err.(syscall.Errno); ok {
      if (errno==1236) { // Microsoft Windows ERROR_CONNECTION_ABORTED
        return true
      }
    }
  }
  return false
}

func (c *connection) receive() {
  defer c.Close()
  for {
    var message string
    err := websocket.Message.Receive(c.ws, &message)
    if err != nil {
      switch {
        case c.Closed: // Suppress error message
        case websocketClosedLocally(err): c.log.Println("Websocket closed locally.")
        case err == io.EOF: c.log.Println("Websocket closed remotely.")
        default: c.log.Println("Error reading from websocket:", err, reflect.TypeOf(err))
      }
      return
    }

    requestType, requestData := message[0:1], strings.ToLower(message[1:])
		switch requestType {
		case "w": // Whois query for one specific domain name.
      c.whoisRequests <- requestData
    case "g": // g: General query asking "what can I do to get a name like this"
      // These are the TLDs that it is OK to add.
      // TODO: make this list come from the request
			tlds := []string{"com", "net"}
			for _, tld := range tlds {
				c.whoisRequests <- requestData + "." + tld
			}

      // TODO: if the fragment happens to end in a TLD, suggest that
    }
  }
}

func (c *connection) work() {
  for domain := range c.whoisRequests {
    exists, err := whoisDomainExists(domain)

    c.log.Println("Result: ", domain, exists, err);
		var resultString string
		if err != nil {
			resultString = "e"  // There was an error.
		} else if exists {
			resultString = "1"  // The domain exists; it is taken.
		} else {
			resultString = "0"  // The domain does not exist.
		}

		str := "r" + domain + "," + resultString
		//log.Println("Sending to websocket:", str)
		err = websocket.Message.Send(c.ws, str)
		if err != nil && !c.Closed {
			log.Println("Error sending to websocket:", err)
		}    
  }
}

func (c *connection) run() {
  c.Id = <-idChannel
  c.log = *log.New(os.Stdout, fmt.Sprintf("#%d ", c.Id), log.Flags())
  c.log.Println("New connection from " + c.ws.Request().RemoteAddr + ".")

  // closeRequest must be buffered in case multiple goroutines
  // call Close() at nearly the same time.  Is there a better way?
  c.closeRequest = make(chan bool, 10)

  c.whoisRequests = make(chan string, 20)

  go c.receive()
  for i := 0; i < 20; i++ {
    go c.work()
  }

  <-c.closeRequest
  c.cleanup()
}

func (c *connection) Close() {
  c.log.Println("Sending request to close.")
  c.closeRequest <- true
}

func (c *connection) cleanup() {
  c.Closed = true
  c.ws.Close()
	close(c.whoisRequests)
  c.log.Println("Connection cleaned up.")
}

func wsHandler(ws *websocket.Conn) {
  c := &connection { ws: ws }
  c.run()
}
