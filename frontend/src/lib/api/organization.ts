import { apiFetch } from '../helper/api.ts';
import { env } from '$env/dynamic/public';

const API_Route = `${env.PUBLIC_API_URL}`;

export interface CreateOrgRequest {
	name: string;
	slug: string;
}

export interface UpdateOrgRequest {
	organization_id: string;
	name?: string;
	slug?: string;
}

export interface InviteMemberRequest {
	invitee_id: string;
	role?: string;
}

function mapMember(m: any): any {
	if (!m) return m;
	const user = m.user ?? m.User ?? {};
	return {
		ID: m.id ?? m.ID,
		OrganizationID: m.organization_id ?? m.OrganizationID,
		UserID: m.user_id ?? m.UserID,
		Role: m.role ?? m.Role,
		User: {
			ID: user.id ?? user.ID ?? m.user_id ?? m.UserID,
			Name: user.name ?? user.Name,
			AvatarURL: user.avatar_url ?? user.AvatarURL
		}
	};
}

function mapOrganization(org: any): any {
	if (!org) return org;
	return {
		ID: org.id ?? org.ID,
		Name: org.name ?? org.Name,
		Slug: org.slug ?? org.Slug,
		CreatedAt: org.created_at ?? org.CreatedAt,
		TotalMonitors: org.total_monitors ?? org.TotalMonitors,
		ActiveMonitors: org.active_monitors ?? org.ActiveMonitors,
		TotalMember: org.total_member ?? org.TotalMember,
		// Provide Fallback arrays if UI relies on .length or .filter
		OrganizationMembers:
			org.organization_members ??
			org.OrganizationMembers ??
			new Array(org.total_member || 0).fill({}),
		Monitors:
			org.monitors ??
			org.Monitors ??
			Array.from({ length: org.total_monitors || 0 }, (_, i) => ({
				Status: i < (org.active_monitors ?? 0) ? 'up' : 'down'
			}))
	};
}

export const OrgAPI = {
	getAll: async () => {
		const res = await apiFetch(`${API_Route}/organization/all`, { method: 'GET' });
		if (!res.ok) throw new Error('Failed to fetch organizations');
		const json = await res.json();
		return {
			...json,
			data: (json.data || []).map(mapOrganization)
		};
	},

	create: async (data: CreateOrgRequest) => {
		const res = await apiFetch(`${API_Route}/organization/create`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(data)
		});
		if (!res.ok) throw new Error('Failed to create organization');
		const json = await res.json();
		return {
			...json,
			data: json.data ? mapOrganization(json.data) : json.data
		};
	},

	getMembers: async (orgId: string) => {
		// organizationID diambil dari path `:orgId` dan divalidasi oleh tenant middleware
		const res = await apiFetch(`${API_Route}/organization/${orgId}/members`, {
			method: 'GET'
		});
		if (!res.ok) throw new Error('Failed to fetch members');
		const json = await res.json();
		return {
			...json,
			data: (json.data || []).map(mapMember)
		};
	},

	inviteMember: async (orgId: string, data: InviteMemberRequest) => {
		// Backend `InviteMember` (BulkInviteMemberRequest) expects: { invitations: [...] }
		const res = await apiFetch(`${API_Route}/organization/${orgId}/invite`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ invitations: [data] })
		});
		if (!res.ok) throw new Error('Failed to invite member');
		return res.json();
	},

	kickMember: async (orgId: string, memberId: string) => {
		// Di backend `KickMember`, di-bind langsung ke `var memberId string` alih-alih `KickMemberRequest`
		const res = await apiFetch(`${API_Route}/organization/${orgId}/kick`, {
			method: 'DELETE',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(memberId)
		});
		if (!res.ok) throw new Error('Failed to kick member');
		return res.json();
	},

	update: async (orgId: string, data: UpdateOrgRequest) => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/update`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(data)
		});
		if (!res.ok) throw new Error('Failed to update organization');
		const json = await res.json();
		return {
			...json,
			data: json.data ? mapOrganization(json.data) : json.data
		};
	},

	disband: async (orgId: string) => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/disband`, {
			method: 'DELETE'
		});
		if (!res.ok) throw new Error('Failed to disband organization');
		return res.json();
	},

	getInvitations: async () => {
		const res = await apiFetch(`${API_Route}/invitation`, { method: 'GET' });
		if (!res.ok) {
			const errText = await res.text();
			throw new Error(`Failed to fetch invitations: ${res.status} ${errText}`);
		}
		return res.json();
	},

	actionInvitation: async (invitationId: string, action: 'accept' | 'reject') => {
		const res = await apiFetch(`${API_Route}/invitation/action`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ invitation_id: invitationId, action })
		});
		if (!res.ok) throw new Error('Failed to process invitation');
		return res.json();
	},

	searchPersons: async (name: string, page: number = 1, orgId?: string) => {
		let url = `${API_Route}/search/persons?name=${encodeURIComponent(name)}&page=${page}`;
		if (orgId) {
			url += `&orgId=${orgId}`;
		}

		const res = await apiFetch(url, { method: 'GET' });
		if (!res.ok) {
			if (res.status === 404) return { data: [] }; // Handle No users found
			throw new Error('Failed to search persons');
		}
		return res.json();
	},

	promoteToAdmin: async (orgId: string, userId: string) => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/promote`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ user_id: userId })
		});
		if (!res.ok) throw new Error('Failed to promote user to admin');
		return res.json();
	},

	demoteToMember: async (orgId: string, userId: string) => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/demote`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ user_id: userId })
		});
		if (!res.ok) throw new Error('Failed to demote user to member');
		return res.json();
	},

	transferOwnership: async (orgId: string, userId: string) => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/transfer`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ user_id: userId })
		});
		if (!res.ok) throw new Error('Failed to transfer ownership');
		return res.json();
	},

	leaveOrganization: async (orgId: string) => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/leave`, {
			method: 'PUT'
		});
		if (!res.ok) throw new Error('Failed to leave organization');
		return res.json();
	}
};
