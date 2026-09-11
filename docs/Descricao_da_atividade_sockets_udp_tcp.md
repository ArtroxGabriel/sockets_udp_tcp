# Atividade: Comunicação com Sockets: UDP vs. TCP

* **Disciplina:** Sistemas Distribuídos / Capítulo 4 (Comunicação entre Processos)
* **Prazo:** Até o dia 20/09 23h59
* **Linguagem:** à sua escolha. Sugestões: Java ou Python ou Go ou Rust ou Javascript ou C/C++

## 1. Objetivo

Implementar o **mesmo serviço cliente-servidor duas vezes**, uma vez sobre **UDP**, outra sobre **TCP**, e comparar empiricamente o comportamento dos dois protocolos diante de perda de mensagens. O trabalho conecta diretamente com o que foi visto no capítulo: a API de sockets, o modelo de requisição-e-resposta e a discussão sobre quando usar cada protocolo.

Ao final, o aluno deve conseguir responder: **"Por que o TCP não perde mensagens e o UDP** **sim (e o que custa resolver isso)?"**

## 2. Cenário: Calculadora remota

Cliente e servidor implementam um protocolo simples de requisição-e-resposta para operações matemáticas:

* O cliente envia requisições no formato CALC:<n>:<operando1>:<op>:<operando2>, onde n é o número de sequência (0, 1, 2, ...), operando1 e operando2 são números (inteiros ou decimais) e op é uma das operações: +, -, *, /.
* O servidor executa a operação e responde com RESULT:<n>:<resultado>.
* Em caso de divisão por zero ou operação inválida, o servidor responde com ERROR:<n>:<mensagem de erro>.
Esse serviço é propositalmente simples: o foco da avaliação está no comportamento de rede, não na lógica de negócio.

### Exemplo de interação

Cliente → CALC:0:10:+:5 Servidor → RESULT:0:15.0

Cliente → CALC:1:8:/:0 Servidor → ERROR:1:divisão por zero

Cliente → CALC:2:3.5:*:2 Servidor → RESULT:2:7.0

## 3. Parte 1: Implementação UDP

* **CalcServerUDP**: recebe datagramas, executa a operação matemática e responde com o resultado.
* **CalcClientUDP**: envia uma sequência de N = 20 requisições de cálculo numeradas (geradas aleatoriamente), uma de cada vez (aguarda a resposta antes de enviar a próxima), e mede o tempo de ida-e-volta (RTT) de cada uma.
* **Perda simulada**: o servidor deve descartar propositalmente uma fração configurável das mensagens recebidas (por exemplo, --loss-rate 0.1 para 10%), sorteando aleatoriamente quais delas "perder" (isto é, recebe, mas não responde). Isso simula, de forma controlada, a falha de omissão que o UDP não trata sozinho.
* **Confiabilidade no cliente**: o CalcClientUDP deve implementar um mecanismo simples de **timeout + retransmissão**: se a resposta não chegar dentro de X ms (sugestão: 500 ms), reenviar a mesma requisição, até um máximo de tentativas (sugestão: 5). Se esgotar as tentativas, registrar a requisição como perdida e continuar.
* O servidor deve suportar **múltiplos clientes concorrentes** sem travar.

## 4. Parte 2: Implementação TCP

* **CalcServerTCP**: aceita conexões e executa operações matemáticas para cada cliente.
* **CalcClientTCP**: mesmo comportamento do cliente UDP (N=20 requisições numeradas, mede RTT), mas **sem lógica de retransmissão** (não deve ser necessária).
* O servidor deve tratar **múltiplos clientes concorrentes**, cada um em sua própria thread.
* Não precisa implementar perda simulada aqui. Note que o TCP entrega tudo, na ordem, sem esse tratamento adicional.

## 5. Parte 3: Experimento e Análise

Execute o cliente **3 vezes contra o servidor UDP**, variando a taxa de perda simulada: **0%, 10%,** **30%**. Depois, execute **uma vez contra o servidor TCP**.

### Para cada execução, registre

* Tempo total da sequência completa;
* RTT médio e RTT máximo;
* Número de retransmissões (apenas UDP); e
* Requisições perdidas definitivamente, se houver (esgotou tentativas).

## 6. Parte 4: Implementação com Protocol Buffers (protobuf)

Nesta parte, você irá reimplementar a calculadora remota usando **Protocol Buffers** como formato de serialização das mensagens, no lugar do protocolo textual das partes anteriores. Você escolhe o formato das mensagens.

### O que implementar

* **CalcServerProto** e **CalcClientProto** usando TCP.

* Use a biblioteca oficial do protobuf para a linguagem escolhida e o compilador protoc para gerar o código a partir do .proto.
* O comportamento deve ser equivalente ao da Parte 2 (TCP, N=20 requisições, mede RTT), mas com as mensagens serializadas em binário pelo protobuf.
* Meça e registre o **tamanho médio das mensagens** (em bytes) e compare com o protocolo textual da Parte 2.

## 7. Requisitos técnicos

* Uso explícito da API de sockets nativa da linguagem (socket/DatagramSocket/Socket em Java, ou o módulo socket em Python). **Não usar bibliotecas de alto nível** que escondam a diferença entre UDP e TCP (ex.: frameworks HTTP prontos).
* Código organizado em pelo menos 6 arquivos/módulos (servidor e cliente para cada parte: UDP, TCP, Proto).
* Tratamento básico de erros (conexão recusada, timeout, divisão por zero, operação inválida, etc.) sem travar o programa.

## 8. Entrega (via GoogleForms)

### Link do formulário: <u>[https://forms.gle/pJw7AR2h2zrNjuX3A](https://forms.gle/pJw7AR2h2zrNjuX3A)</u>

### No formulário, você irá submeter

* O **código-fonte completo** (em arquivo .zip ou link para repositório). Incluir o arquivo **.proto** da Parte 4;
* Um **README.md** com instruções de como compilar/executar cada parte;
* As **respostas às questões elaboradas.**
