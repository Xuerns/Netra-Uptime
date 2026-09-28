export const ssr = false;
import { env } from '$env/dynamic/public';
import type { LayoutLoad } from './$types';
import { apiFetch } from '$lib/helper/api';

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
		if (import.meta.env.DEV) {
			console.error('Gagal mengambil data user:', error);
		}
	}

	// Nilai default jika gagal memuat (misal belum login)
	return {
		user: {
			name: 'Unknown User',
			email: 'unknown@example.com'
		}
	};
};
