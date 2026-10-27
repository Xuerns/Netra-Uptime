import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import { StatusPageAPI } from '$lib/api/statuspage';

export const load: PageLoad = async ({ params, fetch }) => {
	try {
		const res = await StatusPageAPI.getPublic(params.subdomain, fetch);
		return {
			statusPage: res.data
		};
	} catch {
		throw error(404, 'Status Page not found');
	}
};
