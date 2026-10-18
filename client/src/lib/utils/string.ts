/**
 * String and naming utilities for Netra Uptime / PulseWatch.
 */

export const DEFAULT_AVATAR_COLORS = [
	'bg-emerald-500/20 text-emerald-400',
	'bg-blue-500/20 text-blue-400',
	'bg-violet-500/20 text-violet-400',
	'bg-amber-500/20 text-amber-400',
	'bg-rose-500/20 text-rose-400',
	'bg-cyan-500/20 text-cyan-400'
];

/**
 * Extract 1-2 character uppercase initials from a name string.
 * Examples:
 *   "John Doe" -> "JD"
 *   "Acme" -> "AC"
 *   "" -> "U"
 */
export function getInitials(name?: string | null, fallback = 'U'): string {
	if (!name) return fallback;
	const parts = name.trim().split(/\s+/).filter(Boolean);
	if (parts.length === 0) return fallback;
	if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
	return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

/**
 * Deterministically pick an avatar color class based on the hash of the given name.
 */
export function getAvatarColor(name?: string | null, palette = DEFAULT_AVATAR_COLORS): string {
	if (!name) return palette[0];
	let hash = 0;
	for (let i = 0; i < name.length; i++) {
		hash = name.charCodeAt(i) + ((hash << 5) - hash);
	}
	return palette[Math.abs(hash) % palette.length];
}

/**
 * Converts a string to URL-safe kebab-case slug.
 * Example: "My Test Org!" -> "my-test-org"
 */
export function slugify(str?: string | null): string {
	if (!str) return '';
	return (
		str
			.toLowerCase()
			.trim()
			.replace(/[^a-z0-9]+/g, '-')
			.replace(/^-+|-+$/g, '') || ''
	);
}
