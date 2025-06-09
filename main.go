package main

import (
	"flag"
	"fmt"
	"net"
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

	fmt.Printf("%s:%d\n", hostAddr.String(), *portFlag)
}
