package server

import (
	"fmt"
	"io"
	"log/slog"
	"net"

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
				slog.Info("client closed connection", "remoteIP", conn.RemoteAddr().String())
				return
			}
			slog.Error("error while reading packet", "error", err)
			continue
		}
		commandStr := string(buffer[:n])

		commands, err := cache.ParseCommandsFromString(commandStr)
		if err != nil {
			slog.Error("error while parsing packet", "error", err, "remoteIP", conn.RemoteAddr().String())
			_, err = conn.Write([]byte(err.Error() + "\n"))
			if err != nil {
				slog.Error("error while sending error packet", "error", err, "remoteIP", conn.RemoteAddr().String())
			}
			continue
		}

		for _, command := range commands {
			s.cmdChan <- command

			dbErr := <-s.errChan
			if dbErr.Kind() != cache.DBNoError {
				slog.Error("error while running command", "error", dbErr.Error(), "remoteIP", conn.RemoteAddr().String())
				_, err = conn.Write([]byte(dbErr.Error() + "\n"))
				if err != nil {
					slog.Error("error while sending error packet", "error", err, "remoteIP", conn.RemoteAddr().String())
				}
				continue
			}

			output := <-s.outputChan

			_, err = conn.Write([]byte(output + "\n"))
			if err != nil {
				slog.Error("error while sending output packet", "error", err)
				continue
			}
		}
	}
}

func (s *Server) Listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			slog.Error("unable to start cash", "error", err)
			continue
		}

		go s.handleConn(conn)
	}
}

func (s *Server) Close() {
	s.listener.Close()
	slog.Info("closing cash")
}
