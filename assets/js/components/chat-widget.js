class ChatWidget extends HTMLElement {
    constructor() {
        super();
        this.ws = null;
    }

    connectedCallback() {
        this.innerHTML = `
            <div id="resize-handle"></div>
            <div id="chat-container"></div>
            <div id="status">Connecting...</div>
            <div id="input-container">
                <input type="text" id="message-input" placeholder="Type a message..." disabled>
                <button id="send-btn" disabled>Send</button>
            </div>
        `;

        this.chatContainer = this.querySelector('#chat-container');
        this.statusDiv = this.querySelector('#status');
        this.messageInput = this.querySelector('#message-input');
        this.sendBtn = this.querySelector('#send-btn');
        this.resizeHandle = this.querySelector('#resize-handle');

        this.connectWebSocket();
        this.setupEventListeners();
        this.setupResizer();
    }

    setupResizer() {
        let startX, startWidth;
        const appLayout = document.getElementById('app-layout');

        const initDrag = (e) => {
            startX = e.clientX;
            startWidth = this.getBoundingClientRect().width;
            document.documentElement.addEventListener('mousemove', doDrag, false);
            document.documentElement.addEventListener('mouseup', stopDrag, false);
            document.body.style.cursor = 'col-resize';
            document.body.style.userSelect = 'none'; // Prevent text selection
        };

        const doDrag = (e) => {
            // Moving mouse left (decreasing x) increases width
            const newWidth = startWidth + (startX - e.clientX);
            
            // Constraints: Min 250px, Max 60% of window width
            if (newWidth > 250 && newWidth < window.innerWidth * 0.6) {
                appLayout.style.gridTemplateColumns = `1fr ${newWidth}px`;
            }
        };

        const stopDrag = () => {
            document.documentElement.removeEventListener('mousemove', doDrag, false);
            document.documentElement.removeEventListener('mouseup', stopDrag, false);
            document.body.style.cursor = '';
            document.body.style.userSelect = '';
        };

        this.resizeHandle.addEventListener('mousedown', initDrag, false);
    }

    connectWebSocket() {
        const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${wsProtocol}//${window.location.host}/ws`;
        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
            this.statusDiv.textContent = 'Connected';
            this.statusDiv.style.color = 'green';
            this.messageInput.disabled = false;
            this.sendBtn.disabled = false;
        };

        this.ws.onclose = () => {
            this.statusDiv.textContent = 'Disconnected';
            this.statusDiv.style.color = 'red';
            this.messageInput.disabled = true;
            this.sendBtn.disabled = true;
            // Retry connection after 5 seconds
            setTimeout(() => this.connectWebSocket(), 5000);
        };

        this.ws.onmessage = (event) => {
            const data = JSON.parse(event.data);
            if (data.type === 'message') {
                this.appendMessage('bot', data.content);
            }
        };
    }

    setupEventListeners() {
        this.sendBtn.onclick = () => this.sendMessage();
        this.messageInput.onkeypress = (e) => {
            if (e.key === 'Enter') this.sendMessage();
        };
    }

    sendMessage() {
        const text = this.messageInput.value.trim();
        if (text && this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({ content: text }));
            this.appendMessage('user', text);
            this.messageInput.value = '';
        }
    }

    appendMessage(sender, text) {
        const msgDiv = document.createElement('div');
        msgDiv.className = `message ${sender}`;
        // Use marked if available, otherwise just text
        if (typeof marked !== 'undefined') {
            msgDiv.innerHTML = marked.parse(text);
        } else {
            msgDiv.textContent = text;
        }
        this.chatContainer.appendChild(msgDiv);
        this.chatContainer.scrollTop = this.chatContainer.scrollHeight;
    }
}

customElements.define('chat-widget', ChatWidget);
