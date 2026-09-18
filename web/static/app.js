let ws = null;
let currentKey = null;
let controlLoop = null;

const statusEl = document.getElementById('status');
const connectBtn = document.getElementById('connectBtn');
const pinInput = document.getElementById('pinInput');

connectBtn.addEventListener('click', () => {
    if (ws) {
        ws.close();
        return;
    }
    connect();
});

function connect() {
    const pin = pinInput.value;
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws?pin=${pin}`;

    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
        statusEl.textContent = 'Conectado';
        statusEl.className = 'connected';
        connectBtn.textContent = 'Desconectar';
    };

    ws.onclose = () => {
        statusEl.textContent = 'Desconectado';
        statusEl.className = 'disconnected';
        connectBtn.textContent = 'Conectar';
        ws = null;
        stopLoop();
    };

    ws.onerror = (err) => {
        console.error('WebSocket Error:', err);
        ws.close();
    };
}

document.addEventListener('keydown', (e) => {
    if (!ws || ws.readyState !== WebSocket.OPEN) return;
    
    const key = e.key.toUpperCase();
    if (['W', 'A', 'S', 'D', 'Q', 'E'].includes(key)) {
        if (currentKey !== key) {
            currentKey = key;
            startLoop(key);
        }
    }
});

document.addEventListener('keyup', (e) => {
    const key = e.key.toUpperCase();
    if (key === currentKey) {
        currentKey = null;
        stopLoop();
        if (ws && ws.readyState === WebSocket.OPEN) {
            ws.send('X');
        }
    }
});

function startLoop(key) {
    stopLoop();
    // Send immediately then every 100ms
    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(key);
    }
    controlLoop = setInterval(() => {
        if (ws && ws.readyState === WebSocket.OPEN) {
            ws.send(key);
        }
    }, 100);
}

function stopLoop() {
    if (controlLoop) {
        clearInterval(controlLoop);
        controlLoop = null;
    }
}
