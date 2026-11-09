import { apiFetch, handleApiError } from './client';
import { env } from '$env/dynamic/public';

const API_Route = `${env.PUBLIC_API_URL}`;

export interface CreateMonitorRequest {
	name: string;
	type: 'http' | 'tcp' | 'ping';
	target: string;
	interval_seconds: number;
	timeout_seconds: number;
	expectation_status_code?: number;
	http_method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'HEAD';
	http_headers?: string;
	http_body?: string;
	criticality: 'extreme' | 'high' | 'medium' | 'low';
	MonitorPICs?: string[];
	assignees?: string[];
	enable_ssl_domain_check?: boolean;
	alert_template_id?: string | null;
}

export type CreateMonitorPayload = CreateMonitorRequest;

export interface UpdateMonitorRequest {
	name?: string;
	target?: string;
	interval_seconds?: number;
	timeout_seconds?: number;
	expectation_status_code?: number;
	http_method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'HEAD';
	http_headers?: string;
	http_body?: string;
	is_active?: boolean;
	criticality?: 'extreme' | 'high' | 'medium' | 'low';
	MonitorPICs?: string[];
	assignees?: string[];
	enable_ssl_domain_check?: boolean;
	alert_template_id?: string | null;
}

export type UpdateMonitorPayload = UpdateMonitorRequest;

export interface RecentCheck {
	ID: string;
	MonitorID: string;
	Status: 'up' | 'down' | 'degraded' | 'maintenance';
	StatusCode: number;
	ResponseTimeMs: number;
	ErrorMessage?: string;
	ResponseSnippet?: string;
	CheckedAt: string;
}

export interface MonitorDailyStat {
	ID: string;
	MonitorID: string;
	Date: string;
	TotalChecks: number;
	UpChecks: number;
	DownChecks: number;
	UptimePercentage: number;
	AvgLatencyMs: number;
	MinLatencyMs: number;
	MaxLatencyMs: number;
	Count2xx?: number;
	Count3xx?: number;
	Count4xx?: number;
	Count5xx?: number;
	CountError?: number;
}

export interface FleetOverviewResponse {
	window_days: number;
	uptime: {
		current_percent?: number | null;
		previous_percent?: number | null;
		trend_delta?: number | null;
	};
	latency: {
		avg_ms?: number | null;
		median_ms?: number | null;
		p95_ms?: number | null;
		fastest_ms?: number | null;
	};
	fleet: {
		total_targets: number;
		up: number;
		degraded: number;
		down: number;
		paused: number;
		intervals: {
			min_sec: number;
			max_sec: number;
			mode_sec: number;
			is_uniform: boolean;
		};
	};
	certificates_and_dns: {
		ssl_healthy: number;
		ssl_warnings: number;
		dns_healthy: number;
		dns_warnings: number;
		total_notices: number;
	};
}

export interface Monitor {
	ID: string;
	Name: string;
	Type: 'http' | 'tcp' | 'ping';
	Target: string;
	Status: 'up' | 'down' | 'degraded' | 'paused' | 'pending';
	IntervalSeconds: number;
	TimeoutSeconds: number;
	ExpectedStatusCode?: number;
	HTTPMethod?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'HEAD';
	HTTPHeaders?: string;
	HTTPBody?: string;
	OrganizationID: string;
	Organization?: { ID: string; Name: string; Slug: string };
	CreatedAt: string;
	UpdatedAt: string;
	LastCheckedAt?: string;
	ResponseTimeMs?: number;
	UptimePercent?: number;
	FailureCount: number;
	IsActive: boolean;
	Criticality: 'extreme' | 'high' | 'medium' | 'low';
	EnableSSLDomainCheck?: boolean;
	SSLExpiresAt?: string;
	SSLDaysLeft?: number;
	SSLStatus?: 'none' | 'valid' | 'warning' | 'expired' | 'error';
	SSLError?: string;
	DomainExpiresAt?: string;
	DomainDaysLeft?: number;
	DomainStatus?: 'none' | 'valid' | 'warning' | 'expired' | 'error';
	DomainError?: string;
	MonitorPICs?: { UserID: string }[];
	Assignees?: { UserID: string }[];
	AlertTemplateID?: string | null;
	recent_checks?: RecentCheck[];
}

function mapMonitor(m: any): Monitor {
	if (!m) return m;
	const assignees = (m.assignees || m.monitor_pics || m.MonitorPICs || []).map((pic: any) => ({
		UserID: pic.user_id ?? pic.UserID ?? pic.id
	}));
	return {
		ID: m.id ?? m.ID,
		Name: m.name ?? m.Name,
		Type: m.type ?? m.Type,
		Target: m.target ?? m.Target,
		Status: m.status ?? m.Status,
		IntervalSeconds: m.interval_seconds ?? m.IntervalSeconds,
		TimeoutSeconds: m.timeout_seconds ?? m.TimeoutSeconds,
		ExpectedStatusCode:
			m.expectation_status_code ?? m.expectation_status_code ?? m.ExpectedStatusCode,
		HTTPMethod: m.http_method ?? m.HTTPMethod,
		HTTPHeaders: m.http_headers ?? m.HTTPHeaders,
		HTTPBody: m.http_body ?? m.HTTPBody,
		OrganizationID: m.organization_id ?? m.OrganizationID,
		Organization: m.organization
			? {
					ID: m.organization.id ?? m.organization.ID,
					Name: m.organization.name ?? m.organization.Name,
					Slug: m.organization.slug ?? m.organization.Slug
				}
			: m.Organization,
		CreatedAt: m.created_at ?? m.CreatedAt,
		UpdatedAt: m.updated_at ?? m.UpdatedAt,
		LastCheckedAt: m.last_checked_at ?? m.LastCheckedAt,
		ResponseTimeMs: m.response_time_ms ?? m.ResponseTimeMs,
		UptimePercent: m.uptime_percent ?? m.UptimePercent,
		FailureCount: m.failure_count ?? m.FailureCount ?? 0,
		IsActive: m.is_active ?? m.IsActive ?? true,
		Criticality: m.criticality ?? m.Criticality,
		EnableSSLDomainCheck: m.enable_ssl_domain_check ?? m.EnableSSLDomainCheck ?? true,
		SSLExpiresAt: m.ssl_expires_at ?? m.SSLExpiresAt,
		SSLDaysLeft: m.ssl_days_left ?? m.SSLDaysLeft,
		SSLStatus: m.ssl_status ?? m.SSLStatus,
		SSLError: m.ssl_error ?? m.SSLError,
		DomainExpiresAt: m.domain_expires_at ?? m.DomainExpiresAt,
		DomainDaysLeft: m.domain_days_left ?? m.DomainDaysLeft,
		DomainStatus: m.domain_status ?? m.DomainStatus,
		DomainError: m.domain_error ?? m.DomainError,
		MonitorPICs: assignees,
		Assignees: assignees,
		AlertTemplateID: m.alert_template_id ?? m.AlertTemplateID ?? null,
		recent_checks: (m.recent_checks || []).map((c: any) => ({
			ID: c.id ?? c.ID,
			MonitorID: c.monitor_id ?? c.MonitorID,
			Status: c.status ?? c.Status,
			StatusCode: c.status_code ?? c.StatusCode,
			ResponseTimeMs: c.response_time_ms ?? c.ResponseTimeMs,
			ErrorMessage: c.error_message ?? c.ErrorMessage,
			ResponseSnippet: c.response_snippet ?? c.ResponseSnippet,
			CheckedAt: c.checked_at ?? c.CheckedAt
		}))
	};
}

export const MonitorAPI = {
	getAll: async (orgId: string) => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/monitors`, { method: 'GET' });
		if (!res.ok) await handleApiError(res, 'Failed to fetch monitors');
		const json = await res.json();
		return {
			...json,
			data: (json.data || []).map(mapMonitor)
		};
	},

	getById: async (orgId: string, monitorId: string) => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/monitors/${monitorId}`, {
			method: 'GET'
		});
		if (!res.ok) await handleApiError(res, 'Failed to fetch monitor');
		const json = await res.json();
		return {
			...json,
			data: json.data ? mapMonitor(json.data) : json.data
		};
	},

	create: async (orgId: string, data: CreateMonitorRequest) => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/monitors`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(data)
		});
		if (!res.ok) await handleApiError(res, 'Failed to create monitor');
		const json = await res.json();
		return {
			...json,
			data: json.data ? mapMonitor(json.data) : json.data
		};
	},

	update: async (orgId: string, monitorId: string, data: UpdateMonitorRequest) => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/monitors/${monitorId}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(data)
		});
		if (!res.ok) await handleApiError(res, 'Failed to update monitor');
		const json = await res.json();
		return {
			...json,
			data: json.data ? mapMonitor(json.data) : json.data
		};
	},

	delete: async (orgId: string, monitorId: string) => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/monitors/${monitorId}`, {
			method: 'DELETE'
		});
		if (!res.ok) await handleApiError(res, 'Failed to delete monitor');
		return res.json();
	},

	getDailyStats: async (orgId: string, monitorId: string, days = 30) => {
		const res = await apiFetch(
			`${API_Route}/organization/${orgId}/monitors/${monitorId}/stats?days=${days}`
		);
		if (!res.ok) await handleApiError(res, 'Failed to fetch monitor daily stats');
		return res.json();
	},

	getFleetOverview: async (orgId: string, windowDays: number = 30) => {
		const res = await apiFetch(
			`${API_Route}/organization/${orgId}/monitors/overview?window=${windowDays}`,
			{ method: 'GET' }
		);
		if (!res.ok) await handleApiError(res, 'Failed to fetch fleet overview');
		return await res.json();
	},

	triggerCheck: async (orgId: string, monitorId: string) => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/monitors/${monitorId}/check`, {
			method: 'POST'
		});
		if (!res.ok) await handleApiError(res, 'Failed to trigger monitor check');
		return res.json();
	},

	applyTemplate: async (orgId: string, monitorId: string, templateId: string | null) => {
		const res = await apiFetch(
			`${API_Route}/organization/${orgId}/monitors/${monitorId}/template`,
			{
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ template_id: templateId })
			}
		);
		if (!res.ok) await handleApiError(res, 'Failed to update monitor template');
		return res.json();
	}
};
