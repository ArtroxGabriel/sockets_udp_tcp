package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"sockets_udp_tcp/pkg/protocol"
)

const (
	defaultServerAddr = "127.0.0.1:8082"
	defaultRequestNum = 20
)

// RequestMetric holds latency and wire size information for a request.
type RequestMetric struct {
	Seq      int
	RTT      time.Duration
	SentSize int
	RecvSize int
	Response string
}

// ClientMetrics aggregates summary statistics.
type ClientMetrics struct {
	TotalTime   time.Duration
	Delivered   int
	Lost        int
	AvgRTT      time.Duration
	MaxRTT      time.Duration
	AvgSentSize float64
	AvgRecvSize float64
}

func main() {
	addr := flag.String("addr", defaultServerAddr, "TCP server address (host:port)")
	n := flag.Int("n", defaultRequestNum, "Number of requests to send")
	flag.Parse()

	conn, err := net.Dial("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to connect to TCP server %s: %v", *addr, err)
	}
	defer conn.Close()

	log.Printf("CalcClientTCP connected to %s | N=%d", *addr, *n)
	requests := generateTestRequests(*n)

	startTime := time.Now()
	metrics, err := executeRequests(conn, requests)
	if err != nil {
		log.Fatalf("error executing requests: %v", err)
	}
	totalDuration := time.Since(startTime)

	summary := calculateMetrics(metrics, totalDuration)
	printSummary(summary)

	if summary.Lost > 0 {
		os.Exit(1)
	}
}

func generateTestRequests(n int) []protocol.Request {
	operators := []string{"+", "-", "*", "/"}
	requests := make([]protocol.Request, n)
	for i := 0; i < n; i++ {
		op := operators[i%len(operators)]
		op1 := float64((i+1)*5) + 0.5
		op2 := float64(i + 1)
		if i == 1 {
			op = "/"
			op2 = 0
		}
		requests[i] = protocol.Request{Seq: i, Op1: op1, Op: op, Op2: op2}
	}
	return requests
}

func executeRequests(conn net.Conn, reqs []protocol.Request) ([]RequestMetric, error) {
	reader := bufio.NewReader(conn)
	metrics := make([]RequestMetric, len(reqs))

	for i, req := range reqs {
		line := protocol.FormatRequest(req) + "\n"
		start := time.Now()

		if _, err := conn.Write([]byte(line)); err != nil {
			return nil, fmt.Errorf("failed to write seq %d: %w", req.Seq, err)
		}

		respLine, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read response for seq %d: %w", req.Seq, err)
		}
		rtt := time.Since(start)

		cleanResp := strings.TrimSpace(respLine)
		log.Printf("[SUCCESS] seq=%d -> %s (RTT: %v, Wire: %dB/%dB)",
			req.Seq, cleanResp, rtt, len(line), len(respLine))

		metrics[i] = RequestMetric{
			Seq:      req.Seq,
			RTT:      rtt,
			SentSize: len(line),
			RecvSize: len(respLine),
			Response: cleanResp,
		}
	}
	return metrics, nil
}

func calculateMetrics(items []RequestMetric, totalDuration time.Duration) ClientMetrics {
	var totalRTT, maxRTT time.Duration
	var totalSent, totalRecv int

	for _, item := range items {
		totalRTT += item.RTT
		totalSent += item.SentSize
		totalRecv += item.RecvSize
		if item.RTT > maxRTT {
			maxRTT = item.RTT
		}
	}

	count := len(items)
	var avgRTT time.Duration
	var avgSent, avgRecv float64
	if count > 0 {
		avgRTT = totalRTT / time.Duration(count)
		avgSent = float64(totalSent) / float64(count)
		avgRecv = float64(totalRecv) / float64(count)
	}

	return ClientMetrics{
		TotalTime:   totalDuration,
		Delivered:   count,
		Lost:        0,
		AvgRTT:      avgRTT,
		MaxRTT:      maxRTT,
		AvgSentSize: avgSent,
		AvgRecvSize: avgRecv,
	}
}

func printSummary(m ClientMetrics) {
	fmt.Println("\n========== RESUMO DA EXECUÇÃO (TCP) ==========")
	fmt.Printf("Tempo total da sequência : %v\n", m.TotalTime)
	fmt.Printf("Requisições entregues    : %d\n", m.Delivered)
	fmt.Printf("Requisições perdidas     : %d\n", m.Lost)
	fmt.Printf("RTT médio                : %v\n", m.AvgRTT)
	fmt.Printf("RTT máximo               : %v\n", m.MaxRTT)
	fmt.Printf("Tamanho médio da req     : %.1f bytes\n", m.AvgSentSize)
	fmt.Printf("Tamanho médio da resp    : %.1f bytes\n", m.AvgRecvSize)
	fmt.Println("===============================================")
}
