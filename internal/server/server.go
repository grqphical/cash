package server

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/grqphical/cash/internal/cache"
)

type Server struct {
	cmdChan    chan cache.Command
	outputChan chan string
	errChan    chan cache.DBError
	listener   net.Listener
}

func New(port int, hostAddr net.IP) (*Server, error) {
	cache := cache.New()
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", hostAddr, port))
	if err != nil {
		return nil, err
	}

	cmdChan, outputChan, errChan := cache.Run()

	return &Server{
		cmdChan,
		outputChan,
		errChan,
		listener,
	}, nil
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)

	for {

		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				fmt.Println("client closed connection")
				return
			}
			fmt.Printf("error while reading packet: %s\n", err)
			continue
		}
		commandStr := string(buffer[:n])

		command, err := cache.ParseCommandFromString(commandStr)
		if err != nil {
			fmt.Printf("error while parsing packet: %s\n", err)
			continue
		}

		s.cmdChan <- command

		dbErr := <-s.errChan
		if dbErr.Kind() != cache.DBNoError {
			fmt.Printf("error while running command: %s\n", dbErr.Error())
			_, err = conn.Write([]byte(dbErr.Error() + "\n"))
			if err != nil {
				fmt.Printf("error while sending error packet: %s\n", err)
			}
			continue
		}

		output := <-s.outputChan

		_, err = conn.Write([]byte(output + "\n"))
		if err != nil {
			fmt.Printf("error while sending output packet: %s\n", err)
			continue
		}
	}
}

func (s *Server) Listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			continue
		}

		go s.handleConn(conn)
	}
}

func (s *Server) Close() {
	s.listener.Close()
}
