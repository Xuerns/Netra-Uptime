import { env } from '$env/dynamic/public';
import { browser } from '$app/environment';
import { logger } from './logger';

export type MonitorStatus = 'up' | 'down' | 'degraded' | 'paused' | 'pending';

export interface MonitorUpdatePayload {
	id: string;
	status: MonitorStatus;
	response_time: number;
	last_checked: string;
	failure_count: number;
}

export type WebSocketPayload = MonitorUpdatePayload | Record<string, unknown> | null;

export interface WebSocketMessage<T = WebSocketPayload> {
	Room: string;
	Type: string;
	Payload: T;
}

export type WSMessage<T = WebSocketPayload> = WebSocketMessage<T>;

type Subscriber = (msg: WSMessage) => void;

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

		logger.info('[WebSocket] Connecting to:', wsUrl);
		this.ws = new WebSocket(wsUrl);

		this.ws.onopen = () => {
			logger.info('[WebSocket] Connected successfully');
			this.isConnecting = false;
			this.reconnectAttempts = 0;
		};

		this.ws.onmessage = (event) => {
			try {
				const data: WSMessage = JSON.parse(event.data);
				this.notifySubscribers(data);
			} catch (err) {
				logger.error('[WebSocket] Failed to parse message:', err);
			}
		};

		this.ws.onclose = () => {
			logger.info('[WebSocket] Connection closed');
			this.isConnecting = false;
			this.ws = null;
			this.scheduleReconnect();
		};

		this.ws.onerror = (err) => {
			logger.error('[WebSocket] Error:', err);
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
			logger.info('[WebSocket] Max reconnect attempts reached');
			return;
		}

		const delay = Math.min(1000 * 2 ** this.reconnectAttempts, 10000); // Exponential backoff
		this.reconnectAttempts++;

		logger.info(`[WebSocket] Reconnecting in ${delay}ms... (Attempt ${this.reconnectAttempts})`);
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
