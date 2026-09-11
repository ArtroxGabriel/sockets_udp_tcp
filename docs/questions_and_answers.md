# Questões Elaboradas

## Parte 1 — UDP

1. **Dados das execuções UDP — grade com linhas 0%, 10%, 30% e colunas: Tempo total (s) · RTT médio (ms) · Retransmissões · Perdidas definitivamente**

    | Taxa de Perda | Tempo total | RTT médio | Retransmissões | Perdidas definitivamente |
    | --------------- | --------------- | --------------- | --------------- | --------------- |
    | 0% | 1.464755 ms | 64.454 µs | 0 | 0 |
    | 10% | 1.0047391 s | 50.220251 ms | 2 | 0 |
    | 30% | 7.0136952 s | 237.39148 ms | 13 | 1 |

1. **Descreva como você implementou o timeout + retransmissão no cliente UDP. O que acontece quando o número máximo de tentativas é esgotado? O cliente consegue detectar sozinho que uma requisição foi perdida?**

    > Não tem como o cliente detectar sozinho que uma requisição foi peridida, como no exercicio ele espera uma reposta do servidor, se ela nao vier apos um tempo, é reenvido a requisição, isso ocorre um numero determinado de vezes. Como o protocolo é UDP, entao ficou a cargo da implementacao do cliente essa retransmissoes e timeout.

## Parte 2 — TCP

* Dados da execução TCP: Tempo total (s) e RTT médio (ms)

  * Tempo total: 1.727831 ms
  * RTT Medio: 70.191µs

* Como o servidor TCP trata múltiplos clientes?

    > Como  minha implementacao foi feita em go, eu crio um listener e digo para aceitar conexoes, para cada conexao executo o handleConnection em uma goroutine, isso deixa transparente para o desenvolvedor em tratar explicitamente multiplos clientes, o proprio go faz isso para mim. clientes.
    >
    > Para cada conexao nova, o servidor gerava um nova net.Conn, em minha implementacao do server_udp, eu fiz isso com um loop infinito para cada nova conexao, no caso do servidor TCP, a conexao é estabelcida e mantida, dessa forma o servidor fica escutando a mesma conexao, enquanto o cliente UDP, a cada nova requisição, ele cria uma nova conexao e fecha ela apos receber a resposta do servidor. Portanto, para cada cliente TCP, o servidor tera a conexao aberta.

* Como o servidor TCP trata múltiplos clientes?
  * [x] Thread por cliente (goroutines)
  * [ ] Async
  * [ ] Processos
  * [ ] Sem concorrência

* Por que o cliente TCP não precisou de timeout ou retransmissão, enquanto o cliente UDP precisou? Compare os tempos e RTTs das duas implementações com perda 0%.

    > O protocolo TCP por ser confiavel, trata e faz o trabalho de timeout e retransmissoes, enquanto o UDP nao, por ser um protocolo nao confiavel, nao tem isso "built-in", necessitando uma implementacao/tratamento por parte do desenvolvedor.
    >
    > por nao ter essa confiabilidade o protocolo UDP possui um menor overhead, o que faz com que ele seja mais rapido, como pdemos ver nos tempos e RTTs das duas implementacoes, o UDP com perda 0% teve um tempo total de 1.464755 ms e RTT medio de 64.454 µs, enquanto o TCP teve um tempo total de 1.727831 ms e RTT medio de 70.191 µs. Ambos muito proximos por se tratar de um cenario local com perda 0, num cenario em servidor e cliente geograficamente distante, a diferenca seria maior.

## Parte 3 — Análise

* O que acontece com as requisições quando o cliente UDP roda sem retransmissão? Com ela ativada, o que muda — e a que custo em tempo? Apoie com os dados coletados.

    > sem a retransmissao, o cliente nao tem garantia de sua requisicao, na vdd, mesmo sem ela nao teria, pq o servidor nao tem obrigacao pelo protocolo de informar, com ela ativada, temos o impacto do timeout, isto é, tempo que o cliente deve esperar para receber a resposta do servidor antea de considerar uma retransmissao, ou seja, o custo fica timeout mais um tempo de uma nova retransmisao, mas isso tudo vezes o numero de tentativas de retransmissoes.

* Dado que UDP exige todo esse trabalho extra para ser confiável, por que ele ainda é usado em sistemas reais? Cite pelo menos um exemplo concreto e justifique.

    > simples, por que ele é rapido, existem cenarios que mesmo sem essa confiabilidade mas graça a velocidade o seu uso é ideal, como por exemplo streaming de video, a perda de um pacote(datagrama) de um frame, nao vai impactar, ou ser perceptivel num todo, logo um UDP é o ideal para esse tipo de aplicacao, tambem existem implementacao em cima do UDP, como o QUIC da google, que o torna um pouco mais confiavel mas ainda assim mais rapido que o TCP.

## Parte 4 — Protobuf

* Comparar tamanho médio das mensagens. Texto puro em bytes (parte 2) x Protobuf em bytes (parte 4):

    | tipo de mensagem | tam. med. request | tam. med. response |
    | --------------- | --------------- | --------------- |
    | texto puro | 16.1 | 15.4 |
    | protobuf | 21.4 | 13.4 |

* Como você mediu o tamanho das mensagens?

    > em cenario de servidor UDP ou TCP, como se tratar de string, apenas extrai o tamanho de uma string de requisicao e response, por fim peguei uma media do total. ja no servidor protobuff ocorreu semelhante, mas inves de um length de uma string, peguei o tamanho de um array de bytes, que é o que o protobuf gera, e peguei a media do total.

* Em qual cenário real você escolheria protobuf em vez de texto simples? Foi mais fácil ou mais difícil de implementar? E por quê?

    > no cenario abordado nessa atividade, eu manteria a implementacao de um texto simples, pois foi mais simples de implementar, tao rapida quanto um protobuf e mais leve, esse ultimo ponto ocorreu pois o formato abordado e o tipo de mensagem enviada nao foi um que favorece o uso de protobuf, entao nos meus testes, ele consumiu mais memoria que o texto simples. mas em um cenario com mensagens mais complexas, com muita strings e dados complexos, o protobuf seria melhor, pois seria mais eficiente, rapido e leve.
    >
    > sua implementacao foi simples, pois utilizei o pacote gerado pelo protoc-gen-go, mas poderia ter mapeado para um DTO da aplicacao para manter uma desacoplamento, de qualquer forma, a implementacao foi simples e facilitada gracas ao protoc, e de certa forma, facilitou, mais que a string, pois nela tive que definir manualmente o processo de leitura e escrita de mensagens, enquanto no protobuf, o pacote gerado faz isso por mim
