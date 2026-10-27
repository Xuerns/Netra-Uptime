import { Globe, Server, Wifi } from 'lucide-svelte';
import type { Monitor } from '$lib/api/monitor';
import { formatTimeAgo, formatDateOnly } from './datetime';

export function typeIcon(type: Monitor['Type']) {
	switch (type) {
		case 'http':
			return Globe;
		case 'tcp':
			return Server;
		case 'ping':
			return Wifi;
	}
}

export function statusLabel(s: Monitor['Status']): string {
	switch (s) {
		case 'up':
			return 'UP';
		case 'down':
			return 'DOWN';
		case 'degraded':
			return 'DEGRADED';
		case 'paused':
			return 'PAUSED';
		case 'pending':
			return 'PENDING';
		default:
			return (s as string)?.toUpperCase() ?? 'UNKNOWN';
	}
}

export function formatInterval(s: number): string {
	if (s < 60) return `${s}s`;
	return `${s / 60}m`;
}

export function formatLastChecked(dateStr?: string): string {
	if (!dateStr) return 'Never';
	return formatTimeAgo(dateStr);
}

export function formatResponseTime(ms?: number): string {
	if (!ms && ms !== 0) return '—';
	if (ms >= 1000) return `${(ms / 1000).toFixed(1)}s`;
	return `${ms}ms`;
}

export function uptimeColor(pct?: number): string {
	if (!pct && pct !== 0) return 'text-on-surface-variant';
	if (pct >= 99) return 'text-primary';
	if (pct >= 95) return 'text-yellow-400';
	return 'text-red-400';
}

export function trimTarget(target: string, maxLen = 42): string {
	if (target.length <= maxLen) return target;
	return target.slice(0, maxLen) + '…';
}

export interface CertificateBadgeInfo {
	label: string;
	badgeClass: string;
	tooltip: string;
	isAvailable: boolean;
}

export function formatSSLBadge(
	status?: string,
	daysLeft?: number | null,
	expiresAt?: string | null,
	errorMsg?: string
): CertificateBadgeInfo {
	if (status === 'valid' && daysLeft != null) {
		return {
			label: `SSL: ${daysLeft}d`,
			badgeClass: 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400',
			tooltip: expiresAt ? `SSL Expires: ${formatDateOnly(expiresAt)}` : 'SSL Certificate is valid',
			isAvailable: true
		};
	}

	if (status === 'warning' && daysLeft != null) {
		return {
			label: `SSL: ${daysLeft}d left`,
			badgeClass: 'bg-amber-500/10 border-amber-500/20 text-amber-400',
			tooltip: expiresAt
				? `SSL Expiring Soon: ${formatDateOnly(expiresAt)}`
				: 'SSL Certificate expiring soon',
			isAvailable: true
		};
	}

	if (status === 'expired' || (daysLeft != null && daysLeft <= 0)) {
		return {
			label: 'SSL: Expired',
			badgeClass: 'bg-rose-500/10 border-rose-500/20 text-rose-400',
			tooltip: errorMsg || 'SSL Certificate has expired',
			isAvailable: true
		};
	}

	return {
		label: 'SSL: Unavailable',
		badgeClass: 'bg-white/5 border-border text-muted-foreground',
		tooltip: errorMsg || 'SSL details unavailable or non-TLS endpoint',
		isAvailable: false
	};
}

export function formatDomainBadge(
	status?: string,
	daysLeft?: number | null,
	expiresAt?: string | null,
	errorMsg?: string
): CertificateBadgeInfo {
	if (status === 'valid' && daysLeft != null) {
		return {
			label: `Domain: ${daysLeft}d`,
			badgeClass: 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400',
			tooltip: expiresAt
				? `Domain Expires: ${formatDateOnly(expiresAt)}`
				: 'Domain registration is valid',
			isAvailable: true
		};
	}

	if (status === 'warning' && daysLeft != null) {
		return {
			label: `Domain: ${daysLeft}d left`,
			badgeClass: 'bg-amber-500/10 border-amber-500/20 text-amber-400',
			tooltip: expiresAt
				? `Domain Expiring Soon: ${formatDateOnly(expiresAt)}`
				: 'Domain registration expiring soon',
			isAvailable: true
		};
	}

	if (status === 'expired' || (daysLeft != null && daysLeft <= 0)) {
		return {
			label: 'Domain: Expired',
			badgeClass: 'bg-rose-500/10 border-rose-500/20 text-rose-400',
			tooltip: errorMsg || 'Domain registration has expired',
			isAvailable: true
		};
	}

	return {
		label: 'Domain: Unavailable',
		badgeClass: 'bg-white/5 border-border text-muted-foreground',
		tooltip: errorMsg || 'WHOIS registry lookup unavailable for this domain or subdomain',
		isAvailable: false
	};
}
