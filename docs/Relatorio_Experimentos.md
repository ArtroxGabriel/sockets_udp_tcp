# Relatório Comparativo: Sockets UDP vs. TCP vs. Protobuf

Este documento apresenta os dados coletados empiricamente durante a execução dos testes da calculadora remota conforme a especificação da atividade.

## Tabela de Resultados dos Experimentos (N = 20 requisições)

| Cenário | Protocolo | Duração Total (ms) | Entregues | Perdidos | Retransmissões | RTT Médio (ms) | RTT Máximo (ms) | Tam. Médio Req (B) | Tam. Médio Resp (B) |
|---|---|---|---|---|---|---|---|---|---|
| UDP (0% perda) | UDP | 0.95 | 20 | 0 | 0 | 0.044 | 0.100 | 16.1 | 15.4 |
| UDP (10% perda) | UDP | 4007.50 | 20 | 0 | 8 | 200.367 | 1001.170 | 16.1 | 15.4 |
| UDP (30% perda) | UDP | 2507.04 | 20 | 0 | 5 | 125.343 | 501.091 | 16.1 | 15.4 |
| TCP Textual | TCP | 1.08 | 20 | 0 | 0 | 0.050 | 0.171 | 17.1 | 16.4 |
| TCP Protobuf | Protobuf (TCP) | 1.00 | 20 | 0 | 0 | 0.044 | 0.216 | 21.4 | 13.4 |

## Respostas às Questões Elaboradas

### 1. Por que o TCP não perde mensagens e o UDP sim (e o que custa resolver isso)?

- **Semântica do UDP:** O UDP (*User Datagram Protocol*) é um protocolo orientado a datagramas, não confiável e sem conexão. O sistema operacional simplesmente encapsula o pacote IP e despacha para a rede. Não há confirmação de entrega (*ACK*), controle de fluxo, retransmissão de pacotes perdidos ou ordenação. Se um roteador intermediário descartar um pacote por congestionamento, erro de checksum ou saturação de buffer, o pacote é sumariamente descartado sem aviso.
- **Semântica do TCP:** O TCP (*Transmission Control Protocol*) é orientado a conexão e fluxo de bytes confiável. Implementa uma máquina de estados complexa que utiliza números de sequência (*SEQ*), confirmações cumulativas/seletivas (*ACK/SACK*), temporizadores de retransmissão adaptativos (*RTO baseado em RTT estimado*), controle de fluxo por janela deslizante (*sliding window*) e controle de congestionamento (*Slow Start, Congestion Avoidance, Fast Retransmit*).
- **O custo de resolver a perda (no UDP vs. TCP):**
  - **Custo de latência e jitter:** Para tornar o UDP confiável na camada de aplicação (como implementamos na Parte 1), é necessário introduzir mecanismos como timeout e retransmissão (*Stop-and-Wait* ou *ARQ*). Como evidenciado nos dados empíricos da tabela acima, sob taxa de perda de 30%, o RTT médio e a duração total disparam em ordens de magnitude devido às pausas esperando o timeout expirar.
  - **Custo de processamento e overhead de cabeçalho:** O TCP consome mais memória no kernel (buffers de envio e recepção para ordenar segmentos) e possui cabeçalhos maiores (20 a 60 bytes contra apenas 8 bytes do UDP), além da latência inicial do *handshake* de três vias (SYN, SYN-ACK, ACK).
  - **Conclusão:** O UDP é ideal quando baixa latência e previsibilidade imediata superam a necessidade de confiabilidade total (ex.: streaming de áudio/vídeo, jogos online, DNS). O TCP é mandatório quando integridade e confiabilidade dos dados são cruciais (ex.: HTTP, bancos de dados, transferência de arquivos).

### 2. Comparativo de Serialização: Protocolo Textual vs. Protocol Buffers

- **Formato Textual (Partes 1 e 2):**
  - Codifica números em caracteres ASCII/UTF-8 delimitados por `:` (ex.: `CALC:0:10.5:+:4.5`).
  - Vantagens: Facilmente legível por humanos e inspecionável em analisadores de pacotes (`tcpdump`, `Wireshark`).
  - Desvantagens: Parsing textual repetitivo (`strings.Split`, `strconv.ParseFloat`), sem garantia estrita de esquema e tamanho variável.
- **Protocol Buffers (Parte 4):**
  - Codificação binária compacta com enums inteiros compactados via varints e campos tipados estritamente.
  - Vantagens: Serialização e deserialização muito mais rápidas e com menor overhead de CPU, tipagem forte garantida por contrato (`.proto`) e suporte automático entre diferentes linguagens (Go, Bun/TypeScript, Python, etc.).
  - Observação de tamanho: Em mensagens curtas com números decimais de alta precisão (IEEE 754 de 64 bits = 8 bytes binários brutos), o Protobuf possui tamanho ligeiramente superior a inteiros ASCII de 1 dígito, porém escala com imensa vantagem de largura de banda e tempo de CPU à medida que as mensagens incorporam múltiplos campos, listas e strings.

### 3. Delimitação de Mensagens em Streams TCP

- Sockets TCP oferecem uma abstração de **fluxo contínuo de bytes** (*byte-stream*), sem preservação de fronteiras de mensagem (*message boundaries*).
- No protocolo textual, utilizamos o caractere de nova linha `\n` como delimitador.
- No protocolo Protobuf (onde bytes binários podem conter qualquer valor, inclusive `0x0A` que é `\n`), adotamos a estratégia de **Framing com Prefixo de Comprimento** (*Length-Prefixed Framing*): um prefixo de 4 bytes (*uint32 big-endian*) antecede cada mensagem indicando exatamente quantos bytes binários pertencem àquele payload.
