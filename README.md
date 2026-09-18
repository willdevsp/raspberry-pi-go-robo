# RoverPi Server

Servidor de controle, telemetria e streaming de vídeo MJPEG do Raspberry Pi 3B para operação remota.

## Build

Como a compilação é para o Raspberry Pi (ARM64), use as flags corretas na sua máquina de desenvolvimento:

```bash
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o rover-server ./cmd/server
```

Copie para o Raspberry Pi:
```bash
scp rover-server pi@<ip_do_pi>:/home/pi/
```

## Configurar Raspberry Pi como Access Point (AP)

Para usar o carrinho em ambientes sem Wi-Fi (parques, rua, etc), ative o modo Access Point no Raspberry Pi usando o NetworkManager (padrão no Raspberry Pi OS Bookworm):

```bash
# Cria uma rede chamada RoverPi_Net com senha "43214321"
sudo nmcli device wifi hotspot ifname wlan0 ssid RoverPi_Net password 43214321
```

O seu notebook ou celular pode então se conectar a esta rede para controlar o RoverPi diretamente com baixa latência.

## Execução

Inicie o servidor (geralmente via `systemd` ou terminal):

```bash
./rover-server -port=8080 -serial=/dev/ttyUSB0
```

Se o serial não for encontrado, ele rodará em "dry run" para fins de teste da interface UI.

## Uso

Acesse o endereço IP do Pi no browser:
`http://<IP>:8080` (PIN: 4321)

Controles W, A, S, D, Q, E.
