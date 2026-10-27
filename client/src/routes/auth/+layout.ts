export const ssr = false;
import { redirect } from '@sveltejs/kit';
import { env } from '$env/dynamic/public';
import { apiFetch } from '$lib/api/client';
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
			throw redirect(302, '/dashboard/org');
		}
	} catch (error: unknown) {
		if (error && typeof error === 'object' && 'status' in error && error.status === 302) {
			throw error;
		}
		// ignore
	}

	return {};
};
