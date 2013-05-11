package main

import (
  "testing"
)

func TestWhoisIntegration(test *testing.T) {
  whoisInit();

  var numTests int = 0
  var testCounter chan bool

  testExist := func(domainName string, shouldExist bool){
    exists, err := whoisDomainExists(domainName);
    if (err != nil) {
      test.Error("%s: %s", domainName, err)
    } else if (exists != shouldExist) {
      test.Errorf("%s: Expected %s existence, got %s.", domainName, shouldExist, exists)
    } else {
      test.Log(domainName, "success")
    }
    testCounter <- true
  }
  
  goTestExist := func(domainName string, shouldExist bool) {
    go testExist(domainName, shouldExist);
    numTests += 1;
  }

  goTestExist("st.com",           true);
  goTestExist("m7778aadhwQe.com", false);


  testCounter = make(chan bool)
  for i := 0; i < numTests; i++ {
    <-testCounter
  }
}