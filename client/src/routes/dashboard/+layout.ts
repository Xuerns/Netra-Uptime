export const ssr = false;
import { env } from '$env/dynamic/public';
import type { LayoutLoad } from './$types';
import { apiFetch } from '$lib/api/client';
import { logger } from '$lib/utils/logger';

export const load: LayoutLoad = async ({ fetch }) => {
	try {
		const res = await apiFetch(
			`${env.PUBLIC_API_URL}/auth/profile`,
			{
				credentials: 'include'
			},
			fetch
		);

		if (res.ok) {
			const data = await res.json();

			return {
				user: data.user
			};
		}
	} catch (error) {
		logger.error('Failed to fetch user data:', error);
	}

	// Default values if loading fails (e.g. not logged in)
	return {
		user: {
			name: 'Unknown User',
			email: 'unknown@example.com'
		}
	};
};
