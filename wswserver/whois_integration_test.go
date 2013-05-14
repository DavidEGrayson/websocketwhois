package main

import (
  "testing"
  "fmt"
)

func TestWhoisIntegration(test *testing.T) {
  whoisInit();

  var numTests int = 0
  var testCounter chan bool

  testExist := func(domainName string, shouldExist bool){
    exists, err := whoisDomainExists(domainName);
    if (err != nil) {
      test.Errorf("%s: %s", domainName, err)
    } else if (exists != shouldExist) {
      test.Errorf("%s: Expected %s existence, got %s.", domainName, shouldExist, exists)
    } else {
      test.Log(domainName, "success")
      fmt.Println("Success");
    }
    testCounter <- true
  }
  
  goTestExist := func(domainName string, shouldExist bool) {
    go testExist(domainName, shouldExist);
    numTests += 1;
  }

  goTestExist("st.com",           true);
  goTestExist("m7778aadhwQe.com", false);
  goTestExist("golang.org",       true);
  goTestExist("go489999213z.org", false);

  testCounter = make(chan bool)
  for i := 0; i < numTests; i++ {
    <-testCounter
  }
}