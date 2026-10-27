import { apiFetch } from './client';
import { env } from '$env/dynamic/public';
import type { Monitor } from './monitor';

const API_Route = `${env.PUBLIC_API_URL}`;

export interface IncidentUpdateResponse {
	message: string;
	username: string;
	avatarUrl: string;
	createAt: string;
}

export interface IncidentAssignee {
	ID: string;
	IncidentID: string;
	UserID: string;
	Role: string;
	CreatedAt: string;
	User?: {
		Name: string;
		Avatar: string;
	};
}

export interface Incident {
	ID: string;
	MonitorID: string;
	Monitor?: Monitor;
	Status: 'ongoing' | 'resolved';
	Summary: string;
	StartedAt: string;
	ResolvedAt?: string;
	CreatedAt: string;
	UpdatedAt: string;
	Assignees?: IncidentAssignee[];

	Severity: string;
	SeverityCode: 'sev1' | 'sev2' | 'sev3' | 'sev4';
	SeverityLabel: string;
	Assignee?: {
		Name: string;
		Avatar: string;
	};
}

export interface AddIncidentUpdateReq {
	message: string;
}

function mapSeverityFromBackend(severity?: string): {
	code: 'sev1' | 'sev2' | 'sev3' | 'sev4';
	label: string;
} {
	switch (severity?.toLowerCase()) {
		case 'critical':
			return { code: 'sev1', label: 'Sev-1 Critical' };
		case 'high':
			return { code: 'sev2', label: 'Sev-2 High' };
		case 'medium':
			return { code: 'sev3', label: 'Sev-3 Medium' };
		case 'low':
			return { code: 'sev4', label: 'Sev-4 Low' };
		default:
			return { code: 'sev3', label: 'Sev-3 Medium' };
	}
}

function mapIncident(inc: any): Incident {
	const rawSev = inc.severity ?? inc.Severity ?? 'medium';
	const sevMeta = mapSeverityFromBackend(rawSev);

	return {
		ID: inc.id ?? inc.ID,
		MonitorID: inc.monitor_id ?? inc.MonitorID,
		Status: inc.status ?? inc.Status,
		Summary: inc.summary ?? inc.Summary,
		StartedAt: inc.started_at ?? inc.StartedAt,
		ResolvedAt: inc.resolved_at ?? inc.ResolvedAt,
		CreatedAt: inc.created_at ?? inc.CreatedAt,
		UpdatedAt: inc.updated_at ?? inc.UpdatedAt,
		Monitor:
			inc.monitor || inc.Monitor
				? ({
						Name: inc.monitor?.name ?? inc.Monitor?.Name,
						Target: inc.monitor?.target ?? inc.Monitor?.Target,
						OrganizationID: inc.monitor?.organization_id ?? inc.Monitor?.OrganizationID
					} as any)
				: undefined,
		Assignees: (inc.assignees ?? inc.Assignees ?? []).map((a: any) => ({
			ID: a.id ?? a.ID,
			IncidentID: a.incident_id ?? a.IncidentID,
			UserID: a.user_id ?? a.UserID,
			Role: a.role ?? a.Role,
			CreatedAt: a.created_at ?? a.CreatedAt,
			User:
				a.user || a.User
					? {
							Name: a.user?.name ?? a.User?.Name,
							Avatar: a.user?.avatar ?? a.User?.Avatar
						}
					: undefined
		})),
		Severity: rawSev,
		SeverityCode: sevMeta.code,
		SeverityLabel: sevMeta.label,
		Assignee: inc.Assignee
	};
}

export const IncidentAPI = {
	getAll: async (orgId: string, monitorId?: string): Promise<{ data: Incident[] }> => {
		const url = monitorId
			? `${API_Route}/organization/${orgId}/incident?monitor_id=${encodeURIComponent(monitorId)}`
			: `${API_Route}/organization/${orgId}/incident`;
		const res = await apiFetch(url, { method: 'GET' });
		if (!res.ok) throw new Error('Failed to fetch incidents');

		const data = await res.json();
		if (data.data) {
			data.data = data.data.map((rawInc: any) => mapIncident(rawInc));
		}

		return data;
	},

	getById: async (orgId: string, incidentId: string): Promise<{ data: Incident }> => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/incident/${incidentId}`, {
			method: 'GET'
		});
		if (!res.ok) throw new Error('Failed to fetch incident detail');
		const data = await res.json();
		if (data.data) {
			data.data = mapIncident(data.data);
		}
		return data;
	},

	addUpdate: async (orgId: string, incidentId: string, data: AddIncidentUpdateReq) => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/incident/${incidentId}`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(data)
		});
		if (!res.ok) {
			const errData = await res.json().catch(() => ({}));
			throw new Error(errData.error || 'Failed to add update');
		}
		return res.json();
	},

	joinAssignee: async (orgId: string, incidentId: string) => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/incident/${incidentId}/join`, {
			method: 'POST'
		});
		if (!res.ok) {
			const errData = await res.json().catch(() => ({}));
			throw new Error(errData.error || 'Failed to join incident');
		}
		return res.json();
	},

	getUpdates: async (
		orgId: string,
		incidentId: string
	): Promise<{ data: IncidentUpdateResponse[] }> => {
		const res = await apiFetch(
			`${API_Route}/organization/${orgId}/incident/${incidentId}/updates`,
			{
				method: 'GET'
			}
		);
		if (!res.ok) throw new Error('Failed to fetch incident updates');
		return res.json();
	}
};
