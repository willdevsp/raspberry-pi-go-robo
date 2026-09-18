## Why

O projeto necessita de um sistema embarcado em Go no Raspberry Pi 3B (RoverPi) para atuar como o cérebro de controle e telemetria de um carrinho 4WD. O backend atual precisa transmitir vídeo (Pi Camera) em tempo real via MJPEG com baixa latência e prover uma ponte segura e responsiva (WebSocket para porta Serial) para controlar o microcontrolador Arduino do chassi, sem depender de uma rede de infraestrutura (o Pi será um Access Point autônomo).

## What Changes

- Implementação do servidor headless Go (`rover-server`) para Raspberry Pi OS 64-bit Lite.
- Integração da Pi Camera via subsistema libcamera (`rpicam-vid`) com stream MJPEG sobre HTTP (`multipart/x-mixed-replace`).
- Otimização do particionamento MJPEG usando leitura em blocos e `bytes.Index`.
- Implementação de WebSocket bidirecional para recepção de comandos contínuos de direção (W, A, S, D, Q, E, X), protegido por autenticação via PIN fixo ("4321").
- Fail-safe ativo (Watchdog): parada imediata (envio de 'X') se a comunicação contínua no WebSocket for interrompida por mais de 400ms.
- Encaminhamento dos comandos de direção via interface Serial para o Arduino usando `go.bug.st/serial` a 115200 baud.
- Empacotamento de assets de frontend web via `//go:embed` em um binário unificado.
- Arquitetura focada em injeção de dependências para permitir Black-Box Testing com mocks de câmera e serial (armazenados em diretório `tests/` na raiz).
- Configuração do Raspberry Pi como Access Point (AP) para conexão direta e portabilidade.

## Capabilities

### New Capabilities
- `roverpi-backend`: Implementação completa do servidor Go de telemetria e controle teleoperado do carrinho 4WD (câmera, websocket, serial e fail-safe).

### Modified Capabilities
- (Nenhuma capacidade existente modificada)

## Impact

- **Código:** Criação dos pacotes Go em `cmd/server/`, `internal/camera/`, `internal/controller/`, `internal/transport/`, além de arquivos HTML/JS em `web/static/` e suite de testes em `tests/`.
- **APIs:** Novo endpoint HTTP `GET /stream` (MJPEG) e `GET /ws` (WebSocket de controle).
- **Sistemas / SO:** Exigirá configuração de rede via NetworkManager no Raspberry Pi para subir a rede `RoverPi_Net` (Access Point) e acesso a hardware (`/dev/video*`, `/dev/ttyUSB0`).
