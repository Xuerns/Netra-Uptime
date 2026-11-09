import { writable, type Writable } from 'svelte/store';
import { browser } from '$app/environment';

export interface Organization {
	id: string;
	name: string;
	slug: string;
	role?: string;
}

// Stores the list of all organizations owned by the user
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

// Stores the ID of the organization currently active (selected) in the sidebar/dashboard
export const activeOrgId = createActiveOrgIdStore();
