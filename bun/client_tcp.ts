import * as net from "node:net";

const DEFAULT_PORT = 8082;
const DEFAULT_HOST = "127.0.0.1";
const DEFAULT_REQUEST_COUNT = 20;

interface RequestItem {
  seq: number;
  op1: number;
  op: string;
  op2: number;
}

interface RequestResult {
  seq: number;
  rttMs: number;
  sentBytes: number;
  recvBytes: number;
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

function readLine(client: net.Socket, bufferState: { buffer: string }): Promise<string> {
  return new Promise((resolve) => {
    const checkBuffer = () => {
      const idx = bufferState.buffer.indexOf("\n");
      if (idx !== -1) {
        const line = bufferState.buffer.slice(0, idx).trim();
        bufferState.buffer = bufferState.buffer.slice(idx + 1);
        return line;
      }
      return null;
    };

    const existing = checkBuffer();
    if (existing !== null) {
      resolve(existing);
      return;
    }

    const onData = (chunk: Buffer) => {
      bufferState.buffer += chunk.toString();
      const line = checkBuffer();
      if (line !== null) {
        client.off("data", onData);
        resolve(line);
      }
    };
    client.on("data", onData);
  });
}

function printSummary(results: RequestResult[], totalDurationMs: number): void {
  const avgRtt = results.reduce((acc, r) => acc + r.rttMs, 0) / results.length;
  const maxRtt = Math.max(...results.map((r) => r.rttMs));
  const avgSent = results.reduce((acc, r) => acc + r.sentBytes, 0) / results.length;
  const avgRecv = results.reduce((acc, r) => acc + r.recvBytes, 0) / results.length;

  console.log("\n========== RESUMO BUN (TCP) ==========");
  console.log(`Tempo total da sequência : ${totalDurationMs.toFixed(2)}ms`);
  console.log(`Requisições entregues    : ${results.length}`);
  console.log(`RTT médio                : ${avgRtt.toFixed(2)}ms`);
  console.log(`RTT máximo               : ${maxRtt.toFixed(2)}ms`);
  console.log(`Tamanho médio da req     : ${avgSent.toFixed(1)} bytes`);
  console.log(`Tamanho médio da resp    : ${avgRecv.toFixed(1)} bytes`);
  console.log("======================================\n");
}

async function main(): Promise<void> {
  const port = Number(process.env.PORT) || DEFAULT_PORT;
  const host = process.env.HOST || DEFAULT_HOST;

  const client = net.createConnection({ port, host });
  await new Promise<void>((resolve, reject) => {
    client.once("connect", resolve);
    client.once("error", reject);
  });
  console.log(`Bun TCP Client connected to ${host}:${port}`);

  const bufferState = { buffer: "" };
  const requests = generateRequests(DEFAULT_REQUEST_COUNT);
  const results: RequestResult[] = [];

  const startSeqTime = performance.now();
  for (const req of requests) {
    const cleanReq = `CALC:${req.seq}:${req.op1}:${req.op}:${req.op2}`;
    const line = cleanReq + "\n";
    console.log(`[BUN TCP SEND] ${cleanReq}`);
    const startReq = performance.now();

    client.write(line);
    const respLine = await readLine(client, bufferState);
    const rttMs = performance.now() - startReq;

    console.log(`[BUN TCP RECV] seq=${req.seq} -> ${respLine} (RTT: ${rttMs.toFixed(2)}ms)`);
    results.push({
      seq: req.seq,
      rttMs,
      sentBytes: Buffer.byteLength(line),
      recvBytes: Buffer.byteLength(respLine) + 1,
      response: respLine,
    });
  }
  const totalDuration = performance.now() - startSeqTime;

  client.end();
  printSummary(results, totalDuration);
}

main().catch(console.error);
