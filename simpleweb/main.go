package main

import (
	"flag"
	"github.com/go-martini/martini"
	"log"
)

func main() {
	sdir := flag.String("d", ".", "static resource dir to serve")
	port := flag.String("p", "8000", "port to listen")
	flag.Parse()
	log.Printf("Serving %s on port %s", *sdir, *port)
	m := martini.Classic()
	m.Use(martini.Static(*sdir))
	m.RunOnAddr(":" + *port)
}
