import { env } from '$env/dynamic/public';
import { browser } from '$app/environment';

export type WSMessage = {
	Room: string;
	Type: string;
	Payload: any;
};

export type MonitorUpdatePayload = {
	id: string;
	status: 'up' | 'down' | 'paused' | 'pending';
	response_time: number;
	last_checked: string;
	failure_count: number;
};

type Subscriber = (msg: WSMessage) => void;

function logDev(...args: unknown[]) {
	if (import.meta.env.DEV) {
		console.log(...args);
	}
}

function logDevError(...args: unknown[]) {
	if (import.meta.env.DEV) {
		console.error(...args);
	}
}

class WebSocketClient {
	private ws: WebSocket | null = null;
	private subscribers: Set<Subscriber> = new Set();
	private isConnecting = false;
	private reconnectAttempts = 0;
	private maxReconnectAttempts = 5;

	connect() {
		if (!browser || this.ws?.readyState === WebSocket.OPEN || this.isConnecting) return;

		this.isConnecting = true;

		// Convert http:// to ws:// and https:// to wss://
		const apiBase = env.PUBLIC_API_URL || 'http://localhost:27346/api/v1';
		const wsUrl = apiBase.replace(/^http/, 'ws') + '/ws';

		logDev('[WebSocket] Connecting to:', wsUrl);
		this.ws = new WebSocket(wsUrl);

		this.ws.onopen = () => {
			logDev('[WebSocket] Connected successfully');
			this.isConnecting = false;
			this.reconnectAttempts = 0;
		};

		this.ws.onmessage = (event) => {
			try {
				const data: WSMessage = JSON.parse(event.data);
				this.notifySubscribers(data);
			} catch (err) {
				logDevError('[WebSocket] Failed to parse message:', err);
			}
		};

		this.ws.onclose = () => {
			logDev('[WebSocket] Connection closed');
			this.isConnecting = false;
			this.ws = null;
			this.scheduleReconnect();
		};

		this.ws.onerror = (err) => {
			logDevError('[WebSocket] Error:', err);
			this.ws?.close();
		};
	}

	subscribe(callback: Subscriber) {
		this.subscribers.add(callback);
		if (this.subscribers.size === 1) {
			this.connect();
		}

		return () => {
			this.subscribers.delete(callback);
			if (this.subscribers.size === 0) {
				this.disconnect();
			}
		};
	}

	private notifySubscribers(msg: WSMessage) {
		for (const sub of this.subscribers) {
			sub(msg);
		}
	}

	private scheduleReconnect() {
		if (this.reconnectAttempts >= this.maxReconnectAttempts) {
			logDev('[WebSocket] Max reconnect attempts reached');
			return;
		}

		const delay = Math.min(1000 * 2 ** this.reconnectAttempts, 10000); // Exponential backoff
		this.reconnectAttempts++;

		logDev(`[WebSocket] Reconnecting in ${delay}ms... (Attempt ${this.reconnectAttempts})`);
		setTimeout(() => this.connect(), delay);
	}

	disconnect() {
		if (this.ws) {
			this.ws.close();
			this.ws = null;
		}
	}
}

// Export a singleton instance
export const wsClient = new WebSocketClient();
