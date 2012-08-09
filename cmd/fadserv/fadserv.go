
package main

import (
  "fmt"
  "flag"
  "os"
  "log"
  "path/filepath"
  "github.com/rwcarlsen/cas/blobserv"
)

var defaultDB = filepath.Join(os.Getenv("HOME"), ".rcas")

var dbPath = flag.String("db", defaultDB, "path for the blob database to serve")
var addr = flag.String("addr", "0.0.0.0:7777", "address the server will listen on")

func main() {
  certFile := filepath.Join(*dbPath, "cert.pem")
  keyFile := filepath.Join(*dbPath, "key.pem")
  fmt.Println("running blob server...")
  log.Fatal(blobserv.ListenAndServeTLS(*addr, *dbPath, certFile, keyFile))
}

