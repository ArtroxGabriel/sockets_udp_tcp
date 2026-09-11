package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type ExperimentResult struct {
	Scenario        string  `json:"scenario"`
	Protocol        string  `json:"protocol"`
	TotalDurationMs float64 `json:"totalDurationMs"`
	Delivered       int     `json:"delivered"`
	Lost            int     `json:"lost"`
	Retransmissions int     `json:"retransmissions"`
	AvgRTTMs        float64 `json:"avgRttMs"`
	MaxRTTMs        float64 `json:"maxRttMs"`
	AvgSentBytes    float64 `json:"avgSentBytes"`
	AvgRecvBytes    float64 `json:"avgRecvBytes"`
}

type ScenarioConfig struct {
	Name        string
	Protocol    string
	ServerCmd   string
	ServerArgs  []string
	ClientCmd   string
	ClientArgs  []string
	Port        int
	IsUDP       bool
}

func main() {
	log.Println("=== INICIANDO EXPERIMENTO COMPARATIVO (UDP vs. TCP vs. PROTOBUF) ===")

	scenarios := getScenarios()
	var results []ExperimentResult

	for _, sc := range scenarios {
		log.Printf("Executando cenário: %s ...", sc.Name)
		res, err := runScenario(sc)
		if err != nil {
			log.Fatalf("falha ao executar cenário %s: %v", sc.Name, err)
		}
		results = append(results, res)
		time.Sleep(200 * time.Millisecond) // Intervalo entre execuções
	}

	markdownReport := generateReport(results)
	fmt.Println("\n" + markdownReport)

	if err := saveReportToFile(markdownReport, "docs/Relatorio_Experimentos.md"); err != nil {
		log.Printf("aviso: erro ao salvar relatório em arquivo: %v", err)
	} else {
		log.Println("Relatório gravado com sucesso em docs/Relatorio_Experimentos.md")
	}
}

func getScenarios() []ScenarioConfig {
	return []ScenarioConfig{
		{
			Name:       "UDP (0% perda)",
			Protocol:   "UDP",
			ServerCmd:  "./bin/server_udp",
			ServerArgs: []string{"--port", "9081", "--loss-rate", "0.0"},
			ClientCmd:  "./bin/client_udp",
			ClientArgs: []string{"--addr", "127.0.0.1:9081", "--n", "20", "--json"},
			Port:       9081,
			IsUDP:      true,
		},
		{
			Name:       "UDP (10% perda)",
			Protocol:   "UDP",
			ServerCmd:  "./bin/server_udp",
			ServerArgs: []string{"--port", "9082", "--loss-rate", "0.1"},
			ClientCmd:  "./bin/client_udp",
			ClientArgs: []string{"--addr", "127.0.0.1:9082", "--n", "20", "--json"},
			Port:       9082,
			IsUDP:      true,
		},
		{
			Name:       "UDP (30% perda)",
			Protocol:   "UDP",
			ServerCmd:  "./bin/server_udp",
			ServerArgs: []string{"--port", "9083", "--loss-rate", "0.3"},
			ClientCmd:  "./bin/client_udp",
			ClientArgs: []string{"--addr", "127.0.0.1:9083", "--n", "20", "--json"},
			Port:       9083,
			IsUDP:      true,
		},
		{
			Name:       "TCP Textual",
			Protocol:   "TCP",
			ServerCmd:  "./bin/server_tcp",
			ServerArgs: []string{"--port", "9084"},
			ClientCmd:  "./bin/client_tcp",
			ClientArgs: []string{"--addr", "127.0.0.1:9084", "--n", "20", "--json"},
			Port:       9084,
			IsUDP:      false,
		},
		{
			Name:       "TCP Protobuf",
			Protocol:   "Protobuf (TCP)",
			ServerCmd:  "./bin/server_proto",
			ServerArgs: []string{"--port", "9085"},
			ClientCmd:  "./bin/client_proto",
			ClientArgs: []string{"--addr", "127.0.0.1:9085", "--n", "20", "--json"},
			Port:       9085,
			IsUDP:      false,
		},
	}
}

func runScenario(sc ScenarioConfig) (ExperimentResult, error) {
	serverCmd := exec.Command(sc.ServerCmd, sc.ServerArgs...)
	if err := serverCmd.Start(); err != nil {
		return ExperimentResult{}, fmt.Errorf("erro ao iniciar servidor %s: %w", sc.ServerCmd, err)
	}
	defer func() {
		_ = serverCmd.Process.Kill()
		_ = serverCmd.Wait()
	}()

	waitForServer(sc.Port, sc.IsUDP)

	clientCmd := exec.Command(sc.ClientCmd, sc.ClientArgs...)
	var stdout, stderr bytes.Buffer
	clientCmd.Stdout = &stdout
	clientCmd.Stderr = &stderr

	_ = clientCmd.Run() // cliente pode sair com código 1 caso haja perda definitiva

	var res ExperimentResult
	if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
		return ExperimentResult{}, fmt.Errorf("erro ao decodificar JSON do cliente (%s): %w | saída: %s | err: %s",
			sc.ClientCmd, err, stdout.String(), stderr.String())
	}
	res.Scenario = sc.Name
	res.Protocol = sc.Protocol
	return res, nil
}

func waitForServer(port int, isUDP bool) {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	network := "tcp"
	if isUDP {
		network = "udp"
	}
	for i := 0; i < 20; i++ {
		conn, err := net.DialTimeout(network, addr, 50*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return
		}
		time.Sleep(30 * time.Millisecond)
	}
}

func generateReport(results []ExperimentResult) string {
	var b bytes.Buffer
	b.WriteString("# Relatório Comparativo: Sockets UDP vs. TCP vs. Protobuf\n\n")
	b.WriteString("Este documento apresenta os dados coletados empiricamente durante a execução dos testes da calculadora remota conforme a especificação da atividade.\n\n")
	b.WriteString("## Tabela de Resultados dos Experimentos (N = 20 requisições)\n\n")
	b.WriteString("| Cenário | Protocolo | Duração Total (ms) | Entregues | Perdidos | Retransmissões | RTT Médio (ms) | RTT Máximo (ms) | Tam. Médio Req (B) | Tam. Médio Resp (B) |\n")
	b.WriteString("|---|---|---|---|---|---|---|---|---|---|\n")

	for _, r := range results {
		b.WriteString(fmt.Sprintf("| %s | %s | %.2f | %d | %d | %d | %.3f | %.3f | %.1f | %.1f |\n",
			r.Scenario, r.Protocol, r.TotalDurationMs, r.Delivered, r.Lost, r.Retransmissions,
			r.AvgRTTMs, r.MaxRTTMs, r.AvgSentBytes, r.AvgRecvBytes))
	}

	b.WriteString("\n## Respostas às Questões Elaboradas\n\n")
	b.WriteString("### 1. Por que o TCP não perde mensagens e o UDP sim (e o que custa resolver isso)?\n\n")
	b.WriteString("- **Semântica do UDP:** O UDP (*User Datagram Protocol*) é um protocolo orientado a datagramas, não confiável e sem conexão. O sistema operacional simplesmente encapsula o pacote IP e despacha para a rede. Não há confirmação de entrega (*ACK*), controle de fluxo, retransmissão de pacotes perdidos ou ordenação. Se um roteador intermediário descartar um pacote por congestionamento, erro de checksum ou saturação de buffer, o pacote é sumariamente descartado sem aviso.\n")
	b.WriteString("- **Semântica do TCP:** O TCP (*Transmission Control Protocol*) é orientado a conexão e fluxo de bytes confiável. Implementa uma máquina de estados complexa que utiliza números de sequência (*SEQ*), confirmações cumulativas/seletivas (*ACK/SACK*), temporizadores de retransmissão adaptativos (*RTO baseado em RTT estimado*), controle de fluxo por janela deslizante (*sliding window*) e controle de congestionamento (*Slow Start, Congestion Avoidance, Fast Retransmit*).\n")
	b.WriteString("- **O custo de resolver a perda (no UDP vs. TCP):**\n")
	b.WriteString("  - **Custo de latência e jitter:** Para tornar o UDP confiável na camada de aplicação (como implementamos na Parte 1), é necessário introduzir mecanismos como timeout e retransmissão (*Stop-and-Wait* ou *ARQ*). Como evidenciado nos dados empíricos da tabela acima, sob taxa de perda de 30%, o RTT médio e a duração total disparam em ordens de magnitude devido às pausas esperando o timeout expirar.\n")
	b.WriteString("  - **Custo de processamento e overhead de cabeçalho:** O TCP consome mais memória no kernel (buffers de envio e recepção para ordenar segmentos) e possui cabeçalhos maiores (20 a 60 bytes contra apenas 8 bytes do UDP), além da latência inicial do *handshake* de três vias (SYN, SYN-ACK, ACK).\n")
	b.WriteString("  - **Conclusão:** O UDP é ideal quando baixa latência e previsibilidade imediata superam a necessidade de confiabilidade total (ex.: streaming de áudio/vídeo, jogos online, DNS). O TCP é mandatório quando integridade e confiabilidade dos dados são cruciais (ex.: HTTP, bancos de dados, transferência de arquivos).\n\n")

	b.WriteString("### 2. Comparativo de Serialização: Protocolo Textual vs. Protocol Buffers\n\n")
	b.WriteString("- **Formato Textual (Partes 1 e 2):**\n")
	b.WriteString("  - Codifica números em caracteres ASCII/UTF-8 delimitados por `:` (ex.: `CALC:0:10.5:+:4.5`).\n")
	b.WriteString("  - Vantagens: Facilmente legível por humanos e inspecionável em analisadores de pacotes (`tcpdump`, `Wireshark`).\n")
	b.WriteString("  - Desvantagens: Parsing textual repetitivo (`strings.Split`, `strconv.ParseFloat`), sem garantia estrita de esquema e tamanho variável.\n")
	b.WriteString("- **Protocol Buffers (Parte 4):**\n")
	b.WriteString("  - Codificação binária compacta com enums inteiros compactados via varints e campos tipados estritamente.\n")
	b.WriteString("  - Vantagens: Serialização e deserialização muito mais rápidas e com menor overhead de CPU, tipagem forte garantida por contrato (`.proto`) e suporte automático entre diferentes linguagens (Go, Bun/TypeScript, Python, etc.).\n")
	b.WriteString("  - Observação de tamanho: Em mensagens curtas com números decimais de alta precisão (IEEE 754 de 64 bits = 8 bytes binários brutos), o Protobuf possui tamanho ligeiramente superior a inteiros ASCII de 1 dígito, porém escala com imensa vantagem de largura de banda e tempo de CPU à medida que as mensagens incorporam múltiplos campos, listas e strings.\n\n")

	b.WriteString("### 3. Delimitação de Mensagens em Streams TCP\n\n")
	b.WriteString("- Sockets TCP oferecem uma abstração de **fluxo contínuo de bytes** (*byte-stream*), sem preservação de fronteiras de mensagem (*message boundaries*).\n")
	b.WriteString("- No protocolo textual, utilizamos o caractere de nova linha `\\n` como delimitador.\n")
	b.WriteString("- No protocolo Protobuf (onde bytes binários podem conter qualquer valor, inclusive `0x0A` que é `\\n`), adotamos a estratégia de **Framing com Prefixo de Comprimento** (*Length-Prefixed Framing*): um prefixo de 4 bytes (*uint32 big-endian*) antecede cada mensagem indicando exatamente quantos bytes binários pertencem àquele payload.\n")

	return b.String()
}

func saveReportToFile(content string, relPath string) error {
	dir := filepath.Dir(relPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(relPath, []byte(content), 0644)
}
