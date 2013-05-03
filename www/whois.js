$(whois = function() {
  var conn;
  var domains = {};
  var results = {"0":"doesNotExist", "1":"exists"}

  whois.resultCallbacks = $.Callbacks();

  whois.domains = domains;

  whois.Domain = function(name) {
    this.name = name
    this.state = null
    this.toString = function() {
      return this.name + ":" + this.state;
    }
  }

  whois.init = function() {
    conn = new WebSocket("ws://" + window.location.hostname + ":" + window.location.port + "/");
    whois.conn = conn;

    conn.onclose = function(evt) {
      console.log("Connection closed.");
    }

    conn.onmessage = function(evt) {
      console.log("Received message: " + evt.data);
      if (evt.data[0] == "r") {
        // A whois result was received.
        var parts = evt.data.substr(1).split(",");
        var domainName = parts[0];
        var state = results[parts[1]] || "error";

        var domain = domains[domainName];
        if (!domain) {
          // The server helpfully sent something we didn't request.
          domains[domainName] = domain = new whois.Domain(domainName)
        }
        domain.state = state;
        whois.resultCallbacks.fire(domain);
      }
    }

  }

  whois.domain = function(domainName) {
    var domain = domains[domainName];
 
    if (!domain) {
      //domains[domainName] = domain = { name: domainName, status: null };
      domains[domainName] = domain = new whois.Domain(domainName);

      console.log("Requesting " + domainName);
      conn.send("w" + domainName);
    }
    return domain;
  }

});
