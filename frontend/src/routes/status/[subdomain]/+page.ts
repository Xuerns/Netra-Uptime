import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import { env } from '$env/dynamic/public';

export const load: PageLoad = async ({ params, fetch }) => {
	const res = await fetch(`${env.PUBLIC_API_URL}/public/statuspage/${params.subdomain}`);

	if (!res.ok) {
		throw error(404, 'Status Page not found');
	}

	const result = await res.json();
	return {
		statusPage: result.data
	};
};
