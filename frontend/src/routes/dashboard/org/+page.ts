import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import { OrgAPI } from '$lib/api/organization';

export const load: PageLoad = async () => {
	try {
		const res = await OrgAPI.getAll();
		const orgs = res.data ?? [];

		// Auto-redirect if user has exactly 1 org
		if (orgs.length === 1) {
			throw redirect(307, `/dashboard/org/${orgs[0].ID}/monitors`);
		}

		// Check localStorage for last selected org
		if (typeof window !== 'undefined') {
			const lastOrgId = localStorage.getItem('pulsewatch_last_org');
			if (lastOrgId && orgs.some((o: any) => o.ID === lastOrgId)) {
				throw redirect(307, `/dashboard/org/${lastOrgId}/monitors`);
			}
		}

		return { orgs };
	} catch (err: any) {
		// Re-throw redirects
		if (err?.status === 307 || err?.status === 302) throw err;
		return { orgs: [] };
	}
};
