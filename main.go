package main

import (
	"flag"
	"log"
)

func main() {
    var port, limit int
    flag.IntVar(&port, "port", 6666, "port number")
    flag.IntVar(&limit, "limit", 1000, "max client limit")
    flag.Parse()

    srv := NewGodisServer(port, limit)
	if err := srv.Run(); err != nil {
		log.Println("run server failed:", err)
	}
}

