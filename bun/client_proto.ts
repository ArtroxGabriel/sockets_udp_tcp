import * as net from "node:net";
import * as path from "node:path";
import protobuf from "protobufjs";

const DEFAULT_PORT = 8083;
const DEFAULT_HOST = "127.0.0.1";
const DEFAULT_REQUEST_COUNT = 20;
const PREFIX_BYTES = 4;

interface ProtoResult {
  seq: number;
  rttMs: number;
  sentBytes: number;
  recvBytes: number;
  response: string;
}

function readExact(client: net.Socket, count: number, state: { buffer: Buffer }): Promise<Buffer> {
  return new Promise((resolve) => {
    const tryExtract = () => {
      if (state.buffer.length >= count) {
        const result = state.buffer.subarray(0, count);
        state.buffer = state.buffer.subarray(count);
        return result;
      }
      return null;
    };

    const extracted = tryExtract();
    if (extracted) {
      resolve(extracted);
      return;
    }

    const onData = (chunk: Buffer) => {
      state.buffer = Buffer.concat([state.buffer, chunk]);
      const res = tryExtract();
      if (res) {
        client.off("data", onData);
        resolve(res);
      }
    };
    client.on("data", onData);
  });
}

async function readFramedMessage(client: net.Socket, state: { buffer: Buffer }): Promise<Buffer> {
  const header = await readExact(client, PREFIX_BYTES, state);
  const length = header.readUInt32BE(0);
  return readExact(client, length, state);
}

function writeFramedMessage(client: net.Socket, payload: Uint8Array): void {
  const header = Buffer.alloc(PREFIX_BYTES);
  header.writeUInt32BE(payload.length, 0);
  client.write(Buffer.concat([header, Buffer.from(payload)]));
}

function printSummary(results: ProtoResult[], totalDurationMs: number): void {
  const avgRtt = results.reduce((acc, r) => acc + r.rttMs, 0) / results.length;
  const maxRtt = Math.max(...results.map((r) => r.rttMs));
  const avgSent = results.reduce((acc, r) => acc + r.sentBytes, 0) / results.length;
  const avgRecv = results.reduce((acc, r) => acc + r.recvBytes, 0) / results.length;

  console.log("\n========== RESUMO BUN (PROTOBUF) ==========");
  console.log(`Tempo total da sequência : ${totalDurationMs.toFixed(2)}ms`);
  console.log(`Requisições entregues    : ${results.length}`);
  console.log(`RTT médio                : ${avgRtt.toFixed(2)}ms`);
  console.log(`RTT máximo               : ${maxRtt.toFixed(2)}ms`);
  console.log(`Tamanho médio da req     : ${avgSent.toFixed(1)} bytes`);
  console.log(`Tamanho médio da resp    : ${avgRecv.toFixed(1)} bytes`);
  console.log("===========================================\n");
}

async function main(): Promise<void> {
  const protoPath = path.resolve(import.meta.dir, "../proto/calc.proto");
  const root = await protobuf.load(protoPath);
  const CalcRequest = root.lookupType("calc.CalcRequest");
  const CalcResponse = root.lookupType("calc.CalcResponse");

  const port = Number(process.env.PORT) || DEFAULT_PORT;
  const host = process.env.HOST || DEFAULT_HOST;

  const client = net.createConnection({ port, host });
  await new Promise<void>((resolve, reject) => {
    client.once("connect", resolve);
    client.once("error", reject);
  });
  console.log(`Bun Protobuf Client connected to ${host}:${port}`);

  const bufferState = { buffer: Buffer.alloc(0) };
  const ops = [1, 2, 3, 4]; // OP_ADD, OP_SUBTRACT, OP_MULTIPLY, OP_DIVIDE
  const results: ProtoResult[] = [];

  const startSeqTime = performance.now();
  for (let i = 0; i < DEFAULT_REQUEST_COUNT; i++) {
    const isDivZero = i === 1;
    const reqPayload = CalcRequest.encode({
      seq: i,
      operand1: (i + 1) * 5 + 0.5,
      op: isDivZero ? 4 : ops[i % ops.length],
      operand2: isDivZero ? 0 : i + 1,
    }).finish();

    const opNames = ["ADD", "SUBTRACT", "MULTIPLY", "DIVIDE"];
    const opName = isDivZero ? "DIVIDE" : opNames[(ops[i % ops.length] - 1) % opNames.length];
    console.log(`[BUN PROTO SEND] seq=${i} CALC(${opName})`);

    const startReq = performance.now();
    writeFramedMessage(client, reqPayload);
    const respPayload = await readFramedMessage(client, bufferState);
    const rttMs = performance.now() - startReq;

    const resp = CalcResponse.decode(respPayload) as any;
    const desc = resp.status === 1 ? `RESULT: ${resp.result}` : `ERROR: ${resp.errorMessage}`;
    console.log(`[BUN PROTO RECV] seq=${i} -> ${desc} (RTT: ${rttMs.toFixed(2)}ms, Wire: ${reqPayload.length}B/${respPayload.length}B)`);

    results.push({
      seq: i,
      rttMs,
      sentBytes: reqPayload.length,
      recvBytes: respPayload.length,
      response: desc,
    });
  }
  const totalDuration = performance.now() - startSeqTime;

  client.end();
  printSummary(results, totalDuration);
}

main().catch(console.error);
