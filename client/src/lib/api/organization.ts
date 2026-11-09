import { apiFetch } from './client';
import { env } from '$env/dynamic/public';

const API_Route = `${env.PUBLIC_API_URL}`;

export interface CreateOrgRequest {
	name: string;
	slug: string;
	invitations?: Array<{ invitee_id: string; role?: string }>;
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

export interface OrganizationUser {
	ID: string;
	Name: string;
	AvatarURL?: string;
}

export interface OrganizationMember {
	ID: string;
	OrganizationID: string;
	UserID: string;
	Role: string;
	User: OrganizationUser;
}

export interface OrganizationItem {
	ID: string;
	Name: string;
	name?: string;
	Slug: string;
	slug?: string;
	CreatedAt: string;
	created_at?: string;
	TotalMonitors?: number;
	ActiveMonitors?: number;
	TotalMember?: number;
	OrganizationMembers?: Array<{
		ID?: string;
		UserID?: string;
		Role?: string;
		User?: {
			ID?: string;
			Name?: string;
			AvatarURL?: string;
		};
	}>;
	Monitors?: Array<{ Status: string }>;
	role?: string;
}

export interface OrganizationInvitationItem {
	id?: string;
	ID?: string;
	role?: string;
	Role?: string;
	status?: string;
	Status?: string;
	organization?: {
		id?: string;
		name?: string;
		slug?: string;
	};
	Organization?: {
		name?: string;
		slug?: string;
	};
	inviter?: {
		id?: string;
		name?: string;
		avatar_url?: string | null;
	};
	Inviter?: {
		name?: string;
		avatar_url?: string | null;
	};
}

export interface SearchPersonResult {
	id: string;
	name: string;
	avatar_url?: string | null;
}

export interface OrgActionResponse {
	status: string;
	message?: string;
}

export interface OrgListResponse {
	status: string;
	data: OrganizationItem[];
}

export interface OrgDetailResponse {
	status: string;
	data: OrganizationItem;
}

export interface OrgMembersResponse {
	status: string;
	data: OrganizationMember[];
}

export interface OrgInvitationsResponse {
	status: string;
	data: OrganizationInvitationItem[];
}

export interface OrgSearchPersonsResponse {
	status?: string;
	data: SearchPersonResult[];
	pagination?: {
		has_more?: boolean;
		page?: number;
		limit?: number;
		total?: number;
	};
}

function mapMember(m: any): OrganizationMember {
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

function mapOrganization(org: any): OrganizationItem {
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
	getAll: async (): Promise<OrgListResponse> => {
		const res = await apiFetch(`${API_Route}/organization/all`, { method: 'GET' });
		if (!res.ok) throw new Error('Failed to fetch organizations');
		const json = await res.json();
		return {
			...json,
			data: (json.data || []).map(mapOrganization)
		};
	},

	create: async (data: CreateOrgRequest): Promise<OrgDetailResponse> => {
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

	getMembers: async (orgId: string): Promise<OrgMembersResponse> => {
		// organizationID is extracted from the path `:orgId` and validated by tenant middleware
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

	inviteMember: async (
		orgId: string,
		data: InviteMemberRequest | InviteMemberRequest[]
	): Promise<OrgActionResponse> => {
		// Backend `InviteMember` (BulkInviteMemberRequest) expects: { invitations: [...] }
		const invitations = Array.isArray(data) ? data : [data];
		const res = await apiFetch(`${API_Route}/organization/${orgId}/invite`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ invitations })
		});
		if (!res.ok) {
			const err = await res.json().catch(() => ({}));
			throw new Error(err.error || err.message || 'Failed to invite member');
		}
		return res.json();
	},

	kickMember: async (orgId: string, memberId: string): Promise<OrgActionResponse> => {
		// In the backend `KickMember`, directly bound to `var memberId string` instead of `KickMemberRequest`
		const res = await apiFetch(`${API_Route}/organization/${orgId}/kick`, {
			method: 'DELETE',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(memberId)
		});
		if (!res.ok) throw new Error('Failed to kick member');
		return res.json();
	},

	update: async (orgId: string, data: UpdateOrgRequest): Promise<OrgDetailResponse> => {
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

	disband: async (orgId: string): Promise<OrgActionResponse> => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/disband`, {
			method: 'DELETE'
		});
		if (!res.ok) throw new Error('Failed to disband organization');
		return res.json();
	},

	getInvitations: async (): Promise<OrgInvitationsResponse> => {
		const res = await apiFetch(`${API_Route}/invitation`, { method: 'GET' });
		if (!res.ok) {
			const errText = await res.text();
			throw new Error(`Failed to fetch invitations: ${res.status} ${errText}`);
		}
		return res.json();
	},

	actionInvitation: async (
		invitationId: string,
		action: 'accept' | 'reject'
	): Promise<OrgActionResponse> => {
		const res = await apiFetch(`${API_Route}/invitation/action`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ invitation_id: invitationId, action })
		});
		if (!res.ok) throw new Error('Failed to process invitation');
		return res.json();
	},

	searchPersons: async (
		name: string,
		page: number = 1,
		orgId?: string
	): Promise<OrgSearchPersonsResponse> => {
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

	promoteToAdmin: async (orgId: string, userId: string): Promise<OrgActionResponse> => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/promote`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ user_id: userId })
		});
		if (!res.ok) throw new Error('Failed to promote user to admin');
		return res.json();
	},

	demoteToMember: async (orgId: string, userId: string): Promise<OrgActionResponse> => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/demote`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ user_id: userId })
		});
		if (!res.ok) throw new Error('Failed to demote user to member');
		return res.json();
	},

	transferOwnership: async (orgId: string, userId: string): Promise<OrgActionResponse> => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/transfer`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ user_id: userId })
		});
		if (!res.ok) throw new Error('Failed to transfer ownership');
		return res.json();
	},

	leaveOrganization: async (orgId: string): Promise<OrgActionResponse> => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/leave`, {
			method: 'PUT'
		});
		if (!res.ok) throw new Error('Failed to leave organization');
		return res.json();
	}
};
