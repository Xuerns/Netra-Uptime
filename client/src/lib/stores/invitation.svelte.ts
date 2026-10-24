import { OrgAPI } from '$lib/api/organization';

class InvitationStore {
	pendingCount = $state<number>(0);
	hasPending = $derived(this.pendingCount > 0);
	isLoading = $state<boolean>(false);

	async fetch() {
		this.isLoading = true;
		try {
			const res = await OrgAPI.getInvitations();
			const list = Array.isArray(res?.data) ? res.data : Array.isArray(res) ? res : [];

			const pending = list.filter((inv: any) => {
				const status = (inv?.status ?? inv?.Status ?? '').toLowerCase();
				return status === 'pending';
			});
			this.pendingCount = pending.length;
		} catch {
			this.pendingCount = 0;
		} finally {
			this.isLoading = false;
		}
	}
}

export const invitationStore = new InvitationStore();
