package main

import (
	"flag"
	"log"
	"math/rand/v2"
	"net"
	"os"
	"os/signal"
	"syscall"

	"sockets_udp_tcp/pkg/calc"
	"sockets_udp_tcp/pkg/protocol"
)

const (
	udpBufferSize = 1024
	defaultPort   = 8081
)

func main() {
	port := flag.Int("port", defaultPort, "UDP port to listen on")
	lossRate := flag.Float64("loss-rate", 0.0, "Fraction of packets to drop intentionally (0.0 to 1.0)")
	flag.Parse()

	addr := net.UDPAddr{Port: *port, IP: net.ParseIP("0.0.0.0")}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatalf("failed to bind UDP socket on port %d: %v", *port, err)
	}
	defer conn.Close()

	log.Printf("CalcServerUDP listening on %s with loss-rate %.1f%%", conn.LocalAddr(), *lossRate*100)
	handleGracefulShutdown(conn)

	listenLoop(conn, *lossRate)
}

func handleGracefulShutdown(conn *net.UDPConn) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("\nCalcServerUDP shutting down gracefully...")
		conn.Close()
		os.Exit(0)
	}()
}

func listenLoop(conn *net.UDPConn, lossRate float64) {
	buf := make([]byte, udpBufferSize)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("read error: %v", err)
			return
		}

		packet := make([]byte, n)
		copy(packet, buf[:n])
		go processPacket(conn, clientAddr, packet, lossRate)
	}
}

func shouldDrop(lossRate float64) bool {
	if lossRate <= 0 {
		return false
	}
	return rand.Float64() < lossRate
}

func processPacket(conn *net.UDPConn, clientAddr *net.UDPAddr, data []byte, lossRate float64) {
	raw := string(data)
	req, err := protocol.ParseRequest(raw)
	if err != nil {
		log.Printf("malformed request from %s: %v", clientAddr, err)
		return
	}

	if shouldDrop(lossRate) {
		log.Printf("[SIMULATED LOSS] Dropped req seq=%d from %s", req.Seq, clientAddr)
		return
	}

	respStr := executeCalculation(req)
	if _, err := conn.WriteToUDP([]byte(respStr), clientAddr); err != nil {
		log.Printf("error replying to %s: %v", clientAddr, err)
	}
}

func executeCalculation(req protocol.Request) string {
	res, err := calc.Compute(req.Op1, req.Op, req.Op2)
	if err != nil {
		return protocol.FormatError(req.Seq, calc.ErrorMessage(err))
	}
	return protocol.FormatResponse(req.Seq, res)
}
