export const ssr = false;
import { redirect } from '@sveltejs/kit';
import { env } from '$env/dynamic/public';
import { apiFetch } from '$lib/helper/api';
import type { LayoutLoad } from './$types';

export const load: LayoutLoad = async ({ fetch }) => {
	try {
		const res = await apiFetch(
			`${env.PUBLIC_API_URL}/auth/profile`,
			{
				credentials: 'include',
				redirectOnFail: false
			},
			fetch
		);

		if (res.ok) {
			throw redirect(302, '/dashboard/monitor');
		}
	} catch (error: any) {
		if (error?.status === 302) {
			throw error;
		}
		// ignore
	}

	return {};
};
