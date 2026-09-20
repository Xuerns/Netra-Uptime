import { Globe, Server, Wifi } from 'lucide-svelte';
import type { Monitor } from '$lib/api/monitor';

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
	const diff = Date.now() - new Date(dateStr).getTime();
	const secs = Math.floor(diff / 1000);
	if (secs < 60) return `${secs}s ago`;
	const mins = Math.floor(secs / 60);
	if (mins < 60) return `${mins}m ago`;
	const hrs = Math.floor(mins / 60);
	return `${hrs}h ago`;
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
