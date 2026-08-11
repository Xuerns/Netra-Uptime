import { apiFetch } from '../helper/api.ts';
import { env } from '$env/dynamic/public';

const API_Route = `${env.PUBLIC_API_URL}`;

export interface CreateSubscriptionRequest {
	plan: string;
	email: string;
}

export const SubscriptionAPI = {
	getByUserId: async () => {
		const res = await apiFetch(`${API_Route}/subscription`, { method: 'GET' });
		if (!res.ok) throw new Error('Failed to fetch subscriptions');
		return res.json();
	},

	createTransaction: async (orgId: string, data: CreateSubscriptionRequest) => {
		const res = await apiFetch(`${API_Route}/organization/${orgId}/subscription`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(data)
		});
		if (!res.ok) throw new Error('Failed to create subscription transaction');
		return res.json();
	}
};
