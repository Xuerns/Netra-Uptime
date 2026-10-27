import { apiFetch } from './client';
import { env } from '$env/dynamic/public';

const API_Route = `${env.PUBLIC_API_URL}`;

export interface APIKey {
	id: string;
	name: string;
	key_prefix: string;
	permissions: 'read' | 'write';
	last_used_at?: string;
	last_used_ip?: string;
	expires_at?: string;
	created_at: string;
	created_by: string;
}

export interface CreateAPIKeyRequest {
	name: string;
	permissions: 'read' | 'write';
	expires_in_days?: number | null;
}

export interface CreatedAPIKeyResponse {
	status: string;
	message: string;
	data: APIKey & { plaintext_key: string };
}

export const APIKeyAPI = {
	getAllByOrg: async (orgId: string): Promise<{ data: APIKey[] }> => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/api-keys`, { method: 'GET' });
		if (!res.ok) {
			const err = await res.json().catch(() => ({}));
			throw new Error(err.error || 'Failed to fetch API keys');
		}
		return res.json();
	},

	create: async (orgId: string, req: CreateAPIKeyRequest): Promise<CreatedAPIKeyResponse> => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/api-keys`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(req)
		});
		if (!res.ok) {
			const err = await res.json().catch(() => ({}));
			throw new Error(err.error || 'Failed to create API key');
		}
		return res.json();
	},

	revoke: async (orgId: string, keyId: string): Promise<void> => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/api-keys/${keyId}`, {
			method: 'DELETE'
		});
		if (!res.ok) {
			const err = await res.json().catch(() => ({}));
			throw new Error(err.error || 'Failed to revoke API key');
		}
	}
};
