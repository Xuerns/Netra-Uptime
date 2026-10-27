import { writable, type Writable } from 'svelte/store';
import { browser } from '$app/environment';

const STORAGE_KEY = 'netra_timezone';

function createUserTimezoneStore(): Writable<string> {
	const initial = browser ? (localStorage.getItem(STORAGE_KEY) ?? 'UTC') : 'UTC';
	const { subscribe, set: internalSet, update } = writable<string>(initial);

	function persist(value: string) {
		if (browser) {
			try {
				localStorage.setItem(STORAGE_KEY, value);
			} catch {
				// Ignore storage errors
			}
		}
	}

	return {
		subscribe,
		set(value: string) {
			persist(value);
			internalSet(value);
		},
		update(fn: (value: string) => string) {
			update((curr) => {
				const next = fn(curr);
				persist(next);
				return next;
			});
		}
	};
}

/** Svelte store yang menyimpan timezone aktif user (IANA timezone name). */
export const userTimezone = createUserTimezoneStore();

/**
 * Daftar timezone populer yang ditampilkan di User Settings.
 * Format: { value: IANA name, label: display label, region: grouping }
 */
export const POPULAR_TIMEZONES = [
	// UTC
	{ value: 'UTC', label: 'UTC — Coordinated Universal Time', region: 'Global' },

	// Americas
	{ value: 'America/New_York', label: 'America/New_York — ET (UTC−5/−4)', region: 'Americas' },
	{ value: 'America/Chicago', label: 'America/Chicago — CT (UTC−6/−5)', region: 'Americas' },
	{ value: 'America/Denver', label: 'America/Denver — MT (UTC−7/−6)', region: 'Americas' },
	{
		value: 'America/Los_Angeles',
		label: 'America/Los_Angeles — PT (UTC−8/−7)',
		region: 'Americas'
	},
	{ value: 'America/Sao_Paulo', label: 'America/Sao_Paulo — BRT (UTC−3)', region: 'Americas' },
	{ value: 'America/Toronto', label: 'America/Toronto — ET (UTC−5/−4)', region: 'Americas' },
	{ value: 'America/Vancouver', label: 'America/Vancouver — PT (UTC−8/−7)', region: 'Americas' },

	// Europe
	{ value: 'Europe/London', label: 'Europe/London — GMT/BST (UTC+0/+1)', region: 'Europe' },
	{ value: 'Europe/Paris', label: 'Europe/Paris — CET/CEST (UTC+1/+2)', region: 'Europe' },
	{ value: 'Europe/Berlin', label: 'Europe/Berlin — CET/CEST (UTC+1/+2)', region: 'Europe' },
	{ value: 'Europe/Amsterdam', label: 'Europe/Amsterdam — CET/CEST (UTC+1/+2)', region: 'Europe' },
	{ value: 'Europe/Madrid', label: 'Europe/Madrid — CET/CEST (UTC+1/+2)', region: 'Europe' },
	{ value: 'Europe/Rome', label: 'Europe/Rome — CET/CEST (UTC+1/+2)', region: 'Europe' },
	{ value: 'Europe/Moscow', label: 'Europe/Moscow — MSK (UTC+3)', region: 'Europe' },
	{ value: 'Europe/Istanbul', label: 'Europe/Istanbul — TRT (UTC+3)', region: 'Europe' },

	// Africa & Middle East
	{ value: 'Africa/Cairo', label: 'Africa/Cairo — EET (UTC+2)', region: 'Africa & Middle East' },
	{ value: 'Asia/Dubai', label: 'Asia/Dubai — GST (UTC+4)', region: 'Africa & Middle East' },
	{ value: 'Asia/Riyadh', label: 'Asia/Riyadh — AST (UTC+3)', region: 'Africa & Middle East' },

	// Asia
	{ value: 'Asia/Kolkata', label: 'Asia/Kolkata — IST (UTC+5:30)', region: 'Asia' },
	{ value: 'Asia/Dhaka', label: 'Asia/Dhaka — BST (UTC+6)', region: 'Asia' },
	{ value: 'Asia/Bangkok', label: 'Asia/Bangkok — ICT (UTC+7)', region: 'Asia' },
	{ value: 'Asia/Jakarta', label: 'Asia/Jakarta — WIB (UTC+7)', region: 'Asia' },
	{ value: 'Asia/Makassar', label: 'Asia/Makassar — WITA (UTC+8)', region: 'Asia' },
	{ value: 'Asia/Jayapura', label: 'Asia/Jayapura — WIT (UTC+9)', region: 'Asia' },
	{ value: 'Asia/Singapore', label: 'Asia/Singapore — SGT (UTC+8)', region: 'Asia' },
	{ value: 'Asia/Kuala_Lumpur', label: 'Asia/Kuala_Lumpur — MYT (UTC+8)', region: 'Asia' },
	{ value: 'Asia/Hong_Kong', label: 'Asia/Hong_Kong — HKT (UTC+8)', region: 'Asia' },
	{ value: 'Asia/Shanghai', label: 'Asia/Shanghai — CST (UTC+8)', region: 'Asia' },
	{ value: 'Asia/Taipei', label: 'Asia/Taipei — CST (UTC+8)', region: 'Asia' },
	{ value: 'Asia/Seoul', label: 'Asia/Seoul — KST (UTC+9)', region: 'Asia' },
	{ value: 'Asia/Tokyo', label: 'Asia/Tokyo — JST (UTC+9)', region: 'Asia' },

	// Australia & Pacific
	{ value: 'Australia/Perth', label: 'Australia/Perth — AWST (UTC+8)', region: 'Pacific' },
	{
		value: 'Australia/Sydney',
		label: 'Australia/Sydney — AEDT/AEST (UTC+11/+10)',
		region: 'Pacific'
	},
	{ value: 'Pacific/Auckland', label: 'Pacific/Auckland — NZST (UTC+12)', region: 'Pacific' },
	{ value: 'Pacific/Honolulu', label: 'Pacific/Honolulu — HST (UTC−10)', region: 'Pacific' }
] as const;
