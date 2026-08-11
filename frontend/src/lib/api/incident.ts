import { apiFetch } from '../helper/api';
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

    // Dummy UI fields to simulate future backend support
    Severity?: 'Sev-1 Critical' | 'Sev-2 High' | 'Sev-3 Medium';
    SeverityCode?: 'sev1' | 'sev2' | 'sev3';
    Assignee?: {
        Name: string;
        Avatar: string;
    };
}

export interface AddIncidentUpdateReq {
    message: string;
}

export function mapIncident(inc: any): Incident {
    return {
        ID: inc.id ?? inc.ID,
        MonitorID: inc.monitor_id ?? inc.MonitorID,
        Status: inc.status ?? inc.Status,
        Summary: inc.summary ?? inc.Summary,
        StartedAt: inc.started_at ?? inc.StartedAt,
        ResolvedAt: inc.resolved_at ?? inc.ResolvedAt,
        CreatedAt: inc.created_at ?? inc.CreatedAt,
        UpdatedAt: inc.updated_at ?? inc.UpdatedAt,
        Monitor: inc.monitor || inc.Monitor ? {
            Name: inc.monitor?.name ?? inc.Monitor?.Name,
            Target: inc.monitor?.target ?? inc.Monitor?.Target,
        } as any : undefined,
        Assignees: (inc.assignees ?? inc.Assignees ?? []).map((a: any) => ({
            ID: a.id ?? a.ID,
            IncidentID: a.incident_id ?? a.IncidentID,
            UserID: a.user_id ?? a.UserID,
            Role: a.role ?? a.Role,
            CreatedAt: a.created_at ?? a.CreatedAt,
            User: a.user || a.User ? {
                Name: a.user?.name ?? a.User?.Name,
                Avatar: a.user?.avatar ?? a.User?.Avatar
            } : undefined
        })),
        Severity: inc.Severity,
        SeverityCode: inc.SeverityCode,
        Assignee: inc.Assignee
    };
}

export const IncidentAPI = {
    getAll: async (orgId: string): Promise<{ data: Incident[] }> => {
        const res = await apiFetch(`${API_Route}/organization/${orgId}/incident`, { method: 'GET' });
        if (!res.ok) throw new Error('Failed to fetch incidents');
        
        const data = await res.json();
        
        // Inject dummy data for UI testing since backend doesn't support them yet
        if (data.data) {
            data.data = data.data.map((rawInc: any, idx: number) => {
                const inc = mapIncident(rawInc);
                const isCritical = inc.Status === 'ongoing' && idx === 0;
                const isHigh = inc.Status === 'ongoing' && idx === 1;
                
                return {
                    ...inc,
                    Severity: isCritical ? 'Sev-1 Critical' : (isHigh ? 'Sev-2 High' : 'Sev-3 Medium'),
                    SeverityCode: isCritical ? 'sev1' : (isHigh ? 'sev2' : 'sev3'),
                    Assignee: {
                        Name: isCritical ? 'Sarah J. (On-call)' : 'System Admin',
                        Avatar: 'https://lh3.googleusercontent.com/aida-public/AB6AXuCWWg0i3K32PRnzgcgm-KInGbQ_KP6Lyk4dAXO7rAlUNlY70n9HJxF4rRON6By7RRNBkI4NCvt-XomEBfoMyN7AyEiCoAtVox4UJL0rGFfYacyLXFBub_RmMrfJ92hfJvZNtqJ6ocqMHWVk_9y2XhftQPLQTrRhRoUyYoqlGLh4tycxX8fVS4KLzgbPn7ZDneOUKVQRbJ8zgu44APcNLEKB7kXjFRFg2bk_B0ciTQ5NZU3paIZBNS6NBw'
                    }
                };
            });
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

    getUpdates: async (orgId: string, incidentId: string): Promise<{ data: IncidentUpdateResponse[] }> => {
        const res = await apiFetch(`${API_Route}/organization/${orgId}/incident/${incidentId}/updates`, {
            method: 'GET'
        });
        if (!res.ok) throw new Error('Failed to fetch incident updates');
        return res.json();
    }
};
