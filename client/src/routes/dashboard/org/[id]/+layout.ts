import { error } from '@sveltejs/kit';
import type { LayoutLoad } from './$types';
import { OrgAPI } from '$lib/api/organization';
import { getCurrentUserId, findMemberByUserId, getMemberRole, canManageOrg } from '$lib/utils/role';

export const load: LayoutLoad = async ({ params, parent }) => {
	const orgId = params.id;

	try {
		const { user } = await parent();

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

		const currentMember = findMemberByUserId(members, getCurrentUserId(user));
		const currentUserRole = getMemberRole(currentMember, 'member');
		const canManage = canManageOrg(currentUserRole);
		const isOwner = currentUserRole === 'owner';

		return {
			org: currentOrg,
			allOrgs,
			members,
			orgId,
			currentMember,
			currentUserRole,
			canManage,
			isOwner
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
