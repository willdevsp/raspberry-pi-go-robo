## Purpose
O servidor headless em Go gerencia as operações do Raspberry Pi no carrinho teleoperado, fornecendo streaming de vídeo MJPEG da câmera CSI, comunicação contínua via WebSocket, comandos para a interface Serial e proteção contra perda de sinal.

## ADDED Requirements

### Requirement: Streaming MJPEG da Câmera
O servidor SHALL capturar o feed da Pi Camera usando o `rpicam-vid` e disponibilizá-lo como um stream HTTP (MJPEG multipart).
#### Scenario: Client conecta ao stream
- **WHEN** um cliente faz uma requisição GET em `/stream`
- **THEN** o servidor responde com Content-Type `multipart/x-mixed-replace` e inicia o envio de frames JPEG contínuos.

### Requirement: Conexão Segura e Única no WebSocket
O servidor SHALL aceitar conexões em `/ws` apenas se o PIN for correto e permitir apenas um piloto ativo simultaneamente.
#### Scenario: Cliente tenta conectar sem PIN válido
- **WHEN** um cliente solicita conexão ao `/ws` sem fornecer o PIN '4321'
- **THEN** a conexão é rejeitada pelo servidor.
#### Scenario: Segundo piloto tenta conectar
- **WHEN** já existe uma conexão ativa no `/ws` e um novo cliente autenticado tenta conectar
- **THEN** a segunda conexão é bloqueada até que o piloto ativo se desconecte.

### Requirement: Repasse de Comandos de Direção para a Serial
O servidor SHALL escutar eventos de controle do WebSocket (W, A, S, D, Q, E) e repassar o respectivo byte pela porta serial conectada ao Arduino.
#### Scenario: Comando de avançar é recebido
- **WHEN** o servidor recebe o comando 'W' via WebSocket
- **THEN** o byte 'W' é enviado via `/dev/ttyUSB0` ao Arduino.

### Requirement: Fail-Safe (Watchdog) de Perda de Controle
O servidor SHALL abortar a movimentação (byte 'X') se o fluxo contínuo de controle do cliente for interrompido por mais de 400ms.
#### Scenario: Fluxo de controle é interrompido
- **WHEN** o veículo está em movimento e passam 400ms sem receber pacotes de controle pelo WebSocket
- **THEN** o servidor escreve imediatamente 'X' na porta serial.

### Requirement: Asset Bundling com Frontend
O servidor SHALL embutir os arquivos HTML/JS em sua compilação e servi-los na rota raiz.
#### Scenario: Acesso à interface de controle
- **WHEN** um cliente acessa `GET /`
- **THEN** o arquivo HTML e assets vinculados (embed) são retornados.

### Requirement: Topologia de Rede Access Point
O Raspberry Pi SHALL ser configurado como Access Point autônomo.
#### Scenario: Conexão local de um dispositivo externo
- **WHEN** o Pi inicializa fora da área de alcance de uma infraestrutura
- **THEN** ele disponibiliza o SSID local para conexão direta dos clientes.
