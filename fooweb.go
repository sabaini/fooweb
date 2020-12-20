// fooweb
// Example of a REST service connecting to a backend process via the NATS messaging system

package main

import (
	"log"
	"os"
	"sabaini.at/fooweb/node"
	"sabaini.at/fooweb/web"
)

func main() {
	// Role backend: run backend node
	if len(os.Args) > 1 && os.Args[1] == "backend" {
		node.SetupBackend()
		return
	}
	// Role web: run webserver
	if len(os.Args) > 1 && os.Args[1] == "web" {
		r := web.Setup()
		r.Run()
		return
	}

	log.Fatalln("Usage: fooweb web|backend")
}
