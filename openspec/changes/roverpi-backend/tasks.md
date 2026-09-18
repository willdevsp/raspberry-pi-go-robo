## 1. Estrutura do Projeto e Mocks

- [ ] 1.1 Inicializar o módulo Go (`go mod init`) e criar a estrutura de diretórios (`cmd/server`, `internal/camera`, `internal/controller`, `internal/transport`, `tests/unit`, `tests/mocks`, `web/static`). Verificar se todos os diretórios existem.
- [ ] 1.2 Instalar dependências externas (`gorilla/websocket`, `go.bug.st/serial`). Verificar se estão no `go.mod`.
- [ ] 1.3 Criar a interface `Streamer` e seu respectivo `CameraMock` na pasta `tests/mocks`, com testes unitários básicos retornando dummy frames. Verificar com `go test ./tests/unit/...`.
- [ ] 1.4 Criar a interface `io.ReadWriteCloser` (Serial) e seu respectivo `SerialMock` na pasta `tests/mocks`. Verificar com `go test ./tests/unit/...`.

## 2. Lógica da Câmera (Streamer)

- [ ] 2.1 Implementar a struct real em `internal/camera/streamer.go` que execute `rpicam-vid` via `exec.Command` e redirecione stdout. Verificar com teste mockando a execução de comando ou instanciando localmente sem falhas de sintaxe.
- [ ] 2.2 Implementar o parser de multipart MJPEG usando `bufio.Scanner` e `bytes.Index` com os headers HTTP adequados e boundary. Verificar a extração dos bytes com testes unitários usando um buffer falso simulando frames.

## 3. Lógica do Controlador Serial

- [ ] 3.1 Implementar a conexão via `go.bug.st/serial` em `internal/controller/serial.go` encapsulada pela interface genérica de ReadWriteCloser. Verificar configurando a serial mockada em um teste unitário e asserindo que comandos gravam nela.

## 4. Transporte e Websocket (Fail-safe)

- [ ] 4.1 Criar a configuração e o upgrade de conexão WebSocket em `internal/transport/websocket.go`. Verificar escrevendo um teste de Handler do net/http httptest validando o erro para PIN incorreto e sucesso para '4321'.
- [ ] 4.2 Implementar o gerenciador de sessão única (bloqueio de conexões extras). Verificar conectando dois WebSockets falsos no teste unitário, asserindo recusa do segundo.
- [ ] 4.3 Implementar a rotina de recepção de comandos (W, A, S, D, Q, E) que retransmite ao controlador Serial. Verificar com testes unitários que a letra exata chegou no `SerialMock`.
- [ ] 4.4 Implementar o Watchdog (fail-safe). Adicionar um `time.Timer` de 400ms por conexão. Verificar asserindo em testes unitários que na falta de tráfego, o pacote 'X' (parada) é gerado no `SerialMock` após 400ms.

## 5. Frontend (go:embed)

- [ ] 5.1 Codificar o arquivo `web/index.html` e `web/static/app.js` incluindo `setInterval` para os keydowns (W,A,S,D,Q,E) a 100ms e keyup ('X'). Verificar manualmente abrindo o index.html no browser e checando os logs.
- [ ] 5.2 Embutir os arquivos no root mux do Go usando `//go:embed`. Verificar se um client HTTP de teste `GET /` acessa o HTML corretamente sem arquivos externos no disco.

## 6. Integração do Servidor

- [ ] 6.1 Unir Câmera, Controlador, Transport e Web no arquivo `cmd/server/main.go`. Verificar compilando o servidor (`go build ./cmd/server`) sem erros.
- [ ] 6.2 Desenvolver a documentação de script/bash com `nmcli` (Raspberry Pi AP) e instruções de deploy. Verificar leitura visual do README.
