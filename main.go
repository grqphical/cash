package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"os"

	"github.com/grqphical/cash/internal/server"
)

const version string = "1.0.0"

func main() {
	verFlag := flag.Bool("version", false, "Prints the version of cash")
	portFlag := flag.Int("port", 6400, "Port to run cash on")
	persistenceFileName := flag.String("file", "cache.cashlog", "File to persist data to")

	var hostAddr net.IP
	flag.TextVar(&hostAddr, "host", net.IPv4(0, 0, 0, 0), "IPv4 address to host cash on")

	flag.Parse()

	if *verFlag {
		fmt.Printf("cash version: %s\n", version)
		return
	}

	logFile, err := os.OpenFile("log.json", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		slog.Error("could not open log file", "error", err)
		return
	}
	defer logFile.Close()

	logWriter := io.MultiWriter(os.Stderr)
	logger := slog.New(slog.NewJSONHandler(logWriter, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	slog.SetDefault(logger)

	server, err := server.New(*portFlag, hostAddr, *persistenceFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	slog.Info("starting cash", "host", hostAddr.String(), "port", *portFlag)
	server.Listen()

}
