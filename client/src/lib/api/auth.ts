import { apiFetch } from './client';
import { env } from '$env/dynamic/public';

const API_Route = `${env.PUBLIC_API_URL}`;

export interface LoginRequest {
	email: string;
	password: string;
}

export interface RegisterRequest {
	name?: string;
	email: string;
	password: string;
}

export interface UpdateProfileRequest {
	name: string;
}

export interface InitPasswordChangeRequest {
	current_password: string;
}

export interface VerifyPasswordChangeRequest {
	code: string;
	new_password: string;
}

export interface OAuthAuthRequest {
	provider: string;
	code: string;
	state?: string;
}

export interface UserProfile {
	id: string;
	name: string;
	email: string;
	avatar_url?: string;
	created_at?: string;
}

export interface AuthMessageResponse {
	message?: string;
}

export interface UpdateProfileResponse {
	message?: string;
	user?: UserProfile;
}

export const AuthAPI = {
	async login(payload: LoginRequest): Promise<AuthMessageResponse> {
		const res = await fetch(`${API_Route}/auth/login`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload),
			credentials: 'include'
		});

		if (res.status === 429) {
			throw new Error('Too many attempts. Please wait a moment and try again.');
		}

		const data = await res.json();
		if (!res.ok) {
			throw new Error(data.error || data.message || 'Login failed');
		}
		return data;
	},

	async register(payload: RegisterRequest): Promise<AuthMessageResponse> {
		const res = await fetch(`${API_Route}/auth/register`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload),
			credentials: 'include'
		});

		if (res.status === 429) {
			throw new Error('Too many attempts. Please wait a moment and try again.');
		}

		const data = await res.json();
		if (!res.ok) {
			throw new Error(data.error || data.message || 'Registration failed');
		}
		return data;
	},

	async logout(): Promise<void> {
		const res = await apiFetch(`${API_Route}/auth/logout`, {
			method: 'POST',
			redirectOnFail: false
		});
		if (!res.ok) {
			const data = await res.json().catch(() => ({}));
			throw new Error(data.error || data.message || 'Failed to logout');
		}
	},

	async getOAuthState(): Promise<string> {
		try {
			const res = await fetch(`${API_Route}/auth/oauth/state`);
			if (res.ok) {
				const data = await res.json();
				if (data.state) return data.state;
			}
		} catch {
			// Fallback
		}
		return crypto.randomUUID();
	},

	async startGoogleOAuth(): Promise<void> {
		const clientId = env.PUBLIC_GOOGLE_CLIENT_ID ?? '';
		const redirectUri = env.PUBLIC_GOOGLE_REDIRECT_URL ?? '';
		const state = await this.getOAuthState();
		if (typeof sessionStorage !== 'undefined') {
			sessionStorage.setItem('oauth_state', state);
		}
		const params = new URLSearchParams({
			client_id: clientId,
			redirect_uri: redirectUri,
			response_type: 'code',
			scope: 'openid email profile',
			state: state,
			prompt: 'select_account'
		});
		if (typeof window !== 'undefined') {
			window.location.href = `https://accounts.google.com/o/oauth2/v2/auth?${params.toString()}`;
		}
	},

	async authenticateOAuth(
		payload: OAuthAuthRequest,
		signal?: AbortSignal
	): Promise<AuthMessageResponse> {
		const res = await fetch(`${API_Route}/auth`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload),
			credentials: 'include',
			signal
		});
		const data = await res.json();
		if (!res.ok) {
			throw new Error(data.error || data.message || 'Failed to authenticate with provider');
		}
		return data;
	},

	async getProfile(): Promise<UserProfile> {
		const res = await apiFetch(`${API_Route}/auth/profile`);
		if (!res.ok) {
			const err = await res.json().catch(() => ({}));
			throw new Error(err.error || 'Failed to fetch profile');
		}
		const data = await res.json();
		return data.user ?? data;
	},

	async updateProfile(payload: UpdateProfileRequest): Promise<UpdateProfileResponse> {
		const res = await apiFetch(`${API_Route}/auth/profile`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		if (!res.ok) {
			const err = await res.json().catch(() => ({}));
			throw new Error(err.error || 'Failed to update profile');
		}
		return res.json();
	},

	async initPasswordChange(payload: InitPasswordChangeRequest): Promise<AuthMessageResponse> {
		const res = await apiFetch(`${API_Route}/auth/password/change/init`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		if (!res.ok) {
			const err = await res.json().catch(() => ({}));
			throw new Error(err.error || 'Failed to initiate password change');
		}
		return res.json();
	},

	async verifyPasswordChange(payload: VerifyPasswordChangeRequest): Promise<AuthMessageResponse> {
		const res = await apiFetch(`${API_Route}/auth/password/change/verify`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		if (!res.ok) {
			const err = await res.json().catch(() => ({}));
			throw new Error(err.error || 'Failed to verify password change');
		}
		return res.json();
	},

	async forgotPassword(email: string): Promise<{ message?: string }> {
		const res = await fetch(`${API_Route}/auth/forgotpassword`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ email })
		});
		const data = await res.json();
		if (!res.ok) {
			throw new Error(data.error || data.message || 'Failed to send reset link.');
		}
		return data;
	},

	async resetPassword(token: string, password: string): Promise<{ message?: string }> {
		const res = await fetch(`${API_Route}/auth/resetpassword`, {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ token, password })
		});
		const data = await res.json();
		if (!res.ok) {
			throw new Error(data.error || data.message || 'Failed to reset password.');
		}
		return data;
	}
};
