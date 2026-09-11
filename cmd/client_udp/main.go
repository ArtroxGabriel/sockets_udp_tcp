package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"sockets_udp_tcp/pkg/protocol"
)

const (
	defaultServerAddr  = "127.0.0.1:8081"
	defaultRequestNum  = 20
	defaultTimeout     = 500 * time.Millisecond
	defaultMaxAttempts = 5
	udpBufferSize      = 1024
)

// RequestResult tracks metrics for an individual request.
type RequestResult struct {
	Seq             int
	Delivered       bool
	RTT             time.Duration
	Retransmissions int
	SentSize        int
	RecvSize        int
	Response        string
}

// ClientMetrics aggregates the results across all requests.
type ClientMetrics struct {
	TotalTime       time.Duration `json:"-"`
	TotalDurationMs float64       `json:"totalDurationMs"`
	DeliveredCount  int           `json:"delivered"`
	LostCount       int           `json:"lost"`
	Retransmissions int           `json:"retransmissions"`
	AvgRTT          time.Duration `json:"-"`
	AvgRTTMs        float64       `json:"avgRttMs"`
	MaxRTT          time.Duration `json:"-"`
	MaxRTTMs        float64       `json:"maxRttMs"`
	AvgSentBytes    float64       `json:"avgSentBytes"`
	AvgRecvBytes    float64       `json:"avgRecvBytes"`
}

func main() {
	addr := flag.String("addr", defaultServerAddr, "UDP server address (host:port)")
	n := flag.Int("n", defaultRequestNum, "Number of requests to send")
	timeout := flag.Duration("timeout", defaultTimeout, "Timeout per attempt")
	maxAttempts := flag.Int("max-attempts", defaultMaxAttempts, "Maximum attempts per request")
	jsonOutput := flag.Bool("json", false, "Output metrics in JSON format")
	flag.Parse()

	conn, err := net.Dial("udp", *addr)
	if err != nil {
		log.Fatalf("failed to dial UDP server %s: %v", *addr, err)
	}
	defer conn.Close()

	if !*jsonOutput {
		log.Printf("CalcClientUDP connected to %s | N=%d | Timeout=%v | MaxAttempts=%d", *addr, *n, *timeout, *maxAttempts)
	}
	requests := generateTestRequests(*n)

	startTime := time.Now()
	results := executeRequests(conn, requests, *timeout, *maxAttempts, *jsonOutput)
	totalDuration := time.Since(startTime)

	metrics := calculateMetrics(results, totalDuration)
	if *jsonOutput {
		printJSONSummary(metrics)
	} else {
		printSummary(metrics)
	}

	if metrics.LostCount > 0 {
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
		// Include a division by zero test case at seq 1 as in the assignment specification
		if i == 1 {
			op = "/"
			op2 = 0
		}
		requests[i] = protocol.Request{Seq: i, Op1: op1, Op: op, Op2: op2}
	}
	return requests
}

func executeRequests(conn net.Conn, reqs []protocol.Request, timeout time.Duration, maxAttempts int, silent bool) []RequestResult {
	results := make([]RequestResult, len(reqs))
	for i, req := range reqs {
		results[i] = sendWithRetry(conn, req, timeout, maxAttempts, silent)
	}
	return results
}

func sendWithRetry(conn net.Conn, req protocol.Request, timeout time.Duration, maxAttempts int, silent bool) RequestResult {
	payload := []byte(protocol.FormatRequest(req))
	var retransmissions int

	startReqTime := time.Now()
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			retransmissions++
			if !silent {
				log.Printf("[RETRY %d/%d] Resending req seq=%d", attempt-1, maxAttempts-1, req.Seq)
			}
		}

		if _, err := conn.Write(payload); err != nil {
			if !silent {
				log.Printf("write error on seq=%d: %v", req.Seq, err)
			}
		}

		respStr, err := readResponseWithTimeout(conn, timeout)
		if err == nil {
			rtt := time.Since(startReqTime)
			if !silent {
				log.Printf("[SUCCESS] seq=%d -> %s (RTT: %v)", req.Seq, respStr, rtt)
			}
			return RequestResult{
				Seq:             req.Seq,
				Delivered:       true,
				RTT:             rtt,
				Retransmissions: retransmissions,
				SentSize:        len(payload),
				RecvSize:        len(respStr),
				Response:        respStr,
			}
		}

		if !silent {
			log.Printf("[TIMEOUT] seq=%d attempt %d timed out after %v", req.Seq, attempt, timeout)
		}
	}

	if !silent {
		log.Printf("[LOST] seq=%d exhausted %d attempts", req.Seq, maxAttempts)
	}
	return RequestResult{Seq: req.Seq, Delivered: false, Retransmissions: retransmissions, SentSize: len(payload)}
}

func readResponseWithTimeout(conn net.Conn, timeout time.Duration) (string, error) {
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return "", err
	}
	buf := make([]byte, udpBufferSize)
	n, err := conn.Read(buf)
	if err != nil {
		return "", err
	}
	return string(buf[:n]), nil
}

func calculateMetrics(results []RequestResult, totalDuration time.Duration) ClientMetrics {
	var totalRTT, maxRTT time.Duration
	var totalSent, totalRecv int
	var delivered, lost, retries int

	for _, r := range results {
		retries += r.Retransmissions
		totalSent += r.SentSize
		if r.Delivered {
			delivered++
			totalRTT += r.RTT
			totalRecv += r.RecvSize
			if r.RTT > maxRTT {
				maxRTT = r.RTT
			}
		} else {
			lost++
		}
	}

	var avgRTT time.Duration
	var avgSent, avgRecv float64
	totalCount := len(results)
	if delivered > 0 {
		avgRTT = totalRTT / time.Duration(delivered)
		avgRecv = float64(totalRecv) / float64(delivered)
	}
	if totalCount > 0 {
		avgSent = float64(totalSent) / float64(totalCount)
	}

	return ClientMetrics{
		TotalTime:       totalDuration,
		TotalDurationMs: float64(totalDuration.Microseconds()) / 1000.0,
		DeliveredCount:  delivered,
		LostCount:       lost,
		Retransmissions: retries,
		AvgRTT:          avgRTT,
		AvgRTTMs:        float64(avgRTT.Microseconds()) / 1000.0,
		MaxRTT:          maxRTT,
		MaxRTTMs:        float64(maxRTT.Microseconds()) / 1000.0,
		AvgSentBytes:    avgSent,
		AvgRecvBytes:    avgRecv,
	}
}

func printSummary(m ClientMetrics) {
	fmt.Println("\n========== RESUMO DA EXECUÇÃO (UDP) ==========")
	fmt.Printf("Tempo total da sequência : %v\n", m.TotalTime)
	fmt.Printf("Requisições entregues    : %d\n", m.DeliveredCount)
	fmt.Printf("Requisições perdidas     : %d\n", m.LostCount)
	fmt.Printf("Número de retransmissões : %d\n", m.Retransmissions)
	fmt.Printf("RTT médio                : %v\n", m.AvgRTT)
	fmt.Printf("RTT máximo               : %v\n", m.MaxRTT)
	fmt.Printf("Tamanho médio da req     : %.1f bytes\n", m.AvgSentBytes)
	fmt.Printf("Tamanho médio da resp    : %.1f bytes\n", m.AvgRecvBytes)
	fmt.Println("===============================================")
}

func printJSONSummary(m ClientMetrics) {
	enc := json.NewEncoder(os.Stdout)
	_ = enc.Encode(m)
}
