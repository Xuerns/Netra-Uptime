/**
 * Timezone-aware date and time formatting utilities.
 * All functions use native Intl.DateTimeFormat (zero-dependency).
 */

export type DateInput = string | number | Date | undefined | null;

function toDate(input: DateInput): Date | null {
	if (!input) return null;
	const d = input instanceof Date ? input : new Date(input);
	return isNaN(d.getTime()) ? null : d;
}

// ── Core formatters ──────────────────────────────────────────────────────────

/**
 * Format full date and time.
 * Example output: "Aug 20, 2026, 14:35"
 */
export function formatDateTime(dateStr: DateInput, timezone = 'UTC'): string {
	const d = toDate(dateStr);
	if (!d) return '—';
	try {
		return new Intl.DateTimeFormat('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
			timeZone: timezone
		}).format(d);
	} catch {
		return String(dateStr);
	}
}

/**
 * Format short date and time without year.
 * Example output: "Aug 20, 14:35"
 */
export function formatDateTimeShort(dateStr: DateInput, timezone = 'UTC'): string {
	const d = toDate(dateStr);
	if (!d) return '—';
	try {
		return new Intl.DateTimeFormat('en-US', {
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
			timeZone: timezone
		}).format(d);
	} catch {
		return String(dateStr);
	}
}

/**
 * Format date only (no time).
 * Example output: "Aug 20, 2026"
 */
export function formatDateOnly(dateStr: DateInput, timezone = 'UTC'): string {
	const d = toDate(dateStr);
	if (!d) return '—';
	try {
		return new Intl.DateTimeFormat('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			timeZone: timezone
		}).format(d);
	} catch {
		return String(dateStr);
	}
}

/**
 * Format month and day only.
 * Example output: "Aug 20"
 */
export function formatDateMonthDay(dateStr: DateInput, timezone = 'UTC'): string {
	const d = toDate(dateStr);
	if (!d) return '—';
	try {
		return new Intl.DateTimeFormat('en-US', {
			month: 'short',
			day: 'numeric',
			timeZone: timezone
		}).format(d);
	} catch {
		return String(dateStr);
	}
}

/**
 * Format time only with timezone abbreviation.
 * Example output: "14:35 WIB"
 */
export function formatTimeOnly(dateStr: DateInput, timezone = 'UTC'): string {
	const d = toDate(dateStr);
	if (!d) return '—';
	try {
		const time = new Intl.DateTimeFormat('en-US', {
			hour: '2-digit',
			minute: '2-digit',
			timeZone: timezone
		}).format(d);

		const tzAbbr =
			new Intl.DateTimeFormat('en-US', {
				timeZoneName: 'short',
				timeZone: timezone
			})
				.formatToParts(d)
				.find((p) => p.type === 'timeZoneName')?.value ?? timezone;

		return `${time} ${tzAbbr}`;
	} catch {
		return String(dateStr);
	}
}

/**
 * Format relative time (relative to current time).
 * Example output: "just now", "5m ago", "3h ago", "2d ago"
 */
export function formatTimeAgo(dateStr: DateInput): string {
	const date = toDate(dateStr);
	if (!date) return '—';
	try {
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
		return String(dateStr);
	}
}

/**
 * Format incident relative time: "just now", "5m ago", "11h 1m ago", "2d 4h ago"
 */
export function formatIncidentTimeAgo(dateStr: DateInput): string {
	const d = toDate(dateStr);
	if (!d) return '—';
	try {
		const parsed = d.getTime();
		const diff = Math.max(0, Date.now() - parsed);
		const seconds = Math.floor(diff / 1000);
		if (seconds < 5) return 'just now';
		if (seconds < 60) return `${seconds}s ago`;
		const minutes = Math.floor(seconds / 60);
		if (minutes < 60) return `${minutes}m ago`;
		const hours = Math.floor(minutes / 60);
		const remMinutes = minutes % 60;
		if (hours < 24) {
			return remMinutes > 0 ? `${hours}h ${remMinutes}m ago` : `${hours}h ago`;
		}
		const days = Math.floor(hours / 24);
		const remHours = hours % 24;
		if (days < 30) {
			return remHours > 0 ? `${days}d ${remHours}h ago` : `${days}d ago`;
		}
		const months = Math.floor(days / 30);
		if (months < 12) return `${months}mo ago`;
		return `${Math.floor(months / 12)}y ago`;
	} catch {
		return String(dateStr);
	}
}

/**
 * Format archive date header (e.g. "August 20, 2026") from a "YYYY-MM-DD" string or date.
 */
export function formatDateHeader(dateStr?: string | null): string {
	if (!dateStr) return '';
	const parts = dateStr.split('-');
	if (parts.length === 3) {
		const d = new Date(Number(parts[0]), Number(parts[1]) - 1, Number(parts[2]));
		if (!isNaN(d.getTime())) {
			return new Intl.DateTimeFormat('en-US', {
				month: 'long',
				day: 'numeric',
				year: 'numeric'
			}).format(d);
		}
	}
	const d = toDate(dateStr);
	if (d) {
		return new Intl.DateTimeFormat('en-US', {
			month: 'long',
			day: 'numeric',
			year: 'numeric'
		}).format(d);
	}
	return dateStr;
}

/**
 * Format duration between start and resolved time of an incident (e.g., "45m", "2h 15m").
 * If resolvedAt is missing (ongoing), calculates duration from start to now.
 */
export function formatDuration(startedAt?: DateInput, resolvedAt?: DateInput): string {
	const start = toDate(startedAt);
	if (!start) return '—';
	try {
		const end = toDate(resolvedAt) ?? new Date();
		const diffMins = Math.max(1, Math.round((end.getTime() - start.getTime()) / 60000));
		if (diffMins < 60) return `${diffMins}m`;
		const hours = Math.floor(diffMins / 60);
		const remainingMins = diffMins % 60;
		if (remainingMins === 0) return `${hours}h`;
		return `${hours}h ${remainingMins}m`;
	} catch {
		return '—';
	}
}
