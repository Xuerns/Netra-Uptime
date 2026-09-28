import { env } from '$env/dynamic/public';

interface FetchOptions extends RequestInit {
	redirectOnFail?: boolean;
}

export async function apiFetch(
	url: string,
	options: FetchOptions = {},
	customFetch: typeof fetch = window.fetch
): Promise<Response> {
	const { redirectOnFail = true, ...restOptions } = options;
	const fetchOptions: RequestInit = {
		...restOptions,
		credentials: restOptions.credentials || 'include'
	};

	let response = await customFetch(url, fetchOptions);

	// If access token is expired, backend returns 401
	if (response.status === 401) {
		try {
			// Call refresh token endpoint
			const refreshResponse = await customFetch(`${env.PUBLIC_API_URL}/auth/refresh`, {
				method: 'POST',
				credentials: 'include'
			});

			if (refreshResponse.ok) {
				// Token refreshed successfully, try again
				response = await customFetch(url, fetchOptions);
			} else {
				// If Refresh token is expired or invalid
				if (typeof window !== 'undefined' && redirectOnFail) {
					await customFetch(`${env.PUBLIC_API_URL}/auth/logout`, {
						method: 'POST',
						credentials: 'include'
					});
					window.location.href = '/auth/login';
				}
			}
		} catch (error) {
			if (import.meta.env.DEV) {
				console.error('Failed to refresh token:', error);
			}
			if (typeof window !== 'undefined' && redirectOnFail) {
				window.location.href = '/auth/login';
			}
		}
	}

	return response;
}
