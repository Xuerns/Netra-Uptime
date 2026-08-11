import { apiFetch } from '../helper/api.ts';
import { env } from '$env/dynamic/public';

const API_Route = `${env.PUBLIC_API_URL}`;

export interface AuditLogItem {
	id: string;
	organization_id: string;
	user_id: string;
	user_name?: string;
	user_email?: string;
	action: string;
	entity_type: string;
	entity_id?: string;
	description: string;
	details?: string;
	ip_address?: string;
	created_at: string;
}

export interface GetAuditLogsParams {
	page?: number;
	limit?: number;
	action?: string;
	entity_type?: string;
	search?: string;
}

export interface PaginatedAuditLogsResponse {
	logs: AuditLogItem[];
	total_count: number;
	page: number;
	limit: number;
	total_pages: number;
}

export const AuditAPI = {
	getLogs: async (orgId: string, params?: GetAuditLogsParams) => {
		const query = new URLSearchParams();
		if (params?.page) query.append('page', params.page.toString());
		if (params?.limit) query.append('limit', params.limit.toString());
		if (params?.action) query.append('action', params.action);
		if (params?.entity_type) query.append('entity_type', params.entity_type);
		if (params?.search) query.append('search', params.search);

		const queryString = query.toString() ? `?${query.toString()}` : '';
		const res = await apiFetch(`${API_Route}/organization/${orgId}/audit-logs${queryString}`, {
			method: 'GET'
		});

		if (!res.ok) {
			const errJson = await res.json().catch(() => ({}));
			throw new Error(errJson.error || 'Failed to fetch audit logs');
		}

		const json = await res.json();
		return json.data as PaginatedAuditLogsResponse;
	}
};
