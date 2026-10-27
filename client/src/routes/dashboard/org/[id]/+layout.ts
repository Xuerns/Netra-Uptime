import { error } from '@sveltejs/kit';
import type { LayoutLoad } from './$types';
import { OrgAPI } from '$lib/api/organization';

export const load: LayoutLoad = async ({ params }) => {
	const orgId = params.id;

	try {
		// Fetch org list to find current org details
		const orgsRes = await OrgAPI.getAll();
		const allOrgs = orgsRes.data ?? [];
		const currentOrg = allOrgs.find((o: { ID?: string }) => o.ID === orgId);

		if (!currentOrg) {
			throw error(404, 'Organization not found');
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
	} catch (err: unknown) {
		if (err && typeof err === 'object' && 'status' in err) throw err;
		const message =
			err instanceof Error
				? err.message
				: typeof err === 'string'
					? err
					: 'Failed to load organization';
		throw error(500, message);
	}
};
