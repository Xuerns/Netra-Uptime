import { apiFetch, handleApiError } from './client';
import { env } from '$env/dynamic/public';
import type { AlertRule, CreateAlertRuleRequest, UpdateAlertRuleRequest } from './alertRule';

const API_Route = `${env.PUBLIC_API_URL}`;

export interface AlertTemplate {
	id: string;
	organization_id: string;
	name: string;
	description: string;
	is_default: boolean;
	rules: AlertRule[];
	monitors_count: number;
	created_at: string;
	updated_at: string;
}

export interface CreateAlertTemplateRequest {
	name: string;
	description?: string;
	is_default?: boolean;
	rules?: CreateAlertRuleRequest[];
}

export interface UpdateAlertTemplateRequest {
	name?: string;
	description?: string;
	is_default?: boolean;
}

export const AlertTemplateAPI = {
	list: async (orgId: string): Promise<{ status: string; data: AlertTemplate[] }> => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/alert-templates`, {
			method: 'GET'
		});
		if (!res.ok) await handleApiError(res, 'Failed to fetch alert templates');
		return res.json();
	},

	getById: async (
		orgId: string,
		templateId: string
	): Promise<{ status: string; data: AlertTemplate }> => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/alert-templates/${templateId}`, {
			method: 'GET'
		});
		if (!res.ok) await handleApiError(res, 'Failed to fetch alert template');
		return res.json();
	},

	create: async (
		orgId: string,
		payload: CreateAlertTemplateRequest
	): Promise<{ status: string; data: AlertTemplate }> => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/alert-templates`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		if (!res.ok) await handleApiError(res, 'Failed to create alert template');
		return res.json();
	},

	update: async (
		orgId: string,
		templateId: string,
		payload: UpdateAlertTemplateRequest
	): Promise<{ status: string; data: AlertTemplate }> => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/alert-templates/${templateId}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		if (!res.ok) await handleApiError(res, 'Failed to update alert template');
		return res.json();
	},

	delete: async (
		orgId: string,
		templateId: string
	): Promise<{ status: string; message: string }> => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/alert-templates/${templateId}`, {
			method: 'DELETE'
		});
		if (!res.ok) await handleApiError(res, 'Failed to delete alert template');
		return res.json();
	},

	createRule: async (
		orgId: string,
		templateId: string,
		payload: CreateAlertRuleRequest
	): Promise<{ status: string; data: AlertRule }> => {
		const res = await apiFetch(
			`${API_Route}/organization/${orgId}/alert-templates/${templateId}/rules`,
			{
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(payload)
			}
		);
		if (!res.ok) await handleApiError(res, 'Failed to add rule to template');
		return res.json();
	},

	updateRule: async (
		orgId: string,
		templateId: string,
		ruleId: string,
		payload: UpdateAlertRuleRequest
	): Promise<{ status: string; data: AlertRule }> => {
		const res = await apiFetch(
			`${API_Route}/organization/${orgId}/alert-templates/${templateId}/rules/${ruleId}`,
			{
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(payload)
			}
		);
		if (!res.ok) await handleApiError(res, 'Failed to update rule');
		return res.json();
	},

	deleteRule: async (
		orgId: string,
		templateId: string,
		ruleId: string
	): Promise<{ status: string; message: string }> => {
		const res = await apiFetch(
			`${API_Route}/organization/${orgId}/alert-templates/${templateId}/rules/${ruleId}`,
			{
				method: 'DELETE'
			}
		);
		if (!res.ok) await handleApiError(res, 'Failed to remove rule from template');
		return res.json();
	},

	assignToMonitor: async (
		orgId: string,
		monitorId: string,
		templateId: string | null
	): Promise<{ status: string; message: string }> => {
		const res = await apiFetch(
			`${API_Route}/organization/${orgId}/monitors/${monitorId}/template`,
			{
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ template_id: templateId })
			}
		);
		if (!res.ok) await handleApiError(res, 'Failed to assign template to monitor');
		return res.json();
	}
};
