/**
 * Utilitas format tanggal/waktu yang timezone-aware.
 * Semua fungsi menggunakan Intl.DateTimeFormat (native) — zero dependency.
 *
 * Gunakan fungsi ini di seluruh app, bukan toLocaleDateString() inline.
 */

// ── Core formatter ────────────────────────────────────────────────────────────

/**
 * Format tanggal + waktu lengkap.
 * Contoh output: "Aug 20, 2026, 14:35"
 */
export function formatDateTime(dateStr: string | undefined | null, timezone = 'UTC'): string {
	if (!dateStr) return '—';
	try {
		return new Intl.DateTimeFormat('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
			timeZone: timezone
		}).format(new Date(dateStr));
	} catch {
		return dateStr;
	}
}

/**
 * Format tanggal saja (tanpa jam).
 * Contoh output: "Aug 20, 2026"
 */
export function formatDateOnly(dateStr: string | undefined | null, timezone = 'UTC'): string {
	if (!dateStr) return '—';
	try {
		return new Intl.DateTimeFormat('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			timeZone: timezone
		}).format(new Date(dateStr));
	} catch {
		return dateStr;
	}
}

/**
 * Format waktu saja dengan label timezone singkat.
 * Contoh output: "14:35 WIB"
 */
export function formatTimeOnly(dateStr: string | undefined | null, timezone = 'UTC'): string {
	if (!dateStr) return '—';
	try {
		const time = new Intl.DateTimeFormat('en-US', {
			hour: '2-digit',
			minute: '2-digit',
			timeZone: timezone
		}).format(new Date(dateStr));

		// Ambil abbreviasi timezone (misal "WIB", "EST", "UTC")
		const tzAbbr =
			new Intl.DateTimeFormat('en-US', {
				timeZoneName: 'short',
				timeZone: timezone
			})
				.formatToParts(new Date(dateStr))
				.find((p) => p.type === 'timeZoneName')?.value ?? timezone;

		return `${time} ${tzAbbr}`;
	} catch {
		return dateStr;
	}
}

/**
 * Format waktu relatif (tidak perlu timezone — selalu relatif ke "sekarang").
 * Contoh output: "just now", "5m ago", "3h ago", "2d ago"
 */
export function formatTimeAgo(dateStr: string | undefined | null): string {
	if (!dateStr) return '—';
	try {
		const date = new Date(dateStr);
		const now = new Date();
		const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

		if (seconds < 5) return 'just now';
		if (seconds < 60) return `${seconds}s ago`;
		const minutes = Math.floor(seconds / 60);
		if (minutes < 60) return `${minutes}m ago`;
		const hours = Math.floor(minutes / 60);
		if (hours < 24) return `${hours}h ago`;
		const days = Math.floor(hours / 24);
		if (days < 30) return `${days}d ago`;
		const months = Math.floor(days / 30);
		if (months < 12) return `${months}mo ago`;
		return `${Math.floor(months / 12)}y ago`;
	} catch {
		return dateStr;
	}
}

/**
 * Format "last checked" — alias formatTimeAgo dengan fallback 'Never'.
 * Digunakan di monitor cards & list rows.
 */
export function formatLastChecked(dateStr: string | undefined | null): string {
	if (!dateStr) return 'Never';
	return formatTimeAgo(dateStr);
}

/**
 * Format incident time dengan pola militer: "T-HH:MM:00"
 * Contoh output: "T-02:15:00"
 */
export function formatIncidentTimeAgo(dateStr: string | undefined | null): string {
	if (!dateStr) return 'T-??:??:00';
	try {
		const diff = Date.now() - new Date(dateStr).getTime();
		const hours = Math.floor(diff / (1000 * 60 * 60));
		const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
		return `T-${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:00`;
	} catch {
		return 'T-??:??:00';
	}
}
