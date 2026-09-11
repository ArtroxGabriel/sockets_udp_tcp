package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/protobuf/proto"

	pb "sockets_udp_tcp/pkg/pb/calc"
	"sockets_udp_tcp/pkg/protocol"
)

const (
	defaultServerAddr = "127.0.0.1:8083"
	defaultRequestNum = 20
)

// ProtoMetric tracks latency and payload sizes for protobuf messages.
type ProtoMetric struct {
	Seq      int32
	RTT      time.Duration
	SentSize int
	RecvSize int
	Response *pb.CalcResponse
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
	addr := flag.String("addr", defaultServerAddr, "Protobuf TCP server address (host:port)")
	n := flag.Int("n", defaultRequestNum, "Number of requests to send")
	flag.Parse()

	conn, err := net.Dial("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to connect to Protobuf server %s: %v", *addr, err)
	}
	defer conn.Close()

	log.Printf("CalcClientProto connected to %s | N=%d", *addr, *n)
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

func generateTestRequests(n int) []*pb.CalcRequest {
	ops := []pb.Operation{
		pb.Operation_OP_ADD,
		pb.Operation_OP_SUBTRACT,
		pb.Operation_OP_MULTIPLY,
		pb.Operation_OP_DIVIDE,
	}

	requests := make([]*pb.CalcRequest, n)
	for i := 0; i < n; i++ {
		op := ops[i%len(ops)]
		op1 := float64((i+1)*5) + 0.5
		op2 := float64(i + 1)
		if i == 1 {
			op = pb.Operation_OP_DIVIDE
			op2 = 0
		}
		requests[i] = &pb.CalcRequest{
			Seq:      int32(i),
			Operand1: op1,
			Op:       op,
			Operand2: op2,
		}
	}
	return requests
}

func executeRequests(conn net.Conn, reqs []*pb.CalcRequest) ([]ProtoMetric, error) {
	metrics := make([]ProtoMetric, len(reqs))

	for i, req := range reqs {
		payload, err := proto.Marshal(req)
		if err != nil {
			return nil, fmt.Errorf("marshal error on seq %d: %w", req.GetSeq(), err)
		}

		if !silent {
			log.Printf("[SEND] seq=%d CALC(%v %s %v)", req.GetSeq(), req.GetOperand1(), mapProtoOpToString(req.GetOp()), req.GetOperand2())
		}

		start := time.Now()
		if err := protocol.WriteMsg(conn, payload); err != nil {
			return nil, fmt.Errorf("write error on seq %d: %w", req.GetSeq(), err)
		}

		respBytes, err := protocol.ReadMsg(conn)
		if err != nil {
			return nil, fmt.Errorf("read error on seq %d: %w", req.GetSeq(), err)
		}
		rtt := time.Since(start)

		var resp pb.CalcResponse
		if err := proto.Unmarshal(respBytes, &resp); err != nil {
			return nil, fmt.Errorf("unmarshal error on seq %d: %w", req.GetSeq(), err)
		}

		logSuccess(req.GetSeq(), &resp, rtt, len(payload), len(respBytes))
		metrics[i] = ProtoMetric{
			Seq:      req.GetSeq(),
			RTT:      rtt,
			SentSize: len(payload),
			RecvSize: len(respBytes),
			Response: &resp,
		}
	}
	return metrics, nil
}

func mapProtoOpToString(op pb.Operation) string {
	switch op {
	case pb.Operation_OP_ADD:
		return "+"
	case pb.Operation_OP_SUBTRACT:
		return "-"
	case pb.Operation_OP_MULTIPLY:
		return "*"
	case pb.Operation_OP_DIVIDE:
		return "/"
	default:
		return "?"
	}
}

func logSuccess(seq int32, resp *pb.CalcResponse, rtt time.Duration, sentBytes, recvBytes int) {
	if resp.GetStatus() == pb.Status_STATUS_OK {
		log.Printf("[RECV] seq=%d -> RESULT: %v (RTT: %v, Wire: %dB/%dB)",
			seq, resp.GetResult(), rtt, sentBytes, recvBytes)
	} else {
		log.Printf("[RECV] seq=%d -> ERROR: %s (RTT: %v, Wire: %dB/%dB)",
			seq, resp.GetErrorMessage(), rtt, sentBytes, recvBytes)
	}
}

func calculateMetrics(items []ProtoMetric, totalDuration time.Duration) ClientMetrics {
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
	fmt.Println("\n========== RESUMO DA EXECUÇÃO (PROTOBUF) ==========")
	fmt.Printf("Tempo total da sequência : %v\n", m.TotalTime)
	fmt.Printf("Requisições entregues    : %d\n", m.Delivered)
	fmt.Printf("Requisições perdidas     : %d\n", m.Lost)
	fmt.Printf("RTT médio                : %v\n", m.AvgRTT)
	fmt.Printf("RTT máximo               : %v\n", m.MaxRTT)
	fmt.Printf("Tamanho médio da req     : %.1f bytes (payload protobuf)\n", m.AvgSentSize)
	fmt.Printf("Tamanho médio da resp    : %.1f bytes (payload protobuf)\n", m.AvgRecvSize)
	fmt.Println("===================================================")
}
