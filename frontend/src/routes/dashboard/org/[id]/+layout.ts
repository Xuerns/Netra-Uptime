import { error } from '@sveltejs/kit';
import type { LayoutLoad } from './$types';
import { OrgAPI } from '$lib/api/organization';
import { activeOrgId, organizations } from '$lib/stores/organization';

export const load: LayoutLoad = async ({ params }) => {
	const orgId = params.id;

	try {
		// Fetch org list to find current org details
		const orgsRes = await OrgAPI.getAll();
		const allOrgs = orgsRes.data ?? [];
		const currentOrg = allOrgs.find((o: any) => o.ID === orgId);

		if (!currentOrg) {
			throw error(404, 'Organization not found');
		}

		// Update global stores
		organizations.set(allOrgs.map((o: any) => ({ id: o.ID, name: o.Name, slug: o.Slug })));
		activeOrgId.set(orgId);

		// Persist last selected org
		if (typeof window !== 'undefined') {
			localStorage.setItem('pulsewatch_last_org', orgId);
		}

		// Fetch members
		const membersRes = await OrgAPI.getMembers(orgId);
		const members = membersRes.data ?? [];

		return {
			org: currentOrg,
			allOrgs,
			members,
			orgId
		};
	} catch (err: any) {
		if (err?.status) throw err;
		throw error(500, err.message || 'Failed to load organization');
	}
};
