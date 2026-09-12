# Sockets UDP e TCP

- **aluno**: Antonio Gabriel
- **matricula**: 539628

<details>
<summary><b>Sumário</b></summary>

- [Estrutura do projeto](#estrutura-do-projeto)
- [Tecnologias Utilizadas](#tecnologias-utilizadas)
- [Como executar](#como-executar)
  - [1. Compilação e Testes](#1-compilação-e-testes)
  - [2. Execução manual das partes (Servidor e Cliente)](#2-execução-manual-das-partes-servidor-e-cliente)
    - [Parte 1: UDP com Perda Simulada e Retransmissão](#parte-1-udp-com-perda-simulada-e-retransmissão)
    - [Parte 2: TCP Concorrente](#parte-2-tcp-concorrente)
    - [Parte 4: Protocol Buffers sobre TCP](#parte-4-protocol-buffers-sobre-tcp)
  - [3. Execução dos Clientes em Bun (TypeScript)](#3-execução-dos-clientes-em-bun-typescript)
  - [4. Experimento Automatizado (Partes 3 e 4)](#4-experimento-automatizado-partes-3-e-4)
  - [5. Execução em Ambiente Conteinerizado (Docker & Docker Compose)](#5-execução-em-ambiente-conteinerizado-docker--docker-compose)
    - [A. Construir a imagem Docker](#a-construir-a-imagem-docker)
    - [B. Subir os servidores em background (UDP, TCP e Protobuf)](#b-subir-os-servidores-em-background-udp-tcp-e-protobuf)
    - [C. Executar clientes conteinerizados](#c-executar-clientes-conteinerizados)
    - [D. Executar a bateria de experimentos dentro do Docker (Multi-Container + netem)](#d-executar-a-bateria-de-experimentos-dentro-do-docker-multi-container--netem)
    - [E. Parar e limpar os containers](#e-parar-e-limpar-os-containers)
- [Descrição da atividade](#descrição-da-atividade)

</details>


## Estrutura do projeto

```
.
├── Taskfile.yml                  # Orquestração de builds, testes e execuções
├── mise.toml                     # Gerenciamento de ferramentas (go, bun, task, protoc, protoc-gen-go)
├── go.mod                        # Módulo Go e dependências
├── docker/                       # Arquivos de conteinerização
│   ├── Dockerfile                # Imagem Alpine multi-stage otimizada com iproute2
│   ├── docker-compose.yml        # Orquestração dos servidores e clientes interativos
│   ├── docker-compose.experiment.yml # Orquestração multi-container com simulação netem (0%, 10%, 30%)
│   └── entrypoint.sh             # Entrypoint com suporte a tc netem
├── proto/
│   └── calc.proto                # Esquema Protocol Buffers (Parte 4)
├── pkg/
│   ├── calc/                     # Lógica de negócio da calculadora e testes unitários
│   │   ├── calc.go
│   │   └── calc_test.go
│   ├── protocol/                 # Parsing textual e framing de stream binário com testes
│   │   ├── text.go
│   │   └── text_test.go
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

Todas as tarefas são orquestradas via **Taskfile**. Para listar todos os comandos disponíveis:

```bash
task
# ou: task --list
```

> [!TIP]
> **Bizu para quem não tem a ferramenta `task` instalada:**
> O arquivo `Taskfile.yml` funciona como um roteiro simples e transparente. Caso não tenha o `task` instalado em seu sistema, basta abrir o [Taskfile.yml](./Taskfile.yml) e executar diretamente no terminal os comandos descritos no campo `cmds` de cada tarefa (por exemplo, `go run ./cmd/server_udp` no lugar de `task run:server:udp`, ou `docker compose -f docker/docker-compose.yml up -d` no lugar de `task docker:up`).

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

### 5. Execução em Ambiente Conteinerizado (Docker & Docker Compose)

Toda a infraestrutura pode ser construída e executada em containers isolados via Docker, simulando tanto perda na aplicação (`--loss-rate`) quanto perda de pacotes no kernel via `tc netem` (`iproute2`):

#### A. Construir a imagem Docker

```bash
task docker:build
# Sem task: docker build -f docker/Dockerfile -t sockets-calc:latest .
```

#### B. Subir os servidores em background (UDP, TCP e Protobuf)

```bash
task docker:up
# Sem task: docker compose -f docker/docker-compose.yml up -d server-udp server-tcp server-proto
```

Para acompanhar os logs e mensagens recebidas no terminal dos servidores:

```bash
task docker:logs
# Sem task: docker compose -f docker/docker-compose.yml logs -f
```

#### C. Executar clientes conteinerizados

```bash
task docker:run:client:udp     # Cliente UDP (com retransmissões na rede do compose)
task docker:run:client:tcp     # Cliente TCP
task docker:run:client:proto   # Cliente Protobuf
```

#### D. Executar a bateria de experimentos dentro do Docker (Multi-Container + netem)

Sobe uma topologia dedicada de 5 containers com servidores independentes (`server-udp-0`, `server-udp-10` com 10% de perda no kernel, `server-udp-30` com 30% de perda no kernel, `server-tcp` e `server-proto`), executa o runner remoto contra a rede virtual e salva o relatório atualizado em `docs/Relatorio_Experimentos.md` via volume montado:

```bash
task docker:experiment
# Sem task:
# docker compose -f docker/docker-compose.experiment.yml up --abort-on-container-exit --exit-code-from experiment
# docker compose -f docker/docker-compose.experiment.yml down
```

#### E. Parar e limpar os containers

```bash
task docker:down
# Sem task: docker compose -f docker/docker-compose.yml down
```

## Descrição da atividade

- [Descrição da atividade](docs/Descricao_da_atividade_sockets_udp_tcp.md)
