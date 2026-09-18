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

## Execução Automática (Systemd)

Para que o carrinho ligue o servidor sozinho assim que você colocar a bateria no Raspberry Pi, vamos usar o `systemd`.

1. Copie o arquivo de serviço para a pasta do sistema:
```bash
sudo cp deploy/roverpi.service /etc/systemd/system/
```

2. Recarregue os serviços e ative a inicialização junto com o sistema:
```bash
sudo systemctl daemon-reload
sudo systemctl enable roverpi.service
```

3. Inicie o serviço agora mesmo:
```bash
sudo systemctl start roverpi.service
```

### Como Atualizar o Código

Sempre que você recompilar o projeto no seu PC e quiser jogar uma versão nova pro carrinho, siga este fluxo para **não dar conflito de porta ou erro de arquivo em uso**:

1. **PARE** o serviço rodando no Raspberry Pi:
```bash
sudo systemctl stop roverpi.service
```
2. Mande o arquivo novo via SCP do seu PC:
```bash
scp rover-server robocarrinho@<ip_do_pi>:~/
```
3. **INICIE** o serviço novamente no Pi:
```bash
sudo systemctl start roverpi.service
```

*(Dica: Para ver os logs em tempo real e ver os botões que estão sendo apertados, use `sudo journalctl -u roverpi.service -f`)*

## Uso

Acesse o endereço IP do Pi no browser:
`http://<IP>:8080` (PIN: 4321)

Controles W, A, S, D, Q, E.
