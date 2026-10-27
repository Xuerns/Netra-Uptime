import { env } from '$env/dynamic/public';
import { logger } from '$lib/utils/logger';

interface FetchOptions extends RequestInit {
	redirectOnFail?: boolean;
}

interface RefreshResult {
	ok: boolean;
	status: number;
}

// Single-flight promise so parallel 401 requests share a single refresh attempt
let refreshPromise: Promise<RefreshResult> | null = null;

async function executeRefresh(fetchFn: typeof fetch): Promise<RefreshResult> {
	try {
		const refreshResponse = await fetchFn(`${env.PUBLIC_API_URL}/auth/refresh`, {
			method: 'POST',
			credentials: 'include'
		});
		return { ok: refreshResponse.ok, status: refreshResponse.status };
	} catch (error) {
		logger.error('Failed to refresh token:', error);
		return { ok: false, status: 0 };
	}
}

export async function apiFetch(
	url: string,
	options: FetchOptions = {},
	customFetch: typeof fetch = typeof window !== 'undefined' ? window.fetch : fetch
): Promise<Response> {
	const { redirectOnFail = true, ...restOptions } = options;
	const fetchOptions: RequestInit = {
		...restOptions,
		credentials: restOptions.credentials || 'include'
	};

	let response = await customFetch(url, fetchOptions);

	// If access token is expired, backend returns 401.
	// Never attempt refresh on auth endpoints (e.g. login, refresh) to prevent loops.
	if (response.status === 401 && !url.includes('/auth/refresh') && !url.includes('/auth/login')) {
		if (!refreshPromise) {
			refreshPromise = executeRefresh(customFetch).finally(() => {
				refreshPromise = null;
			});
		}

		const refreshResult = await refreshPromise;

		if (refreshResult.ok) {
			// Token refreshed successfully; retry the original request with new cookies
			response = await customFetch(url, fetchOptions);
			if (response.status === 401 && typeof window !== 'undefined' && redirectOnFail) {
				window.location.href = '/auth/login';
			}
		} else {
			// Only redirect to login if the refresh token is truly expired/invalid (401).
			// Do NOT redirect on transient errors like 429 (Rate Limit) or 5xx/network errors.
			// Do NOT call /auth/logout here because it could accidentally revoke a valid session.
			if (refreshResult.status === 401 && typeof window !== 'undefined' && redirectOnFail) {
				window.location.href = '/auth/login';
			}
		}
	}

	return response;
}
