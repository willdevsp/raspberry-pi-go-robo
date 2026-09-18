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

function handleKeyDown(key) {
    if (!ws || ws.readyState !== WebSocket.OPEN) return;
    
    if (['W', 'A', 'S', 'D', 'Q', 'E'].includes(key)) {
        if (currentKey !== key) {
            currentKey = key;
            startLoop(key);
        }
    }
}

function handleKeyUp(key) {
    if (key === currentKey) {
        currentKey = null;
        stopLoop();
        if (ws && ws.readyState === WebSocket.OPEN) {
            ws.send('X');
        }
    }
}

document.addEventListener('keydown', (e) => handleKeyDown(e.key.toUpperCase()));
document.addEventListener('keyup', (e) => handleKeyUp(e.key.toUpperCase()));

// Virtual D-pad logic
const dpadBtns = document.querySelectorAll('.dpad-btn');
dpadBtns.forEach(btn => {
    const key = btn.getAttribute('data-key');
    
    // Touch events for mobile
    btn.addEventListener('touchstart', (e) => {
        e.preventDefault();
        handleKeyDown(key);
    });
    btn.addEventListener('touchend', (e) => {
        e.preventDefault();
        handleKeyUp(key);
    });
    btn.addEventListener('touchcancel', (e) => {
        e.preventDefault();
        handleKeyUp(key);
    });
    
    // Mouse events for desktop
    btn.addEventListener('mousedown', () => handleKeyDown(key));
    btn.addEventListener('mouseup', () => handleKeyUp(key));
    btn.addEventListener('mouseleave', () => handleKeyUp(key));
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
