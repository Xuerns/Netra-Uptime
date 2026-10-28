<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import type { PageData } from './$types';
	import { StatusPageAPI, type PublicStatusPage } from '$lib/api/statuspage';
	import {
		PublicStatusBanner,
		PublicHeader,
		PublicActiveIncidents,
		PublicMaintenanceList,
		PublicComponentGroups,
		PublicIncidentArchive,
		PublicFooter,
		getStatusConfig,
		computeMonitorGrouping,
		filterIncidentHistory,
		getSeverityBadge,
		sanitizeCustomCss
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
	let safeCustomCss = $derived(sanitizeCustomCss(statusPage?.custom_css));

	const statusConfig = $derived(getStatusConfig(overallStatus, themeColor));

	// Group monitors by ComponentGroup
	let monitorGrouping = $derived(
		computeMonitorGrouping(statusPage?.monitors, statusPage?.component_groups)
	);

	// Filtered incident history (incidents-only timeline) by selected date range
	let displayedIncidentHistory = $derived(
		filterIncidentHistory(statusPage?.incident_history, historyRange)
	);

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
			if (subdomain) {
				const res = await StatusPageAPI.getPublic(subdomain);
				if (res.data) {
					liveStatusPage = res.data;
					lastUpdated = new Date();
				}
			}
		} catch {
			// Auto-refresh failed silently
		} finally {
			isRefreshing = false;
			secondsUntilRefresh = 60;
		}
	}

	function toggleGroupCollapse(groupId: string, currentlyCollapsed?: boolean) {
		const current = currentlyCollapsed ?? collapsedGroups[groupId] ?? false;
		collapsedGroups[groupId] = !current;
	}
</script>

<svelte:head>
	<title>{statusPage.title} | System Status</title>
	{#if statusPage.favicon_url}
		<link rel="icon" href={statusPage.favicon_url} />
	{/if}
	{#if safeCustomCss}
		<svelte:element this={"style"}>{safeCustomCss}</svelte:element>
	{/if}
</svelte:head>

<div
	class="min-h-screen relative overflow-x-hidden text-foreground bg-background pattern-bg flex flex-col selection:bg-primary/20"
	style="--primary: {themeColor}; --theme-primary: {themeColor};"
>
	<!-- Top Navigation Bar Component -->
	<PublicHeader
		{statusPage}
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
