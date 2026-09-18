# PRD — RoverPi: Sistema de Controle Teleoperado e Telemetria em Go

Documento de Requisitos de Produto (PRD) para especificação, arquitetura e implementação do serviço embarcado em Go no **Raspberry Pi 3B** (Raspberry Pi OS 64-bit Lite).

---

## 1. Visão Geral e Escopo

O RoverPi é o núcleo de controle de software do carrinho 4WD. Ele roda diretamente no Raspberry Pi 3B sem interface gráfica (headless), com três responsabilidades centrais:

1. Ingerir o feed da **Pi Camera (CSI)** via subsistema `libcamera`/`rpicam` e servir como stream MJPEG HTTP de baixa latência.
2. Manter conexão persistente full-duplex via **WebSockets** com a interface web do cliente (PC/Navegador).
3. Traduzir eventos de controle recebidos pela rede em comandos seriais síncronos despachados via USB (`/dev/ttyUSB0`) para o microcontrolador **Arduino**.

---

## 2. Arquitetura de Software no Raspberry Pi

```text
  [ Cliente: Navegador Web ]
      │             │
      │ HTTP (8080) │ WS: /ws
      ▼             ▼
┌─────────────────────────────────────────────────────────────┐
│                 Binário Go (rover-server)                   │
│                                                             │
│  ┌────────────────┐  ┌────────────────┐  ┌───────────────┐  │
│  │ Embedded Web   │  │ WebSocket Hub  │  │ Camera Worker │  │
│  │ (go:embed)     │  │ & Watchdog     │  │ (rpicam-vid)  │  │
│  └────────────────┘  └───────┬────────┘  └───────┬───────┘  │
│                              │                   │          │
│                              ▼                   ▼          │
│                      ┌───────────────┐   /stream (MJPEG)    │
│                      │ Serial Worker │                      │
│                      └───────┬───────┘                      │
└──────────────────────────────┼──────────────────────────────┘
                               │ /dev/ttyUSB0 (115200 baud)
                               ▼
                    [ Arduino Uno Clone ]

```

---

## 3. Requisitos Funcionais

### RF-01: Ingestão e Streaming da Pi Camera (CSI)

* O processo Go deve gerenciar a execução do pipeline de vídeo nativo do Raspberry Pi OS 64-bit (`rpicam-vid` ou `libcamera-vid`).
* O vídeo deve ser capturado com resolução de **640x480** a **30 FPS** no formato MJPEG.
* O stream deve ser disponibilizado via endpoint HTTP `GET /stream` (Content-Type: `multipart/x-mixed-replace; boundary=frame`).
* Latência visual de ponta a ponta não deve exceder **120 ms** em rede local.

### RF-02: Endpoint WebSocket de Baixa Latência e Autenticação

* Rota `GET /ws` atualizada via handshake WebSocket (`gorilla/websocket`).
* O servidor deve processar mensagens de entrada instantaneamente sem enfileiramento bloqueante.
* Acesso protegido por PIN fixo (`4321`). O frontend deve enviar este PIN na inicialização da conexão.
* Deve suportar conexão única ativa por segurança operacional (bloquear novas conexões enquanto já houver um piloto ativo autenticado; o piloto atual deve se desconectar para liberar).

### RF-03: Ponte Serial com Arduino (CH340)

* O backend deve abrir comunicação com o descritor serial correspondente ao conversor CH340 do Arduino (geralmente `/dev/ttyUSB0` no Linux).
* Configuração do barramento: **115200 baud**, 8 bits de dados, sem paridade, 1 stop bit (8N1).
* Não deve haver retenção de buffer: cada comando despachado pelo cliente deve gerar escrita imediata de 1 byte no descritor serial.

### RF-04: Fail-Safe Ativo (Watchdog de Conexão)

* O servidor Go deve conter um temporizador de segurança (*dead-man's switch*).
* O frontend é responsável por enviar o comando de movimento em loop contínuo (ex: a cada 100ms) enquanto a tecla for mantida pressionada.
* Se a conexão WebSocket for interrompida bruscamente, o ping falhar ou nenhum comando for recebido por mais de **400 ms** enquanto o veículo estiver em movimento, o Go deve forçar o envio do byte de parada (`'X'`) para o Arduino.

### RF-05: Distribuição em Binário Único (Embedded Assets)

* Todos os arquivos estáticos do frontend (HTML, CSS, JS) devem ser empacotados dentro do binário final utilizando a diretiva `//go:embed`.
* O executável não deve depender de arquivos externos no filesystem do Pi além das permissões de hardware (`/dev/video*`, `/dev/ttyUSB*`).

---

## 4. Requisitos Não-Funcionais

| Métrica | Limite Aceitável | Motivo Técnico |
| --- | --- | --- |
| **Consumo de Memória (RAM)** | < 45 MB | O Pi 3B possui 1 GB compartilhado com a GPU (`gpu_mem`). |
| **Uso de CPU (Processo Go)** | < 15% de 1 núcleo | Deixar núcleos livres para o encoder de vídeo e SO. |
| **Latência de Controle (WS ➔ Serial)** | < 10 ms | Evitar atrasos na parada ou manobras do carrinho. |
| **Arquitetura de Compilação** | `linux/arm64` | Compatibilidade com o Pi OS 64-bit instalado. |
| **Topologia de Rede (SO)** | `Access Point (AP)` | O Pi deve atuar como hotspot Wi-Fi autônomo, provendo rede direta para portabilidade outdoor e menor latência. |

---

## 5. Especificação de Interfaces e Protocolos

### Contrato Serial (Raspberry Pi ➔ Arduino)

Comunicação orientada a caracteres ASCII únicos:

| Byte | Ação do Chassi | Estado dos Pinos da Ponte H |
| --- | --- | --- |
| `'W'` | Frente | E: Avanço | D: Avanço |
| `'S'` | Ré | E: Recuo | D: Recuo |
| `'A'` | Giro à Esquerda | E: Recuo | D: Avanço |
| `'D'` | Giro à Direita | E: Avanço | D: Recuo |
| `'Q'` | Diagonal Frente-Esquerda | E: Parado/Lento | D: Avanço |
| `'E'` | Diagonal Frente-Direita | E: Avanço | D: Parado/Lento |
| `'X'` | Parada Imediata | Todos os pinos em nível lógico LOW |

### Protocolo WebSocket (Cliente ➔ Go)

Cargas em formato texto contendo exatamente o comando de ação:

* `KEYDOWN` no cliente (em loop contínuo): envia `'W'`, `'S'`, `'A'`, `'D'`, `'Q'` ou `'E'` em intervalos regulares enquanto pressionado.
* `KEYUP` no cliente: interrompe o envio contínuo e envia imediatamente `'X'`.
* `PING`/`PONG`: frames de controle padrão RFC 6455 a cada 2 segundos para monitoramento da integridade do enlace.

---

## 6. Integração do Vídeo no Pi OS 64-bit

Como o sistema é a versão Lite (terminal) do Raspberry Pi OS Bookworm/Bullseye 64-bit, a captura direta via V4L2 tradicional é substituída pela API libcamera. O Go atuará como supervisor do processo `rpicam-vid`:

```bash
# Comando base encapsulado pelo pacote de câmera em Go:
rpicam-vid -t 0 --inline --width 640 --height 480 --framerate 30 --codec mjpeg -o -

```

O Go captura o `stdout` desse subprocesso via `io.Reader`, particiona o fluxo nos marcadores JPEG (SOI `0xFFD8` e EOI `0xFFD9`) e despacha para os clientes HTTP conectados ao endpoint `/stream`. 
*Nota de Performance:* Para respeitar o limite de CPU (< 15%), o particionamento não deve ler os dados byte a byte em um loop manual. Deve-se realizar leituras em blocos maiores (chunks) usando a função nativa `bytes.Index()` para localizar as fronteiras das imagens de forma otimizada.

---

## 7. Estrutura do Projeto Go e Testabilidade (Mocking)

Para suportar a validação da pipeline de CI estabelecida (`gitflow.md`), o projeto deve fazer uso de injeção de dependência e interfaces (ex: abstraindo a Porta Serial para um `io.ReadWriteCloser`). Para garantir que o código funcional fique 100% isolado, todos os testes e implementações fictícias (Mocks) devem ser alocados em uma pasta separada `tests/` na raiz do repositório (Padrão de Black-box testing).

```text
roverpi/
├── cmd/
│   └── server/
│       └── main.go           # Ponto de entrada, injeção de dependências e flags
├── internal/
│   ├── camera/
│   │   └── streamer.go       # Gestão do rpicam-vid e HTTP MJPEG (implementa interface)
│   ├── controller/
│   │   └── serial.go         # Conexão serial genérica via abstração io.ReadWriteCloser
│   └── transport/
│       └── websocket.go      # Handshake, loop de leitura, watchdog e fail-safe
├── tests/                    # Diretório isolado contendo apenas código de validação CI
│   ├── unit/                 # Casos de teste unitários (camera_test.go, serial_test.go, etc)
│   └── mocks/                # Implementações fakes (CameraMock, SerialMock) para testes sem hardware
├── web/
│   ├── static/
│   │   └── app.js            # Keydown/Keyup listeners e conexão WebSocket
│   └── index.html            # Interface limpa com stream e status de conexão
├── go.mod
└── go.sum

```

---

## 8. Procedimento de Build e Deploy

Compilação cruzada direta a partir do computador de desenvolvimento para o Pi OS 64-bit:

```bash
# No computador de desenvolvimento:
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o rover-server ./cmd/server

# Cópia do binário para o Raspberry Pi:
scp rover-server pi@<IP_DO_RASPBERRY>:/home/pi/

# Execução no Raspberry Pi:
./rover-server -port=8080 -serial=/dev/ttyUSB0

```

---

## 9. Critérios de Aceite

* O comando `curl -I http://<IP_DO_PI>:8080/stream` retorna cabeçalho `multipart/x-mixed-replace` e exibe imagem em tempo real sem travamentos.
* Abertura do painel web em navegador desktop carrega imagem e status `Conectado`.
* O acionamento contínuo da tecla `W` mantém o comando ativo; a liberação da tecla envia `'X'` e o chassi interrompe o movimento em menos de 50 ms.
* Desconexão física do cabo de rede ou fechamento da aba do navegador resulta no envio automático do comando `'X'` para o Arduino em no máximo 400 ms.