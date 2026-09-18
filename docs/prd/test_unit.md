# Plano de Testes Unitários — RoverPi

Esta documentação define a estratégia, a arquitetura de testabilidade, os casos de teste e a implementação dos testes unitários para o serviço em Go no Raspberry Pi, garantindo isolamento total de hardware físico (Arduino, portas seriais e câmera CSI).

---

## 1. Estratégia de Testabilidade e Inversão de Dependências

Como o código roda em ambiente embarcado, nenhum teste unitário deve depender de dispositivos `/dev/*` presentes no sistema operacional. Para permitir isolamento e mock rápido, os pacotes dependem de interfaces:

* **Abstração Serial:** O transmissor não escreve diretamente na porta física; ele interage com uma interface que implementa escrita de comandos.
* **Abstração de Vídeo:** O processador de frames consome um `io.Reader` mockado em memória em vez do processo `rpicam-vid`.
* **Abstração Temporal (Watchdog):** O fail-safe opera com canais ou intervalos configuráveis via injeção para testes sem atrasos artificiais longos.

---

## 2. Matriz de Cobertura e Casos de Teste

| Pacote | Função / Componente | Cenário de Teste | Resultado Esperado |
| --- | --- | --- | --- |
| `controller` | `SendCommand(b byte)` | Envio de comandos válidos (`W`, `S`, `A`, `D`, `X`) | Byte gravado com sucesso no buffer |
| `controller` | `SendCommand(b byte)` | Envio de caractere inválido (ex: `Z`, `1`, `\n`) | Retorno de `ErrInvalidCommand`, sem escrita |
| `controller` | `SendCommand(b byte)` | Chamadas concorrentes simultâneas | Mutex garante escrita sequencial sem corrupção |
| `transport` | `HandleMessage(payload)` | Mensagem WebSocket válida (`W`) | Comando repassado ao `SerialTransmitter` |
| `transport` | `Watchdog / Fail-Safe` | Inatividade superior ao threshold (ex: 400ms) | Disparo automático do byte `'X'` |
| `transport` | `Watchdog / Reset` | Comandos recebidos dentro da janela de tempo | Timer resetado; comando `'X'` não é emitido |
| `transport` | Desconexão do Cliente | Fechamento da conexão WS com veículo em movimento | Interrupção imediata via emissão do byte `'X'` |
| `camera` | `ExtractFrames(reader)` | Leitura de fluxo contendo delimitadores JPEG | Detecção correta de início (`0xFFD8`) e fim (`0xFFD9`) |
| `camera` | `ExtractFrames(reader)` | Fluxo corrompido ou stream truncado | Descarte gracioso sem pânico de runtime |

---

## 3. Estrutura de Interfaces e Mocks

### Interface do Controlador Serial (`internal/controller/controller.go`)

```go
package controller

import "errors"

var ErrInvalidCommand = errors.New("comando invalido")

type SerialDevice interface {
	Write(p []byte) (n int, err error)
}

type CommandController struct {
	dev SerialDevice
}

func NewCommandController(dev SerialDevice) *CommandController {
	return &CommandController{dev: dev}
}

func (c *CommandController) SendCommand(cmd byte) error {
	switch cmd {
	case 'W', 'S', 'A', 'D', 'X':
		_, err := c.dev.Write([]byte{cmd})
		return err
	default:
		return ErrInvalidCommand
	}
}

```

### Mock em Memória da Serial para Testes

```go
package controller_test

import (
	"bytes"
	"sync"
)

type MockSerialDevice struct {
	mu     sync.Mutex
	Buffer bytes.Buffer
	Err    error
}

func (m *MockSerialDevice) Write(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Err != nil {
		return 0, m.Err
	}
	return m.Buffer.Write(p)
}

func (m *MockSerialDevice) Bytes() []byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Buffer.Bytes()
}

```

---

## 4. Implementação dos Testes Unitários

### Teste do Controlador Serial (`internal/controller/controller_test.go`)

```go
package controller_test

import (
	"errors"
	"sync"
	"testing"

	"roverpi/internal/controller"
)

func TestSendCommand_ValidInputs(t *testing.T) {
	validCommands := []byte{'W', 'S', 'A', 'D', 'X'}

	for _, cmd := range validCommands {
		mockDev := &MockSerialDevice{}
		ctrl := controller.NewCommandController(mockDev)

		if err := ctrl.SendCommand(cmd); err != nil {
			t.Fatalf("esperava sucesso para comando '%c', obteve: %v", cmd, err)
		}

		out := mockDev.Bytes()
		if len(out) != 1 || out[0] != cmd {
			t.Errorf("esperava byte '%c' gravado, obteve '%v'", cmd, out)
		}
	}
}

func TestSendCommand_InvalidInput(t *testing.T) {
	mockDev := &MockSerialDevice{}
	ctrl := controller.NewCommandController(mockDev)

	err := ctrl.SendCommand('Z')
	if !errors.Is(err, controller.ErrInvalidCommand) {
		t.Fatalf("esperava ErrInvalidCommand, obteve: %v", err)
	}

	if len(mockDev.Bytes()) > 0 {
		t.Errorf("buffer nao deveria receber bytes para comandos invalidos")
	}
}

func TestSendCommand_Concurrency(t *testing.T) {
	mockDev := &MockSerialDevice{}
	ctrl := controller.NewCommandController(mockDev)
	var wg sync.WaitGroup

	iterations := 100
	wg.Add(iterations)

	for i := 0; i < iterations; i++ {
		go func() {
			defer wg.Done()
			_ = ctrl.SendCommand('W')
		}()
	}

	wg.Wait()

	if len(mockDev.Bytes()) != iterations {
		t.Errorf("esperava %d bytes gravados concorrentemente, obteve %d", iterations, len(mockDev.Bytes()))
	}
}

```

### Teste do Watchdog / Fail-Safe (`internal/transport/watchdog_test.go`)

```go
package transport_test

import (
	"sync/atomic"
	"testing"
	"time"
)

type MockCommander struct {
	stopDispatched atomic.Int32
}

func (m *MockCommander) DispatchStop() {
	m.stopDispatched.Add(1)
}

type Watchdog struct {
	timeout   time.Duration
	timer     *time.Timer
	commander *MockCommander
	alive     chan struct{}
}

func NewWatchdog(timeout time.Duration, cmd *MockCommander) *Watchdog {
	w := &Watchdog{
		timeout:   timeout,
		commander: cmd,
		alive:     make(chan struct{}, 1),
	}
	return w
}

func (w *Watchdog) Ping() {
	select {
	case w.alive <- struct{}{}:
	default:
	}
}

func (w *Watchdog) Start(stopSignal <-chan struct{}) {
	w.timer = time.NewTimer(w.timeout)
	go func() {
		for {
			select {
			case <-w.alive:
				if !w.timer.Stop() {
					select {
					case <-w.timer.C:
					default:
					}
				}
				w.timer.Reset(w.timeout)
			case <-w.timer.C:
				w.commander.DispatchStop()
				w.timer.Reset(w.timeout)
			case <-stopSignal:
				w.timer.Stop()
				w.commander.DispatchStop()
				return
			}
		}
	}()
}

func TestWatchdog_TriggersStopOnInactivity(t *testing.T) {
	commander := &MockCommander{}
	timeout := 50 * time.Millisecond
	wd := NewWatchdog(timeout, commander)
	stopChan := make(chan struct{})
	defer close(stopChan)

	wd.Start(stopChan)

	// Aguarda além da janela sem enviar ping
	time.Sleep(90 * time.Millisecond)

	if val := commander.stopDispatched.Load(); val == 0 {
		t.Errorf("fail-safe deveria ter disparado parada ('X'), total disparos: %d", val)
	}
}

func TestWatchdog_ResetOnActivity(t *testing.T) {
	commander := &MockCommander{}
	timeout := 80 * time.Millisecond
	wd := NewWatchdog(timeout, commander)
	stopChan := make(chan struct{})
	defer close(stopChan)

	wd.Start(stopChan)

	// Envia pings periódicos antes do timer expirar
	for i := 0; i < 4; i++ {
		time.Sleep(30 * time.Millisecond)
		wd.Ping()
	}

	if val := commander.stopDispatched.Load(); val != 0 {
		t.Errorf("fail-safe nao deveria ter disparado durante atividade continua, total: %d", val)
	}
}

```

### Teste do Parser de Marcadores JPEG da Câmera (`internal/camera/parser_test.go`)

```go
package camera_test

import (
	"bytes"
	"testing"
)

func FindJPEGFrame(stream []byte) ([]byte, int) {
	soi := []byte{0xFF, 0xD8}
	eoi := []byte{0xFF, 0xD9}

	start := bytes.Index(stream, soi)
	if start == -1 {
		return nil, 0
	}

	end := bytes.Index(stream[start:], eoi)
	if end == -1 {
		return nil, 0
	}

	endPos := start + end + 2
	return stream[start:endPos], endPos
}

func TestFindJPEGFrame_Extraction(t *testing.T) {
	fakeData := []byte{
		0x00, 0x12, // lixo de buffer anterior
		0xFF, 0xD8, // SOI
		0xAA, 0xBB, 0xCC, // carga do frame
		0xFF, 0xD9, // EOI
		0x01, 0x02, // bytes do próximo frame
	}

	frame, consumed := FindJPEGFrame(fakeData)

	if frame == nil {
		t.Fatalf("esperava encontrar um frame valido, obteve nil")
	}

	if len(frame) != 7 {
		t.Errorf("tamanho do frame incorreto. esperado 7, obteve %d", len(frame))
	}

	if consumed != 9 {
		t.Errorf("bytes consumidos incorretos. esperado 9, obteve %d", consumed)
	}
}

```

---

## 5. Instruções de Execução dos Testes

Execute a suite completa de testes no seu computador antes de gerar a compilação cruzada para o Raspberry Pi:

```bash
# Execução padrão dos testes unitários
go test -v ./...

# Execução com verificação estrita de concorrência (Race Detector)
go test -v -race ./...

# Relatório de cobertura por pacote
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

```

### Critérios de Qualidade para Aprovação (CI/CD)

* **Zero Race Conditions:** Execução limpa usando a flag `-race`.
* **Cobertura Mínima:** 85% nos pacotes `internal/controller` e `internal/transport`.
* **Isolamento de SO:** Nenhum teste pode falhar quando executado fora de um ambiente Linux/ARM64.