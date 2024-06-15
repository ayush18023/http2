package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
)

const defaultListenAddr = ":5001"

type Config struct {
	ListenAddr string
}

type Request struct {
	conn net.Conn
}

func NewConn(con net.Conn) *Request {
	return &Request{
		conn: con,
	}
}

type Server struct {
	Config
	ln       net.Listener
	requestQ chan *Request
	quitCh   chan struct{}
}

func NewServer(cfg Config) *Server {
	if len(cfg.ListenAddr) == 0 {
		cfg.ListenAddr = defaultListenAddr
	}
	return &Server{
		Config:   cfg,
		requestQ: make(chan *Request),
		quitCh:   make(chan struct{}),
	}
}

// testing this comment out
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	s.ln = ln

	go s.loop()

	slog.Info("server running", "listenAddr", s.ListenAddr)

	return s.acceptLoop()
}

func (s *Server) loop() {
	// this will resolve the request
	for {
		select {
		case _ = <-s.requestQ:
			fmt.Println("Request recieved")
			// req.conn.readLoop()
		case <-s.quitCh:
			return
		}
	}
}

func (s *Server) acceptLoop() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			slog.Error("accept error", "err", err)
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	s.requestQ <- NewConn(conn)
}

func main() {
	listenAddr := flag.String("listenAddr", defaultListenAddr, "listen address")
	flag.Parse()
	server := NewServer(Config{
		ListenAddr: *listenAddr,
	})
	log.Fatal(server.Start())
}
