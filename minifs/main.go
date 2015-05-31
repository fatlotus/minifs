package main

import (
  "flag"
  "fmt"
  "strings"
  "net/http"
  "io/ioutil"
  "crypto/tls"
  "crypto/x509"
  "time"
  "stathat.com/c/consistent"
  "github.com/fatlotus/minifs"
)

var (
  port = flag.String("bind", ":8080", "what port to use for this server")
  cert = flag.String("cert", "", "what certificate to use")
  key = flag.String("key", "", "what private key to use")
  peers = flag.String("peers", "", "comma separated list of peers")
)

func main() {
  flag.Parse()
  
  if *port == "" || *cert == "" || *key == "" {
    fmt.Printf("Usage: csilfs -bind=:8080 -cert=server.crt -key=server.key\n")
    return
  }
  
  pm := make(map[string] int64)

  for _, peer := range strings.Split(*peers, ",") {
    if peer != "" {
      pm[peer] = 0
    }
  }
  
  cas, err := ioutil.ReadFile(*cert)
  if err != nil {
    panic(err)
  }
  
  pool := x509.NewCertPool()
  pool.AppendCertsFromPEM(cas)
  
  s := &minifs.Server{
    Prefix: "data",
    State: &minifs.ServerState{
      Peers: pm,
    },
    Port: *port,
    HTTPClient: &http.Client{
      Transport: &http.Transport{
    	  TLSClientConfig: &tls.Config{RootCAs: pool},
    	  DisableCompression: true,
      },
    },
    Cons: consistent.New(),
  }
  
  go func() {
    ticker := time.NewTicker(1 * time.Second)
    
    for range ticker.C {
      s.RunInspection()
    }
  }()
  
  err = http.ListenAndServeTLS(s.Port, *cert, *key, s)
  if err != nil {
    panic(err)
  }
}
