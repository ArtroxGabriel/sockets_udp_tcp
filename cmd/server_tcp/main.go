package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"sockets_udp_tcp/pkg/calc"
	"sockets_udp_tcp/pkg/protocol"
)

const defaultPort = 8082

func main() {
	port := flag.Int("port", defaultPort, "TCP port to listen on")
	flag.Parse()

	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to bind TCP listener on %s: %v", addr, err)
	}
	defer listener.Close()

	log.Printf("CalcServerTCP listening on %s", listener.Addr())
	handleGracefulShutdown(listener)

	acceptConnections(listener)
}

func handleGracefulShutdown(listener net.Listener) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("\nCalcServerTCP shutting down gracefully...")
		listener.Close()
		os.Exit(0)
	}()
}

func acceptConnections(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Printf("accept error: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	remoteAddr := conn.RemoteAddr()
	log.Printf("new client connected: %s", remoteAddr)

	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("read error from %s: %v", remoteAddr, err)
			}
			break
		}

		resp := processLine(line)
		if _, writeErr := conn.Write([]byte(resp + "\n")); writeErr != nil {
			log.Printf("write error to %s: %v", remoteAddr, writeErr)
			break
		}
	}
	log.Printf("client disconnected: %s", remoteAddr)
}

func processLine(line string) string {
	req, err := protocol.ParseRequest(line)
	if err != nil {
		return protocol.FormatError(-1, err.Error())
	}
	res, computeErr := calc.Compute(req.Op1, req.Op, req.Op2)
	if computeErr != nil {
		return protocol.FormatError(req.Seq, calc.ErrorMessage(computeErr))
	}
	return protocol.FormatResponse(req.Seq, res)
}
