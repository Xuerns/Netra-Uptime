import type { PageLoad } from './$types';
import { OrgAPI } from '$lib/api/organization';

export const load: PageLoad = async () => {
	try {
		const res = await OrgAPI.getAll();
		const orgs = res.data ?? [];
		return { orgs };
	} catch {
		return { orgs: [] };
	}
};
