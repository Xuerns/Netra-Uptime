import { writable, type Writable } from 'svelte/store';
import { browser } from '$app/environment';

export interface Organization {
	id: string;
	name: string;
	slug: string;
	role?: string;
}

// Menyimpan daftar seluruh organisasi milik user
export const organizations = writable<Organization[]>([]);

function createActiveOrgIdStore(): Writable<string | null> {
	const initialOrgId = browser ? localStorage.getItem('netra_last_org') : null;
	const { subscribe, set: internalSet, update } = writable<string | null>(initialOrgId);

	function persist(value: string | null) {
		if (browser) {
			try {
				if (value) {
					localStorage.setItem('netra_last_org', value);
				} else {
					localStorage.removeItem('netra_last_org');
				}
			} catch {
				// Ignore storage errors
			}
		}
	}

	return {
		subscribe,
		set(value: string | null) {
			persist(value);
			internalSet(value);
		},
		update(fn: (value: string | null) => string | null) {
			update((curr) => {
				const next = fn(curr);
				persist(next);
				return next;
			});
		}
	};
}

// Menyimpan ID organisasi yang saat ini sedang aktif (dipilih) di sidebar/dashboard
export const activeOrgId = createActiveOrgIdStore();
