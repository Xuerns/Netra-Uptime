<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import { env } from '$env/dynamic/public';
	import type { PageData } from './$types';
	import type { PublicStatusPage } from '$lib/api/statuspage';
	import { CheckCircle2, AlertCircle, AlertTriangle, XCircle } from 'lucide-svelte';
	import {
		PublicStatusBanner,
		PublicHeader,
		PublicActiveIncidents,
		PublicMaintenanceList,
		PublicComponentGroups,
		PublicIncidentArchive,
		PublicFooter
	} from '$lib/components/statuspage/public';
	import {
		formatDateTime,
		formatTimeAgo,
		formatDateHeader,
		formatDuration
	} from '$lib/utils/datetime';

	let { data }: { data: PageData } = $props();

	// Live reactive status page state
	let liveStatusPage = $state<PublicStatusPage | null>(null);
	let statusPage = $derived(liveStatusPage ?? (data.statusPage as PublicStatusPage));

	// Auto-refresh states
	let lastUpdated = $state(new Date());
	let secondsUntilRefresh = $state(60);
	let isRefreshing = $state(false);
	let refreshTimer: ReturnType<typeof setInterval> | null = null;

	// Component group collapse states (groupId -> boolean)
	let collapsedGroups = $state<Record<string, boolean>>({});

	// Incident history view filter: '7d' | '14d' | 'all'
	let historyRange = $state<'7d' | '14d' | 'all'>('7d');

	let overallStatus = $derived(statusPage?.overall_status || 'operational');
	let themeColor = $derived(statusPage?.theme_color || '#10b981');

	const statusConfig = $derived.by(() => {
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
	});

	function computeMonitorSetStatus(mons: PublicStatusPage['monitors']): string {
		if (!mons || mons.length === 0) return 'operational';
		if (mons.some((m) => m.status === 'down')) return 'major_outage';
		if (mons.some((m) => m.status === 'degraded')) return 'degraded';
		if (mons.every((m) => m.status === 'paused')) return 'paused';
		return 'operational';
	}

	// Group monitors by ComponentGroup
	let monitorGrouping = $derived.by(() => {
		const monitors = statusPage.monitors || [];
		const groups = statusPage.component_groups || [];

		if (groups.length === 0) {
			return {
				hasGroups: false,
				groups: [],
				ungrouped: monitors,
				ungroupedStatus: computeMonitorSetStatus(monitors)
			};
		}

		const grouped = groups.map((g) => {
			const mons = monitors.filter((m) => m.group_id === g.id);
			return {
				group: g,
				monitors: mons,
				status: computeMonitorSetStatus(mons)
			};
		});

		const groupIds = new Set(groups.map((g) => g.id));
		const ungrouped = monitors.filter((m) => !m.group_id || !groupIds.has(m.group_id));

		return {
			hasGroups: true,
			groups: grouped,
			ungrouped,
			ungroupedStatus: computeMonitorSetStatus(ungrouped)
		};
	});

	// Filtered incident history (incidents-only timeline) by selected date range
	let displayedIncidentHistory = $derived.by(() => {
		const history = (statusPage.incident_history || []).filter(
			(day) => day.incidents && day.incidents.length > 0
		);
		if (historyRange === 'all') {
			return history;
		}

		const days = historyRange === '7d' ? 7 : 14;
		const cutoff = new Date();
		cutoff.setHours(0, 0, 0, 0);
		cutoff.setDate(cutoff.getDate() - (days - 1));

		return history.filter((day) => {
			const [y, m, d] = day.date.split('-').map(Number);
			if (!y || !m || !d) return true;
			const dayDate = new Date(y, m - 1, d);
			return dayDate >= cutoff;
		});
	});

	onMount(() => {
		refreshTimer = setInterval(() => {
			if (secondsUntilRefresh > 1) {
				secondsUntilRefresh--;
			} else {
				secondsUntilRefresh = 60;
				triggerRefresh();
			}
		}, 1000);
	});

	onDestroy(() => {
		if (refreshTimer) clearInterval(refreshTimer);
	});

	async function triggerRefresh() {
		if (isRefreshing) return;
		isRefreshing = true;
		try {
			const subdomain = $page.params.subdomain;
			const apiUrl = env.PUBLIC_API_URL || 'http://localhost:27346/api/v1';
			const res = await fetch(`${apiUrl}/public/statuspage/${subdomain}`);
			if (res.ok) {
				const json = await res.json();
				if (json.data) {
					liveStatusPage = json.data;
					lastUpdated = new Date();
				}
			}
		} catch (err) {
			if (import.meta.env.DEV) {
				console.error('Error refreshing status page:', err);
			}
		} finally {
			isRefreshing = false;
			secondsUntilRefresh = 60;
		}
	}

	function toggleGroupCollapse(groupId: string, currentlyCollapsed?: boolean) {
		const current = currentlyCollapsed ?? collapsedGroups[groupId] ?? false;
		collapsedGroups[groupId] = !current;
	}

	function getGroupStatusBadge(status: string) {
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

	function getSeverityBadge(severity: string) {
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
</script>

<svelte:head>
	<title>{statusPage.title} | System Status</title>
	{#if statusPage.favicon_url}
		<link rel="icon" href={statusPage.favicon_url} />
	{/if}
	{#if statusPage.custom_css}
		{@html `<style>${statusPage.custom_css.replace(/<\/?style[^>]*>/gi, '').replace(/<[^>]*>/g, '')}</style>`}
	{/if}
</svelte:head>

<div
	class="min-h-screen relative overflow-x-hidden text-foreground bg-background pattern-bg flex flex-col selection:bg-primary/20"
	style="--primary: {themeColor}; --theme-primary: {themeColor};"
>
	<!-- Top Navigation Bar Component -->
	<PublicHeader
		{statusPage}
		{themeColor}
		{statusConfig}
		{lastUpdated}
		{secondsUntilRefresh}
		{isRefreshing}
		onRefresh={triggerRefresh}
	/>

	<!-- Main Content Container -->
	<main class="flex-1 pb-16 px-4 sm:px-6 max-w-5xl mx-auto w-full space-y-10 mt-6">
		<!-- 1. Overall Status Banner Component -->
		<PublicStatusBanner {statusConfig} {lastUpdated} />

		<!-- 2. Maintenance Schedules (Ongoing & Scheduled) Component -->
		<PublicMaintenanceList
			ongoingMaintenances={statusPage.ongoing_maintenances || []}
			upcomingMaintenances={statusPage.upcoming_maintenances || []}
		/>

		<!-- 3. Active Incidents Section Component -->
		<PublicActiveIncidents
			incidents={statusPage.active_incidents || []}
			{getSeverityBadge}
			timeAgo={formatTimeAgo}
			{formatDateTime}
		/>

		<!-- 4. Services & Systems Section Component -->
		<PublicComponentGroups
			{statusPage}
			{themeColor}
			{monitorGrouping}
			{collapsedGroups}
			onToggleGroup={toggleGroupCollapse}
			{getGroupStatusBadge}
		/>

		<!-- 5. Incident History Archive Component -->
		<PublicIncidentArchive
			{historyRange}
			{displayedIncidentHistory}
			{themeColor}
			onSelectRange={(range) => (historyRange = range)}
			{formatDateHeader}
			{getSeverityBadge}
			{formatDuration}
			{formatDateTime}
		/>
	</main>

	<!-- Footer Component -->
	<PublicFooter title={statusPage.title} {overallStatus} />
</div>

<style>
	.pattern-bg {
		background-color: #f1f5f9;
		background-image: radial-gradient(#cbd5e1 1px, transparent 1px);
		background-size: 24px 24px;
	}
	:global(.dark) .pattern-bg {
		background-color: #0b0c0e;
		background-image: radial-gradient(rgba(255, 255, 255, 0.05) 1px, transparent 1px);
		background-size: 24px 24px;
	}
</style>
