import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import { OrgAPI } from '$lib/api/organization';

export const load: PageLoad = async ({ params }) => {
	try {
		const res = await OrgAPI.getAll();
		const orgs = res.data || [];
		if (orgs.length > 0) {
			const lastOrgId =
				typeof window !== 'undefined' ? localStorage.getItem('pulsewatch_last_org') : null;
			const targetOrg = (lastOrgId && orgs.find((o: any) => o.ID === lastOrgId)) || orgs[0];
			throw redirect(307, `/dashboard/org/${targetOrg.ID}/incidents/${params.id}`);
		}
	} catch (err: any) {
		if (err?.status === 307 || err?.status === 302) throw err;
	}
	throw redirect(307, '/dashboard/org');
};
