## Context

O ambiente de implantação é um Raspberry Pi 3B rodando Raspberry Pi OS 64-bit Lite (Headless). Por possuir recursos limitados (1GB de RAM compartilhado com a GPU) e restrição rigorosa de CPU definida em requisitos, a extração de frames da saída do pipeline de câmera (`rpicam-vid`) requer atenção. Além disso, a aplicação usará a biblioteca `go.bug.st/serial` para comunicação com um Arduino Uno usando chip CH340 (`/dev/ttyUSB0`), sendo as instruções encapsuladas em Mocks (`tests/mocks/`) nas execuções de testes em ambiente CI (GitHub Actions).

## Goals / Non-Goals

**Goals:**
- Prover um executável monolítico leve em Go usando módulos nativos de I/O eficientes para o parser de vídeo.
- Estabelecer uma estrutura de injeção de dependência via interfaces (`io.ReadWriteCloser`, `camera.Streamer`) para testabilidade isolada.
- Criar a mecânica de verificação contínua no WebSocket que reinicie um ticker temporizador no lado servidor para cada evento recebido, cortando o sinal em 400ms na ausência de pacotes (fail-safe).

**Non-Goals:**
- Não há controle de usuários complexo ou base de dados, a autenticação restringe-se a uma comparação estática de token (PIN).
- Não haverá suporte a protocolo WebRTC devido à complexidade da pilha em contrapartida ao simples MJPEG que atende à LAN com eficiência.

## Decisions

- **Parser de MJPEG**: O Go lerá a saída (stdout) do `rpicam-vid` usando buffers de chunk `bufio.Scanner` combinados com `bytes.Index()` em memória para achar as fronteiras dos JPEG (0xFFD8, 0xFFD9). 
  - *Alternativa*: Ler byte a byte ou usar parsers complexos de H264. O uso de `bytes.Index` com bytes pre-alocados diminui varreduras de array para `O(N)` nativo C e é mais escalável sob a restrição de <15% de CPU.
- **Topologia Access Point**: Ao invés de clientes Wi-Fi em um roteador de terceiros, rodaremos o serviço nativo do sistema do Pi (como o NetworkManager ou HostAPD) para provisionar o SSID `RoverPi_Net`.
- **Implementação do Watchdog**: Utilizar `time.Timer` na struct do Client WebSocket que é "resetado" a cada pacote recebido, disparando um goroutine para enviar o byte de parada ('X') quando expirar.
  - *Alternativa*: Enviar timestamp do lado do cliente para cálculo de delay. Descartado, pois introduz overhead de dessincronização de relógios.
- **Isolamento de Testes (Black-Box Testing)**: Todo o código de Mock e os arquivos test (*_test.go) serão armazenados externamente à `/internal` na pasta `/tests`, promovendo uma separação estrita.
  
## Risks / Trade-offs

- **Risk: Limitação de Bandwidth local** -> O MJPEG trafega grandes volumes de dados (cada frame é uma foto full).
  - *Mitigation*: Restrição explícita do FPS para 30 e resolução VGA 640x480 na chamada do comando `rpicam-vid`.
- **Risk: Disparos falsos de Fail-safe** -> Se a thread agendadora do Go demorar a receber mensagens, o carro pode falhar durante movimento contínuo.
  - *Mitigation*: Utilizar um temporizador (100-200ms) alto no client-side via JavaScript `setInterval` forçando o buffer do socket a receber bytes frequentemente em pequenos pacotes textuais.
