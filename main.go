package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/grqphical/cash/internal/server"
)

const version string = "1.0.0"

func main() {
	verFlag := flag.Bool("version", false, "Prints the version of cash")
	portFlag := flag.Int("port", 6400, "Port to run cash on")

	var hostAddr net.IP
	flag.TextVar(&hostAddr, "host", net.IPv4(0, 0, 0, 0), "IPv4 address to host cash on")

	flag.Parse()

	if *verFlag {
		fmt.Printf("cash version: %s\n", version)
		return
	}

	server, err := server.New(*portFlag, hostAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	fmt.Printf("starting cash on %s:%d\n", hostAddr, *portFlag)
	server.Listen()

}
