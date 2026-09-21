import { apiFetch } from '../helper/api.ts';
import { env } from '$env/dynamic/public';

const API_Route = `${env.PUBLIC_API_URL}`;

export interface CreateStatusPageRequest {
	subdomain: string;
	title: string;
	description?: string | null;
	is_public: boolean;
	monitor_id: string[];
}

export interface UpdateStatusPageRequest {
	subdomain?: string;
	title?: string;
	description?: string | null;
	is_public?: boolean;
	monitor_id?: string[];
}

export interface StatusPageMonitor {
	ID: string;
	StatusPageID: string;
	MonitorID: string;
	DisplayOrder: number;
}

export interface PublicIncidentUpdate {
	message: string;
	created_at: string;
}

export interface PublicIncident {
	id: string;
	status: string; // ongoing | resolved
	severity: string; // critical | high | medium | low
	summary: string;
	started_at: string;
	resolved_at?: string | null;
	monitor_name?: string;
	updates: PublicIncidentUpdate[];
}

export interface PublicUptimeDay {
	date: string;
	uptime_percentage: number;
	down_checks: number;
	total_checks: number;
	has_data: boolean;
}

export interface PublicMaintenance {
	id: string;
	title: string;
	description: string;
	status: 'scheduled' | 'in_progress' | 'completed' | string;
	scheduled_start: string;
	scheduled_end: string;
	affected_monitors: string[];
}

export interface PublicMonitor {
	id: string;
	name: string;
	type: string;
	status: 'up' | 'down' | 'degraded' | 'paused' | string;
	uptime_percent?: number | null;
	recent_checks: number[];
	daily_stats?: PublicUptimeDay[];
}

export interface PublicStatusPage {
	title: string;
	description?: string | null;
	subdomain: string;
	overall_status: 'operational' | 'degraded' | 'partial_outage' | 'major_outage';
	monitors: PublicMonitor[];
	active_incidents: PublicIncident[];
	recent_incidents: PublicIncident[];
	upcoming_maintenances?: PublicMaintenance[];
	ongoing_maintenances?: PublicMaintenance[];
}

export interface Maintenance {
	id: string;
	status_page_id: string;
	title: string;
	description?: string;
	status: 'scheduled' | 'in_progress' | 'completed';
	scheduled_start: string;
	scheduled_end: string;
	actual_start?: string | null;
	actual_end?: string | null;
	created_at: string;
	affected_monitors: string[];
	affected_monitor_ids?: string[];
}

export interface CreateMaintenanceRequest {
	title: string;
	description?: string;
	scheduled_start: string;
	scheduled_end: string;
	monitor_ids?: string[];
}

export interface UpdateMaintenanceRequest {
	title?: string;
	description?: string;
	status?: 'scheduled' | 'in_progress' | 'completed';
	scheduled_start?: string;
	scheduled_end?: string;
	monitor_ids?: string[];
}

export interface StatusPage {
	ID: string;
	OrganizationID: string;
	Organization?: { ID: string; Name: string; Slug: string };
	Subdomain: string;
	Title: string;
	Description?: string | null;
	IsPublic: boolean;
	CreatedAt: string;
	StatusPageMonitors?: StatusPageMonitor[];
}

function mapStatusPage(sp: any): StatusPage {
	if (!sp) return sp;
	return {
		ID: sp.id ?? sp.ID,
		OrganizationID: sp.organization_id ?? sp.OrganizationID,
		Organization: sp.organization
			? {
					ID: sp.organization.id ?? sp.organization.ID,
					Name: sp.organization.name ?? sp.organization.Name,
					Slug: sp.organization.slug ?? sp.organization.Slug
				}
			: sp.Organization,
		Subdomain: sp.subdomain ?? sp.Subdomain,
		Title: sp.title ?? sp.Title,
		Description: sp.description ?? sp.Description,
		IsPublic: sp.is_public ?? sp.IsPublic,
		CreatedAt: sp.created_at ?? sp.CreatedAt,
		StatusPageMonitors: (sp.status_page_monitors || sp.StatusPageMonitors || []).map((m: any) => ({
			ID: m.id ?? m.ID,
			StatusPageID: m.status_page_id ?? m.StatusPageID,
			MonitorID: m.monitor_id ?? m.MonitorID,
			DisplayOrder: m.display_order ?? m.DisplayOrder
		}))
	};
}

export const StatusPageAPI = {
	getAll: async (orgId: string) => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/statuspage`, { method: 'GET' });
		if (!res.ok) throw new Error('Failed to fetch status pages');
		const json = await res.json();
		return {
			...json,
			data: (json.data || []).map(mapStatusPage)
		};
	},

	getAllByUser: async () => {
		const res = await apiFetch(`${API_Route}/statuspages`, { method: 'GET' });
		if (!res.ok) throw new Error('Failed to fetch user status pages');
		const json = await res.json();
		return {
			...json,
			data: (json.data || []).map(mapStatusPage)
		};
	},

	getMonitors: async (orgId: string, statusPageId: string) => {
		const res = await apiFetch(
			`${API_Route}/organization/${orgId}/statuspage/${statusPageId}/monitors`,
			{
				method: 'GET'
			}
		);
		if (!res.ok) throw new Error('Failed to fetch status page monitors');
		const json = await res.json();
		return {
			...json,
			data: (json.data || []).map((m: any) => ({
				ID: m.id ?? m.ID,
				StatusPageID: m.status_page_id ?? m.StatusPageID,
				MonitorID: m.monitor_id ?? m.MonitorID,
				DisplayOrder: m.display_order ?? m.DisplayOrder
			}))
		};
	},

	create: async (orgId: string, data: CreateStatusPageRequest) => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/statuspage`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(data)
		});
		if (!res.ok) throw new Error('Failed to create status page');
		const json = await res.json();
		return {
			...json,
			data: json.data ? mapStatusPage(json.data) : json.data
		};
	},

	update: async (orgId: string, statusPageId: string, data: UpdateStatusPageRequest) => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/statuspage/${statusPageId}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(data)
		});
		if (!res.ok) throw new Error('Failed to update status page');
		const json = await res.json();
		return {
			...json,
			data: json.data ? mapStatusPage(json.data) : json.data
		};
	},

	delete: async (orgId: string, statusPageId: string) => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/statuspage/${statusPageId}`, {
			method: 'DELETE'
		});
		if (!res.ok) throw new Error('Failed to delete status page');
		return res.json();
	},

	createMaintenance: async (orgId: string, statusPageId: string, data: CreateMaintenanceRequest) => {
		const res = await apiFetch(
			`${API_Route}/organization/${orgId}/statuspage/${statusPageId}/maintenance`,
			{
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(data)
			}
		);
		if (!res.ok) throw new Error('Failed to create maintenance');
		return res.json();
	},

	getMaintenances: async (orgId: string, statusPageId: string) => {
		const res = await apiFetch(
			`${API_Route}/organization/${orgId}/statuspage/${statusPageId}/maintenance`,
			{
				method: 'GET'
			}
		);
		if (!res.ok) throw new Error('Failed to fetch maintenances');
		return res.json();
	},

	updateMaintenance: async (
		orgId: string,
		statusPageId: string,
		maintenanceId: string,
		data: UpdateMaintenanceRequest
	) => {
		const res = await apiFetch(
			`${API_Route}/organization/${orgId}/statuspage/${statusPageId}/maintenance/${maintenanceId}`,
			{
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(data)
			}
		);
		if (!res.ok) throw new Error('Failed to update maintenance');
		return res.json();
	},

	deleteMaintenance: async (orgId: string, statusPageId: string, maintenanceId: string) => {
		const res = await apiFetch(
			`${API_Route}/organization/${orgId}/statuspage/${statusPageId}/maintenance/${maintenanceId}`,
			{
				method: 'DELETE'
			}
		);
		if (!res.ok) throw new Error('Failed to delete maintenance');
		return res.json();
	}
};
