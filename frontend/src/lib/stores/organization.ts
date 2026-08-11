import { writable, derived } from 'svelte/store';

export interface Organization {
	id: string;
	name: string;
	slug: string;
	role?: string; // Role dari user pada organisasi ini (owner, admin, member)
}

// Menyimpan daftar seluruh organisasi milik user
export const organizations = writable<Organization[]>([]);

// Menyimpan ID organisasi yang saat ini sedang aktif (dipilih) di sidebar/dashboard
export const activeOrgId = writable<string | null>(null);

// Store turunan (derived) yang secara otomatis akan mengembalikan objek Organization 
// secara penuh berdasarkan activeOrgId yang sedang terpilih.
export const activeOrganization = derived(
	[organizations, activeOrgId],
	([$organizations, $activeOrgId]) => {
		if (!$activeOrgId) return null;
		return $organizations.find(org => org.id === $activeOrgId) || null;
	}
);
