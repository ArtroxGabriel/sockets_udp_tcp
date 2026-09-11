package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/protobuf/proto"

	"sockets_udp_tcp/pkg/calc"
	pb "sockets_udp_tcp/pkg/pb/calc"
	"sockets_udp_tcp/pkg/protocol"
)

const defaultPort = 8083

func main() {
	port := flag.Int("port", defaultPort, "Protobuf TCP port to listen on")
	flag.Parse()

	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to bind Protobuf TCP listener on %s: %v", addr, err)
	}
	defer listener.Close()

	log.Printf("CalcServerProto listening on %s", listener.Addr())
	handleGracefulShutdown(listener)

	acceptConnections(listener)
}

func handleGracefulShutdown(listener net.Listener) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("\nCalcServerProto shutting down gracefully...")
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
	log.Printf("new protobuf client connected: %s", remoteAddr)

	for {
		payload, err := protocol.ReadMsg(conn)
		if err != nil {
			if err != io.EOF {
				log.Printf("read error from %s: %v", remoteAddr, err)
			}
			break
		}

		respBytes, processErr := processProtobufRequest(remoteAddr, payload)
		if processErr != nil {
			log.Printf("process error: %v", processErr)
			break
		}

		if writeErr := protocol.WriteMsg(conn, respBytes); writeErr != nil {
			log.Printf("write error to %s: %v", remoteAddr, writeErr)
			break
		}
	}
	log.Printf("protobuf client disconnected: %s", remoteAddr)
}

func processProtobufRequest(remoteAddr net.Addr, payload []byte) ([]byte, error) {
	var req pb.CalcRequest
	if err := proto.Unmarshal(payload, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	opStr, err := mapProtoOpToString(req.GetOp())
	log.Printf("[RECV] from %s: seq=%d CALC(%v %s %v)", remoteAddr, req.GetSeq(), req.GetOperand1(), opStr, req.GetOperand2())

	var resp pb.CalcResponse
	resp.Seq = req.GetSeq()

	if err != nil {
		resp.Status = pb.Status_STATUS_ERROR
		resp.ErrorMessage = err.Error()
	} else {
		res, computeErr := calc.Compute(req.GetOperand1(), opStr, req.GetOperand2())
		if computeErr != nil {
			resp.Status = pb.Status_STATUS_ERROR
			resp.ErrorMessage = calc.ErrorMessage(computeErr)
		} else {
			resp.Status = pb.Status_STATUS_OK
			resp.Result = res
		}
	}

	if resp.Status == pb.Status_STATUS_OK {
		log.Printf("[SENT] to %s: seq=%d RESULT:%v", remoteAddr, resp.GetSeq(), resp.GetResult())
	} else {
		log.Printf("[SENT] to %s: seq=%d ERROR:%s", remoteAddr, resp.GetSeq(), resp.GetErrorMessage())
	}

	return proto.Marshal(&resp)
}

func mapProtoOpToString(op pb.Operation) (string, error) {
	switch op {
	case pb.Operation_OP_ADD:
		return "+", nil
	case pb.Operation_OP_SUBTRACT:
		return "-", nil
	case pb.Operation_OP_MULTIPLY:
		return "*", nil
	case pb.Operation_OP_DIVIDE:
		return "/", nil
	default:
		return "", fmt.Errorf("unsupported operation enum: %v", op)
	}
}
