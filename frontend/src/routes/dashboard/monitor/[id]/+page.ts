import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import { MonitorAPI } from '$lib/api/monitor';

export const load: PageLoad = async ({ params }) => {
	try {
		const res = await MonitorAPI.getAllByUser();
		const monitor = (res.data || []).find((m: any) => m.ID === params.id);
		if (monitor?.OrganizationID) {
			throw redirect(307, `/dashboard/org/${monitor.OrganizationID}/monitors/${params.id}`);
		}
	} catch (err: any) {
		if (err?.status === 307 || err?.status === 302) throw err;
	}
	throw redirect(307, '/dashboard/org');
};
