# Sockets UDP e TCP

- **aluno**: Antonio Gabriel
- **matricula**: 539628

## Estrutura do projeto

```
.
├── Taskfile.yml                  # Orquestração de builds, testes e execuções
├── mise.toml                     # Gerenciamento de ferramentas (go, bun, task, protoc, protoc-gen-go)
├── go.mod                        # Módulo Go e dependências
├── proto/
│   └── calc.proto                # Esquema Protocol Buffers (Parte 4)
├── pkg/
│   ├── calc/                     # Lógica de negócio da calculadora e testes unitários
│   │   ├── calc.go
│   │   └── calc_test.go
│   ├── protocol/                 # Parsing textual e framing de stream binário com testes
│   │   ├── text.go
│   │   ├── text_test.go
│   │   ├── framing.go
│   │   └── framing_test.go
│   └── pb/calc/                  # Código Go gerado pelo protoc
│       └── calc.pb.go
├── cmd/
│   ├── server_udp/main.go        # CalcServerUDP (com simulação de perda configurável)
│   ├── client_udp/main.go        # CalcClientUDP (timeout + retransmissão)
│   ├── server_tcp/main.go        # CalcServerTCP (concorrente multi-cliente)
│   ├── client_tcp/main.go        # CalcClientTCP (streaming TCP sem perdas)
│   ├── server_proto/main.go      # CalcServerProto (TCP com serialização Protobuf)
│   ├── client_proto/main.go      # CalcClientProto (medição de tamanho e RTT)
│   └── experiment/main.go        # Executor automatizado dos experimentos da Parte 3 e 4
├── bun/                          # Clientes adicionais em Bun (TypeScript) para validação cruzada
│   ├── package.json
│   ├── client_udp.ts             # Cliente UDP em Bun (node:dgram)
│   ├── client_tcp.ts             # Cliente TCP em Bun (node:net)
│   └── client_proto.ts           # Cliente Protobuf em Bun (protobufjs + framing TCP)
└── docs/
    ├── Descricao_da_atividade_sockets_udp_tcp.md
    └── Relatorio_Experimentos.md # Relatório comparativo com tabela de métricas e respostas
```

## Tecnologias utilizadas

- **Go v1.27.1** (sockets de baixo nível via `net`, sem frameworks de alto nível)
- **Bun v1.4.2** (clientes TypeScript para interoperabilidade entre linguagens)
- **Taskfile v3.53.1** (automação de tarefas)
- **Protoc v36.1** + **protoc-gen-go v1.36.12** (Protocol Buffers v3)

## Como executar

Todas as tarefas são gerenciadas pelo **Taskfile**. Para listar os comandos disponíveis, use `task --list`.

### 1. Compilação e Testes

```bash
# Executar todos os testes unitários Go com cobertura de código
task test

# Compilar todos os binários para bin/
task build
```

### 2. Execução manual das partes (Servidor e Cliente)

#### Parte 1: UDP com Perda Simulada e Retransmissão
Em um terminal, inicie o servidor (parâmetro `--loss-rate` aceita valores de `0.0` a `1.0`):
```bash
task run:server:udp -- --port 8081 --loss-rate 0.3
```
Em outro terminal, execute o cliente:
```bash
task run:client:udp -- --addr 127.0.0.1:8081 --n 20 --timeout 500ms --max-attempts 5
```

#### Parte 2: TCP Concorrente
Em um terminal:
```bash
task run:server:tcp -- --port 8082
```
Em outro terminal:
```bash
task run:client:tcp -- --addr 127.0.0.1:8082 --n 20
```

#### Parte 4: Protocol Buffers sobre TCP
Em um terminal:
```bash
task run:server:proto -- --port 8083
```
Em outro terminal:
```bash
task run:client:proto -- --addr 127.0.0.1:8083 --n 20
```

### 3. Execução dos Clientes em Bun (TypeScript)

Para testar a interoperabilidade entre diferentes linguagens:
```bash
task run:bun:udp     # Cliente Bun UDP contra CalcServerUDP
task run:bun:tcp     # Cliente Bun TCP contra CalcServerTCP
task run:bun:proto   # Cliente Bun Protobuf contra CalcServerProto
```

### 4. Experimento Automatizado (Partes 3 e 4)

Executa toda a bateria de testes requerida (UDP 0%, 10%, 30%, TCP Textual e TCP Protobuf), mede os tempos, perdas, retransmissões e tamanhos de mensagem, gerando a tabela comparativa:
```bash
task experiment
```

Os resultados detalhados e as respostas às perguntas do trabalho estão disponíveis em:
- [Relatório Comparativo de Experimentos](docs/Relatorio_Experimentos.md)

## Descrição da atividade

- [Descrição da atividade](docs/Descricao_da_atividade_sockets_udp_tcp.md)
