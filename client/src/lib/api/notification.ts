import { apiFetch, handleApiError } from './client';
import { env } from '$env/dynamic/public';

const API = `${env.PUBLIC_API_URL}`;

export type ChannelType = 'email' | 'slack' | 'webhook' | 'discord';

export interface CreateChannelPayload {
	name: string;
	type: ChannelType;
	config: Record<string, string>;
}

export interface UpdateChannelPayload {
	name?: string;
	is_active?: boolean;
	config?: Record<string, string>;
}

export interface NotificationChannel {
	id: string;
	organization_id: string;
	name: string;
	type: ChannelType;
	config: Record<string, string>;
	is_active: boolean;
	created_at: string;
}

export interface NotificationLog {
	id: string;
	monitor_id?: string;
	incident_id?: string;
	channel_id?: string;
	channel: string;
	message: string;
	status: 'sent' | 'failed';
	error_detail?: string;
	sent_at: string;
}

export interface LogListResponse {
	logs: NotificationLog[];
	total: number;
	limit: number;
	offset: number;
}

export const NotificationAPI = {
	// ── Channels ──────────────────────────────────────────────────────────
	getChannels: async (orgId: string): Promise<NotificationChannel[]> => {
		const res = await apiFetch(`${API}/organization/${orgId}/notifications/channels`, {
			method: 'GET'
		});
		if (!res.ok) await handleApiError(res, 'Failed to fetch notification channels');
		const json = await res.json();
		return json.data ?? [];
	},

	getChannel: async (orgId: string, channelId: string): Promise<NotificationChannel> => {
		const res = await apiFetch(`${API}/organization/${orgId}/notifications/channels/${channelId}`, {
			method: 'GET'
		});
		if (!res.ok) await handleApiError(res, 'Failed to fetch notification channel');
		const json = await res.json();
		return json.data;
	},

	createChannel: async (
		orgId: string,
		payload: CreateChannelPayload
	): Promise<NotificationChannel> => {
		const res = await apiFetch(`${API}/organization/${orgId}/notifications/channels`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		if (!res.ok) await handleApiError(res, 'Failed to create notification channel');
		const json = await res.json();
		return json.data;
	},

	updateChannel: async (
		orgId: string,
		channelId: string,
		payload: UpdateChannelPayload
	): Promise<NotificationChannel> => {
		const res = await apiFetch(`${API}/organization/${orgId}/notifications/channels/${channelId}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		if (!res.ok) await handleApiError(res, 'Failed to update notification channel');
		const json = await res.json();
		return json.data;
	},

	deleteChannel: async (orgId: string, channelId: string): Promise<void> => {
		const res = await apiFetch(`${API}/organization/${orgId}/notifications/channels/${channelId}`, {
			method: 'DELETE'
		});
		if (!res.ok) await handleApiError(res, 'Failed to delete notification channel');
	},

	// ── Logs ──────────────────────────────────────────────────────────────
	getLogs: async (orgId: string, limit = 50, offset = 0): Promise<LogListResponse> => {
		const res = await apiFetch(
			`${API}/organization/${orgId}/notifications/logs?limit=${limit}&offset=${offset}`,
			{ method: 'GET' }
		);
		if (!res.ok) await handleApiError(res, 'Failed to fetch notification logs');
		const json = await res.json();
		return json.data ?? { logs: [], total: 0, limit, offset };
	}
};
