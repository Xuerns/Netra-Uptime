import type {
	PublicMonitor,
	PublicComponentGroup,
	PublicIncidentDayGroup
} from '$lib/api/statuspage';
import { CheckCircle2, AlertCircle, AlertTriangle, XCircle } from 'lucide-svelte';

export interface GroupedMonitorSection {
	hasGroups: boolean;
	groups: Array<{
		group: PublicComponentGroup;
		monitors: PublicMonitor[];
		status: string;
	}>;
	ungrouped: PublicMonitor[];
	ungroupedStatus: string;
}

export interface StatusBannerConfig {
	title: string;
	description: string;
	bannerClass: string;
	bannerStyle: string;
	dotClass: string;
	dotStyle: string;
	icon: typeof CheckCircle2;
	accentColor: string;
	isThemed: boolean;
}

export function computeMonitorSetStatus(mons: PublicMonitor[] | undefined): string {
	if (!mons || mons.length === 0) return 'operational';
	if (mons.some((m) => m.status === 'down')) return 'major_outage';
	if (mons.some((m) => m.status === 'degraded')) return 'degraded';
	if (mons.every((m) => m.status === 'paused')) return 'paused';
	return 'operational';
}

export function computeMonitorGrouping(
	monitors: PublicMonitor[] | undefined,
	groups: PublicComponentGroup[] | undefined
): GroupedMonitorSection {
	const mons = monitors || [];
	const grps = groups || [];

	if (grps.length === 0) {
		return {
			hasGroups: false,
			groups: [],
			ungrouped: mons,
			ungroupedStatus: computeMonitorSetStatus(mons)
		};
	}

	const grouped = grps.map((g) => {
		const mList = mons.filter((m) => m.group_id === g.id);
		return {
			group: g,
			monitors: mList,
			status: computeMonitorSetStatus(mList)
		};
	});

	const groupIds = new Set(grps.map((g) => g.id));
	const ungrouped = mons.filter((m) => !m.group_id || !groupIds.has(m.group_id));

	return {
		hasGroups: true,
		groups: grouped,
		ungrouped,
		ungroupedStatus: computeMonitorSetStatus(ungrouped)
	};
}

export function getStatusConfig(overallStatus: string, themeColor: string): StatusBannerConfig {
	switch (overallStatus) {
		case 'major_outage':
			return {
				title: 'Major System Outage',
				description: 'We are currently experiencing severe disruption across our services.',
				bannerClass:
					'bg-red-500/10 border-red-500/30 text-red-400 shadow-[0_0_30px_rgba(239,68,68,0.1)]',
				bannerStyle: '',
				dotClass: 'bg-red-500 shadow-[0_0_14px_rgba(239,68,68,0.8)]',
				dotStyle: '',
				icon: XCircle,
				accentColor: '#ef4444',
				isThemed: false
			};
		case 'partial_outage':
			return {
				title: 'Partial System Outage',
				description: 'Some of our services are currently experiencing disruptions.',
				bannerClass:
					'bg-orange-500/10 border-orange-500/30 text-orange-400 shadow-[0_0_30px_rgba(249,115,22,0.1)]',
				bannerStyle: '',
				dotClass: 'bg-orange-500 shadow-[0_0_14px_rgba(249,115,22,0.8)]',
				dotStyle: '',
				icon: AlertTriangle,
				accentColor: '#f97316',
				isThemed: false
			};
		case 'degraded':
			return {
				title: 'Degraded Performance',
				description: 'We are experiencing elevated latency or minor service issues.',
				bannerClass:
					'bg-amber-500/10 border-amber-500/30 text-amber-400 shadow-[0_0_30px_rgba(245,158,11,0.1)]',
				bannerStyle: '',
				dotClass: 'bg-amber-500 shadow-[0_0_14px_rgba(245,158,11,0.8)]',
				dotStyle: '',
				icon: AlertCircle,
				accentColor: '#f59e0b',
				isThemed: false
			};
		default:
			return {
				title: 'All Systems Operational',
				description: 'All services are online and operating within optimal parameters.',
				bannerClass: '',
				bannerStyle: `background-color: color-mix(in srgb, ${themeColor} 12%, transparent); border-color: color-mix(in srgb, ${themeColor} 30%, transparent); color: ${themeColor}; box-shadow: 0 0 30px color-mix(in srgb, ${themeColor} 15%, transparent);`,
				dotClass: '',
				dotStyle: `background-color: ${themeColor}; box-shadow: 0 0 14px ${themeColor};`,
				icon: CheckCircle2,
				accentColor: themeColor,
				isThemed: true
			};
	}
}

export function filterIncidentHistory(
	incidentHistory: PublicIncidentDayGroup[] | undefined,
	range: '7d' | '14d' | 'all'
): PublicIncidentDayGroup[] {
	const history = (incidentHistory || []).filter(
		(day) => day.incidents && day.incidents.length > 0
	);
	if (range === 'all') {
		return history;
	}

	const days = range === '7d' ? 7 : 14;
	const cutoff = new Date();
	cutoff.setHours(0, 0, 0, 0);
	cutoff.setDate(cutoff.getDate() - (days - 1));

	return history.filter((day) => {
		const [y, m, d] = day.date.split('-').map(Number);
		if (!y || !m || !d) return true;
		const dayDate = new Date(y, m - 1, d);
		return dayDate >= cutoff;
	});
}

export function getGroupStatusBadge(status: string, themeColor: string) {
	switch (status) {
		case 'major_outage':
			return {
				label: 'Disrupted',
				dotClass: 'bg-red-500',
				dotStyle: '',
				textClass: 'text-red-400',
				textStyle: ''
			};
		case 'degraded':
			return {
				label: 'Degraded',
				dotClass: 'bg-amber-500',
				dotStyle: '',
				textClass: 'text-amber-400',
				textStyle: ''
			};
		case 'paused':
			return {
				label: 'Paused',
				dotClass: 'bg-blue-400',
				dotStyle: '',
				textClass: 'text-blue-400',
				textStyle: ''
			};
		default:
			return {
				label: 'Operational',
				dotClass: '',
				dotStyle: `background-color: ${themeColor};`,
				textClass: '',
				textStyle: `color: ${themeColor};`
			};
	}
}

export function getSeverityBadge(severity: string) {
	switch (severity?.toLowerCase()) {
		case 'critical':
			return 'bg-red-500/20 text-red-300 border-red-500/30';
		case 'high':
			return 'bg-orange-500/20 text-orange-300 border-orange-500/30';
		case 'medium':
			return 'bg-amber-500/20 text-amber-300 border-amber-500/30';
		case 'low':
		default:
			return 'bg-blue-500/20 text-blue-300 border-blue-500/30';
	}
}

export function sanitizeCustomCss(css?: string | null): string {
	if (!css) return '';
	return css
		.replace(/<\s*\/\s*style[^>]*>/gi, '')
		.replace(/<\s*style[^>]*>/gi, '')
		.replace(/<\s*script[^>]*>[\s\S]*?<\s*\/\s*script[^>]*>/gi, '')
		.replace(/<!--[\s\S]*?-->/g, '')
		.replace(/<[^>]+>/g, '');
}
