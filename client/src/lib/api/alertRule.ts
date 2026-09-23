import { apiFetch } from '../helper/api.ts';
import { env } from '$env/dynamic/public';

const API_Route = `${env.PUBLIC_API_URL}`;

export interface AlertRule {
	id: string;
	organization_id: string;
	monitor_id?: string;
	name: string;
	metric: 'status_code' | 'response_time_ms' | 'timeout';
	operator: '==' | '!=' | '>' | '>=' | '<' | '<=' | 'in' | 'not_in';
	target_value: string;
	consecutive: number;
	severity: 'critical' | 'warning';
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

export interface CreateAlertRuleRequest {
	name: string;
	metric: 'status_code' | 'response_time_ms' | 'timeout';
	operator: '==' | '!=' | '>' | '>=' | '<' | '<=' | 'in' | 'not_in';
	target_value: string;
	consecutive: number;
	severity: 'critical' | 'warning';
	is_active?: boolean;
}

export interface UpdateAlertRuleRequest {
	name?: string;
	metric?: 'status_code' | 'response_time_ms' | 'timeout';
	operator?: '==' | '!=' | '>' | '>=' | '<' | '<=' | 'in' | 'not_in';
	target_value?: string;
	consecutive?: number;
	severity?: 'critical' | 'warning';
	is_active?: boolean;
}

export const AlertRuleAPI = {
	getByMonitor: async (
		orgId: string,
		monitorId: string
	): Promise<{ status: string; data: AlertRule[] }> => {
		const res = await apiFetch(
			`${API_Route}/organization/${orgId}/monitors/${monitorId}/alert-rules`,
			{
				method: 'GET'
			}
		);
		if (!res.ok) throw new Error('Failed to fetch alert rules');
		return res.json();
	},

	create: async (
		orgId: string,
		monitorId: string,
		payload: CreateAlertRuleRequest
	): Promise<{ status: string; data: AlertRule }> => {
		const res = await apiFetch(
			`${API_Route}/organization/${orgId}/monitors/${monitorId}/alert-rules`,
			{
				method: 'POST',
				body: JSON.stringify(payload)
			}
		);
		if (!res.ok) {
			const err = await res.json().catch(() => ({}));
			throw new Error(err.error || 'Failed to create alert rule');
		}
		return res.json();
	},

	update: async (
		orgId: string,
		ruleId: string,
		payload: UpdateAlertRuleRequest
	): Promise<{ status: string; data: AlertRule }> => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/alert-rules/${ruleId}`, {
			method: 'PUT',
			body: JSON.stringify(payload)
		});
		if (!res.ok) {
			const err = await res.json().catch(() => ({}));
			throw new Error(err.error || 'Failed to update alert rule');
		}
		return res.json();
	},

	delete: async (orgId: string, ruleId: string): Promise<{ status: string; message: string }> => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/alert-rules/${ruleId}`, {
			method: 'DELETE'
		});
		if (!res.ok) {
			const err = await res.json().catch(() => ({}));
			throw new Error(err.error || 'Failed to delete alert rule');
		}
		return res.json();
	}
};
