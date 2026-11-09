import { apiFetch, handleApiError } from './client';
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

		if (!res.ok) await handleApiError(res, 'Login failed');
		return res.json();
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

		if (!res.ok) await handleApiError(res, 'Registration failed');
		return res.json();
	},

	async logout(): Promise<void> {
		const res = await apiFetch(`${API_Route}/auth/logout`, {
			method: 'POST',
			redirectOnFail: false
		});
		if (!res.ok) await handleApiError(res, 'Failed to logout');
	},

	async getProfile(): Promise<UserProfile> {
		const res = await apiFetch(`${API_Route}/auth/profile`);
		if (!res.ok) await handleApiError(res, 'Failed to fetch profile');
		const data = await res.json();
		return data.user ?? data;
	},

	async updateProfile(payload: UpdateProfileRequest): Promise<UpdateProfileResponse> {
		const res = await apiFetch(`${API_Route}/auth/profile`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		if (!res.ok) await handleApiError(res, 'Failed to update profile');
		return res.json();
	},

	async initPasswordChange(payload: InitPasswordChangeRequest): Promise<AuthMessageResponse> {
		const res = await apiFetch(`${API_Route}/auth/password/change/init`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		if (!res.ok) await handleApiError(res, 'Failed to initiate password change');
		return res.json();
	},

	async verifyPasswordChange(payload: VerifyPasswordChangeRequest): Promise<AuthMessageResponse> {
		const res = await apiFetch(`${API_Route}/auth/password/change/verify`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		if (!res.ok) await handleApiError(res, 'Failed to verify password change');
		return res.json();
	},

	async forgotPassword(email: string): Promise<{ message?: string }> {
		const res = await fetch(`${API_Route}/auth/forgotpassword`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ email })
		});
		if (!res.ok) await handleApiError(res, 'Failed to send reset link.');
		return res.json();
	},

	async resetPassword(token: string, password: string): Promise<{ message?: string }> {
		const res = await fetch(`${API_Route}/auth/resetpassword`, {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ token, password })
		});
		if (!res.ok) await handleApiError(res, 'Failed to reset password.');
		return res.json();
	}
};
