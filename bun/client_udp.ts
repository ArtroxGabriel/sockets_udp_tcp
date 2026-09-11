import * as dgram from "node:dgram";

const DEFAULT_PORT = 8081;
const DEFAULT_HOST = "127.0.0.1";
const DEFAULT_REQUEST_COUNT = 20;
const TIMEOUT_MS = 500;
const MAX_ATTEMPTS = 5;

interface RequestItem {
  seq: number;
  op1: number;
  op: string;
  op2: number;
}

interface RequestResult {
  seq: number;
  delivered: boolean;
  rttMs: number;
  retransmissions: number;
  response: string;
}

function generateRequests(n: number): RequestItem[] {
  const ops = ["+", "-", "*", "/"];
  return Array.from({ length: n }, (_, i) => ({
    seq: i,
    op1: (i + 1) * 5 + 0.5,
    op: i === 1 ? "/" : ops[i % ops.length],
    op2: i === 1 ? 0 : i + 1,
  }));
}

function formatRequest(item: RequestItem): string {
  return `CALC:${item.seq}:${item.op1}:${item.op}:${item.op2}`;
}

function sendPacket(socket: dgram.Socket, message: string, port: number, host: string): Promise<void> {
  return new Promise((resolve, reject) => {
    socket.send(Buffer.from(message), port, host, (err) => {
      if (err) reject(err);
      else resolve();
    });
  });
}

function waitForResponse(socket: dgram.Socket, timeoutMs: number): Promise<string> {
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => {
      socket.removeAllListeners("message");
      reject(new Error("TIMEOUT"));
    }, timeoutMs);

    socket.once("message", (msg) => {
      clearTimeout(timer);
      resolve(msg.toString());
    });
  });
}

async function sendWithRetry(
  socket: dgram.Socket,
  req: RequestItem,
  port: number,
  host: string
): Promise<RequestResult> {
  const message = formatRequest(req);
  let retries = 0;
  const startTime = performance.now();

  for (let attempt = 1; attempt <= MAX_ATTEMPTS; attempt++) {
    if (attempt > 1) {
      retries++;
      console.log(`[BUN UDP RETRY ${attempt - 1}] seq=${req.seq}`);
    }

    console.log(`[BUN UDP SEND] ${message}`);
    await sendPacket(socket, message, port, host);
    try {
      const response = await waitForResponse(socket, TIMEOUT_MS);
      const rttMs = performance.now() - startTime;
      console.log(`[BUN UDP RECV] seq=${req.seq} -> ${response} (RTT: ${rttMs.toFixed(2)}ms)`);
      return { seq: req.seq, delivered: true, rttMs, retransmissions: retries, response };
    } catch {
      console.log(`[BUN UDP TIMEOUT] seq=${req.seq} attempt ${attempt} timed out`);
    }
  }

  console.log(`[BUN UDP LOST] seq=${req.seq} exhausted ${MAX_ATTEMPTS} attempts`);
  return { seq: req.seq, delivered: false, rttMs: 0, retransmissions: retries, response: "" };
}

function printSummary(results: RequestResult[], totalDurationMs: number): void {
  const delivered = results.filter((r) => r.delivered);
  const totalRetries = results.reduce((acc, r) => acc + r.retransmissions, 0);
  const avgRtt = delivered.length > 0 ? delivered.reduce((acc, r) => acc + r.rttMs, 0) / delivered.length : 0;
  const maxRtt = delivered.length > 0 ? Math.max(...delivered.map((r) => r.rttMs)) : 0;

  console.log("\n========== RESUMO BUN (UDP) ==========");
  console.log(`Tempo total da sequência : ${totalDurationMs.toFixed(2)}ms`);
  console.log(`Requisições entregues    : ${delivered.length}`);
  console.log(`Requisições perdidas     : ${results.length - delivered.length}`);
  console.log(`Retransmissões           : ${totalRetries}`);
  console.log(`RTT médio                : ${avgRtt.toFixed(2)}ms`);
  console.log(`RTT máximo               : ${maxRtt.toFixed(2)}ms`);
  console.log("======================================\n");
}

async function main(): Promise<void> {
  const port = Number(process.env.PORT) || DEFAULT_PORT;
  const host = process.env.HOST || DEFAULT_HOST;
  const socket = dgram.createSocket("udp4");

  console.log(`Bun UDP Client connecting to ${host}:${port}`);
  const requests = generateRequests(DEFAULT_REQUEST_COUNT);
  const results: RequestResult[] = [];

  const startSeqTime = performance.now();
  for (const req of requests) {
    const res = await sendWithRetry(socket, req, port, host);
    results.push(res);
  }
  const totalDuration = performance.now() - startSeqTime;

  socket.close();
  printSummary(results, totalDuration);
}

main().catch(console.error);
