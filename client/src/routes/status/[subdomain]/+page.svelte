<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import { env } from '$env/dynamic/public';
	import type { PageData } from './$types';
	import type {
		PublicStatusPage,
		PublicMonitor,
		PublicIncident,
		PublicMaintenance,
		PublicComponentGroup,
		PublicIncidentDayGroup
	} from '$lib/api/statuspage';
	import {
		CheckCircle2,
		AlertCircle,
		AlertTriangle,
		XCircle,
		Clock,
		ShieldCheck,
		Activity,
		Radio,
		Wrench,
		Calendar,
		ChevronDown,
		RefreshCw,
		FolderTree,
		Archive,
		Check
	} from 'lucide-svelte';
	import logo from '$lib/assets/Logo.svg';
	import Sparkline from '$lib/components/Sparkline.svelte';
	import UptimeBar from '$lib/components/statuspage/UptimeBar.svelte';
	import MaintenanceCard from '$lib/components/statuspage/MaintenanceCard.svelte';

	let { data } = $props<{ data: PageData }>();

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

	// Group monitors by ComponentGroup
	let monitorGrouping = $derived.by(() => {
		const monitors = statusPage.monitors || [];
		const groups = statusPage.component_groups || [];

		if (groups.length === 0) {
			return {
				hasGroups: false,
				groups: [],
				ungrouped: monitors
			};
		}

		const grouped = groups.map((g) => {
			const mons = monitors.filter((m) => m.group_id === g.id);
			let groupStatus = 'operational';
			if (mons.some((m) => m.status === 'down')) {
				groupStatus = 'major_outage';
			} else if (mons.some((m) => m.status === 'degraded')) {
				groupStatus = 'degraded';
			} else if (mons.length > 0 && mons.every((m) => m.status === 'paused')) {
				groupStatus = 'paused';
			}

			return {
				group: g,
				monitors: mons,
				status: groupStatus
			};
		});

		const groupIds = new Set(groups.map((g) => g.id));
		const ungrouped = monitors.filter((m) => !m.group_id || !groupIds.has(m.group_id));

		return {
			hasGroups: true,
			groups: grouped,
			ungrouped
		};
	});

	// Filtered incident history by selected date range
	let displayedIncidentHistory = $derived.by(() => {
		const history = statusPage.incident_history || [];
		if (historyRange === '7d') {
			return history.slice(0, 7);
		} else if (historyRange === '14d') {
			return history.slice(0, 14);
		}
		return history;
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
			const apiUrl = env.PUBLIC_API_URL || 'http://localhost:8080/api/v1';
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

	function toggleGroupCollapse(groupId: string) {
		collapsedGroups[groupId] = !collapsedGroups[groupId];
	}

	function getMonitorStatusInfo(status: string) {
		switch (status) {
			case 'up':
				return {
					label: 'Operational',
					dotClass: '',
					dotStyle: `background-color: ${themeColor}; box-shadow: 0 0 8px ${themeColor}b3;`,
					textClass: '',
					textStyle: `color: ${themeColor};`,
					sparklineColor: themeColor
				};
			case 'degraded':
				return {
					label: 'Degraded',
					dotClass: 'bg-amber-500 shadow-[0_0_8px_rgba(245,158,11,0.7)]',
					dotStyle: '',
					textClass: 'text-amber-400',
					textStyle: '',
					sparklineColor: '#f59e0b'
				};
			case 'paused':
				return {
					label: 'Paused',
					dotClass: 'bg-blue-400 shadow-[0_0_8px_rgba(96,165,250,0.7)]',
					dotStyle: '',
					textClass: 'text-blue-400',
					textStyle: '',
					sparklineColor: '#60a5fa'
				};
			default:
				return {
					label: 'Down',
					dotClass: 'bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.7)]',
					dotStyle: '',
					textClass: 'text-red-400',
					textStyle: '',
					sparklineColor: '#ef4444'
				};
		}
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

	function timeAgo(dateString?: string): string {
		if (!dateString) return '';
		const date = new Date(dateString);
		const now = new Date();
		const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);
		if (seconds < 60) return 'just now';
		const minutes = Math.floor(seconds / 60);
		if (minutes < 60) return `${minutes}m ago`;
		const hours = Math.floor(minutes / 60);
		if (hours < 24) return `${hours}h ago`;
		const days = Math.floor(hours / 24);
		if (days === 1) return 'yesterday';
		return `${days}d ago`;
	}

	function formatDateTime(dateString?: string): string {
		if (!dateString) return '';
		const date = new Date(dateString);
		return date.toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
			hour12: false
		});
	}

	function formatDateHeader(dateStr: string): string {
		if (!dateStr) return '';
		const parts = dateStr.split('-');
		if (parts.length !== 3) return dateStr;
		const d = new Date(Number(parts[0]), Number(parts[1]) - 1, Number(parts[2]));
		return d.toLocaleDateString('en-US', {
			month: 'long',
			day: 'numeric',
			year: 'numeric'
		});
	}

	function formatDuration(startedAt?: string, resolvedAt?: string | null): string {
		if (!startedAt || !resolvedAt) return '';
		const start = new Date(startedAt).getTime();
		const end = new Date(resolvedAt).getTime();
		const diffMins = Math.max(1, Math.round((end - start) / 60000));
		if (diffMins < 60) return `${diffMins}m`;
		const hours = Math.floor(diffMins / 60);
		const remainingMins = diffMins % 60;
		if (remainingMins === 0) return `${hours}h`;
		return `${hours}h ${remainingMins}m`;
	}
</script>

<svelte:head>
	<title>{statusPage.title} | System Status</title>
	{#if statusPage.favicon_url}
		<link rel="icon" href={statusPage.favicon_url} />
	{/if}
	{#if statusPage.custom_css}
		{@html `<style>${statusPage.custom_css}</style>`}
	{/if}
</svelte:head>

<div
	class="min-h-screen relative overflow-x-hidden text-foreground bg-background pattern-bg flex flex-col selection:bg-primary/20"
	style="--primary: {themeColor}; --theme-primary: {themeColor};"
>
	<!-- Top Navigation Bar -->
	<header
		class="sticky top-0 w-full z-50 bg-background/80 backdrop-blur-md border-b border-border/60 transition-colors"
	>
		<div class="max-w-5xl mx-auto flex justify-between items-center px-4 sm:px-6 py-4 w-full">
			<div class="flex items-center gap-3">
				{#if statusPage.logo_url}
					<img
						class="h-8 max-w-36 object-contain rounded"
						src={statusPage.logo_url}
						alt={`${statusPage.title} Logo`}
					/>
				{:else}
					<img class="h-8 w-8 object-contain" src={logo} alt="PulseWatch Logo" />
				{/if}
				<div class="flex flex-col">
					<div class="flex items-center gap-2">
						<span class="font-headline-md text-headline-md font-bold text-foreground">
							{statusPage.title}
						</span>
						<span
							style="border-color: color-mix(in srgb, {themeColor} 35%, transparent); color: {themeColor};"
							class="text-[11px] font-mono px-2 py-0.5 rounded-full bg-surface-dim border"
						>
							STATUS
						</span>
					</div>
					{#if statusPage.description}
						<span class="text-xs sm:text-sm text-muted-foreground font-body-md mt-0.5 line-clamp-1">
							{statusPage.description}
						</span>
					{/if}
				</div>
			</div>

			<!-- Live & Refresh Controls -->
			<div class="flex items-center gap-2.5">
				<button
					type="button"
					class="hidden sm:inline-flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground bg-white/5 hover:bg-white/10 border border-border/80 px-2.5 py-1 rounded-full transition-colors cursor-pointer"
					title="Click to refresh status immediately"
					onclick={triggerRefresh}
				>
					<RefreshCw
						class="size-3 text-muted-foreground {isRefreshing ? 'animate-spin text-primary' : ''}"
					/>
					<span
						>{isRefreshing ? 'Updating...' : `Updated ${timeAgo(lastUpdated.toISOString())}`}</span
					>
				</button>

				<span
					class="inline-flex items-center gap-1.5 text-xs text-muted-foreground bg-white/5 border border-border/80 px-2.5 py-1 rounded-full"
					title={`Auto-refresh in ${secondsUntilRefresh}s`}
				>
					<span
						style={statusConfig.isThemed ? statusConfig.dotStyle : ''}
						class="size-2 rounded-full {statusConfig.dotClass} animate-pulse"
					></span>
					<span class="font-mono">{secondsUntilRefresh}s</span>
				</span>
			</div>
		</div>
	</header>

	<!-- Main Content Container -->
	<main class="flex-1 pb-16 px-4 sm:px-6 max-w-5xl mx-auto w-full space-y-10 mt-6">
		<!-- 1. Overall Status Banner -->
		<section
			style={statusConfig.isThemed ? statusConfig.bannerStyle : ''}
			class="rounded-xl border p-6 sm:p-7 backdrop-blur-md transition-all duration-300 {statusConfig.bannerClass}"
		>
			<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
				<div class="flex items-center gap-4">
					<div class="relative flex items-center justify-center">
						<span
							style={statusConfig.isThemed ? statusConfig.dotStyle : ''}
							class="size-3.5 rounded-full {statusConfig.dotClass} animate-ping absolute opacity-75"
						></span>
						<span
							style={statusConfig.isThemed ? statusConfig.dotStyle : ''}
							class="size-3.5 rounded-full {statusConfig.dotClass} relative"
						></span>
					</div>
					<div>
						<h1
							class="text-xl sm:text-2xl font-bold tracking-tight text-foreground flex items-center gap-2.5"
						>
							{statusConfig.title}
						</h1>
						<p class="text-xs sm:text-sm text-muted-foreground mt-1">
							{statusConfig.description}
						</p>
					</div>
				</div>
				<div
					class="flex items-center gap-1.5 text-xs text-muted-foreground self-start sm:self-auto shrink-0 bg-black/20 sm:bg-transparent px-3 py-1.5 sm:p-0 rounded-lg"
				>
					<Clock class="size-3.5 text-muted-foreground" />
					<span
						>Last verified: {lastUpdated.toLocaleTimeString('en-US', {
							hour: '2-digit',
							minute: '2-digit',
							second: '2-digit',
							hour12: false
						})}</span
					>
				</div>
			</div>
		</section>

		<!-- 2. Ongoing Maintenance Section (if active) -->
		{#if statusPage.ongoing_maintenances && statusPage.ongoing_maintenances.length > 0}
			<section class="space-y-4">
				<div class="flex items-center gap-2">
					<Wrench class="size-4 text-blue-400" />
					<h2 class="text-lg font-semibold tracking-tight text-foreground">Ongoing Maintenance</h2>
				</div>
				<div class="space-y-3">
					{#each statusPage.ongoing_maintenances as maintenance}
						<MaintenanceCard {maintenance} isOngoing={true} />
					{/each}
				</div>
			</section>
		{/if}

		<!-- 3. Upcoming Maintenance Section (if scheduled) -->
		{#if statusPage.upcoming_maintenances && statusPage.upcoming_maintenances.length > 0}
			<section class="space-y-4">
				<div class="flex items-center gap-2">
					<Calendar class="size-4 text-indigo-400" />
					<h2 class="text-lg font-semibold tracking-tight text-foreground">
						Scheduled Maintenance
					</h2>
				</div>
				<div class="space-y-3">
					{#each statusPage.upcoming_maintenances as maintenance}
						<MaintenanceCard {maintenance} isOngoing={false} />
					{/each}
				</div>
			</section>
		{/if}

		<!-- 4. Active Incidents Section -->
		{#if statusPage.active_incidents && statusPage.active_incidents.length > 0}
			<section class="space-y-4">
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-2">
						<AlertTriangle class="size-4 text-amber-400" />
						<h2 class="text-lg font-semibold tracking-tight text-foreground">Active Incidents</h2>
						<span
							class="px-2 py-0.5 rounded-full text-xs font-mono font-medium bg-amber-500/20 text-amber-300 border border-amber-500/30"
						>
							{statusPage.active_incidents.length}
						</span>
					</div>
				</div>

				<div class="space-y-4">
					{#each statusPage.active_incidents as incident}
						<div
							class="rounded-xl border border-border/80 bg-white/3 backdrop-blur-md overflow-hidden transition-all duration-200 hover:border-border"
						>
							<!-- Incident Header -->
							<div class="p-5 sm:p-6 border-b border-border/40 space-y-3">
								<div class="flex flex-wrap items-center justify-between gap-2">
									<div class="flex items-center gap-2">
										<span
											class="text-[11px] font-mono uppercase tracking-wider px-2 py-0.5 rounded border {getSeverityBadge(
												incident.severity
											)}"
										>
											{incident.severity}
										</span>
										<span
											class="text-[11px] font-mono uppercase tracking-wider px-2 py-0.5 rounded bg-amber-500/10 text-amber-400 border border-amber-500/20"
										>
											Ongoing
										</span>
										{#if incident.monitor_name}
											<span class="text-xs text-muted-foreground flex items-center gap-1">
												• <span class="text-foreground/90 font-medium">{incident.monitor_name}</span
												>
											</span>
										{/if}
									</div>
									<span class="text-xs text-muted-foreground font-mono flex items-center gap-1">
										<Clock class="size-3" />
										Started {timeAgo(incident.started_at)}
									</span>
								</div>

								<h3 class="text-base sm:text-lg font-semibold text-foreground">
									{incident.summary}
								</h3>
							</div>

							<!-- Incident Timeline Updates -->
							{#if incident.updates && incident.updates.length > 0}
								<div class="p-5 sm:p-6 bg-black/10">
									<div
										class="space-y-4 relative before:absolute before:left-2 before:top-2 before:bottom-2 before:w-[1px] before:bg-border/60"
									>
										{#each incident.updates as update}
											<div class="relative pl-6 space-y-1">
												<span
													class="absolute left-1.5 top-1.5 size-1.5 rounded-full bg-amber-400 -translate-x-1/2"
												></span>
												<div
													class="flex items-center gap-2 text-xs text-muted-foreground font-mono"
												>
													<span>{formatDateTime(update.created_at)}</span>
												</div>
												<p class="text-sm text-foreground/90 leading-relaxed">
													{update.message}
												</p>
											</div>
										{/each}
									</div>
								</div>
							{/if}
						</div>
					{/each}
				</div>
			</section>
		{/if}

		<!-- 5. Services & Systems Section (With Component Groups) -->
		<section class="space-y-4">
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-2">
					<Radio class="size-4 text-primary" />
					<h2 class="text-lg font-semibold tracking-tight text-foreground">Services & Systems</h2>
				</div>
				<span class="text-xs text-muted-foreground font-mono">
					{statusPage.monitors?.length || 0} component{statusPage.monitors?.length === 1 ? '' : 's'}
				</span>
			</div>

			{#if statusPage.monitors && statusPage.monitors.length > 0}
				{#if monitorGrouping.hasGroups}
					<!-- Grouped Component Sections -->
					<div class="space-y-6">
						{#each monitorGrouping.groups as gInfo (gInfo.group.id)}
							{@const isCollapsed = collapsedGroups[gInfo.group.id] ?? false}
							{@const gBadge = getGroupStatusBadge(gInfo.status)}

							<div class="rounded-xl border border-border/80 bg-white/3 overflow-hidden shadow-sm">
								<!-- Group Header (Collapsible Accordion Trigger) -->
								<button
									type="button"
									class="w-full px-5 py-4 flex items-center justify-between gap-3 bg-white/2 hover:bg-white/5 transition-colors border-b border-border/50 text-left cursor-pointer"
									onclick={() => toggleGroupCollapse(gInfo.group.id)}
								>
									<div class="flex items-center gap-3 min-w-0">
										<ChevronDown
											class="size-4 text-muted-foreground transition-transform duration-200 shrink-0 {isCollapsed
												? '-rotate-90'
												: ''}"
										/>
										<div class="flex items-center gap-2 min-w-0">
											<h3 class="text-base font-semibold text-foreground truncate">
												{gInfo.group.name}
											</h3>
											<span
												class="text-[11px] font-mono text-muted-foreground px-2 py-0.5 rounded-full bg-white/5 border border-border/60"
											>
												{gInfo.monitors.length}
											</span>
										</div>
									</div>

									<div
										class="flex items-center gap-1.5 text-xs font-medium px-2.5 py-1 rounded-full bg-white/5 border border-border/60 shrink-0"
									>
										<span style={gBadge.dotStyle} class="size-2 rounded-full {gBadge.dotClass}"></span>
										<span style={gBadge.textStyle} class={gBadge.textClass}>{gBadge.label}</span>
									</div>
								</button>

								<!-- Group Monitor Rows -->
								{#if !isCollapsed}
									<div class="p-4 sm:p-5 space-y-4">
										{#if gInfo.monitors.length === 0}
											<div class="text-xs text-muted-foreground py-3 text-center italic">
												No monitors currently in this group.
											</div>
										{:else}
											{#each gInfo.monitors as monitor}
												{@render monitorRow(monitor)}
											{/each}
										{/if}
									</div>
								{/if}
							</div>
						{/each}

						<!-- Ungrouped Monitors (if any) -->
						{#if monitorGrouping.ungrouped.length > 0}
							<div class="space-y-4 pt-2">
								<div class="flex items-center gap-2 px-1">
									<FolderTree class="size-3.5 text-muted-foreground" />
									<h3 class="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
										Other Components
									</h3>
								</div>
								<div class="space-y-4">
									{#each monitorGrouping.ungrouped as monitor}
										{@render monitorRow(monitor)}
									{/each}
								</div>
							</div>
						{/if}
					</div>
				{:else}
					<!-- Flat List Layout (Fallback when no groups exist) -->
					<div class="space-y-4">
						{#each statusPage.monitors as monitor}
							{@render monitorRow(monitor)}
						{/each}
					</div>
				{/if}
			{:else}
				<div
					class="bg-surface-container border-dashed border border-border rounded-xl p-8 text-center text-muted-foreground font-body-md"
				>
					No services are currently being monitored for this status page.
				</div>
			{/if}
		</section>

		<!-- Monitor Row Snippet -->
		{#snippet monitorRow(monitor: PublicMonitor)}
			{@const statusInfo = getMonitorStatusInfo(monitor.status)}
			<div
				class="group rounded-xl border border-border/70 hover:border-border/90 bg-white/3 p-5 sm:p-6 backdrop-blur-md transition-all duration-200 space-y-4"
			>
				<!-- Monitor Top Info Row -->
				<div class="flex flex-wrap items-center justify-between gap-3">
					<div class="flex items-center gap-2.5 min-w-0">
						<p
							class="text-base sm:text-lg font-semibold text-foreground truncate group-hover:text-primary transition-colors"
						>
							{monitor.name}
						</p>
						<span
							class="text-[10px] font-mono px-2 py-0.5 bg-surface-dim text-muted-foreground uppercase border border-outline-variant rounded-sm shrink-0"
						>
							{monitor.type}
						</span>
					</div>

					<div class="flex items-center gap-3 shrink-0">
						{#if monitor.recent_checks && monitor.recent_checks.length > 0}
							<span
								class="text-xs font-mono text-muted-foreground hidden sm:inline-flex items-center gap-1"
							>
								<Activity class="size-3 text-muted-foreground" />
								{monitor.recent_checks[0]}ms
							</span>
						{/if}
						<div
							class="flex items-center gap-1.5 text-xs font-medium px-2.5 py-1 rounded-full bg-white/5 border border-border/60"
						>
							<span style={statusInfo.dotStyle} class="size-2 rounded-full {statusInfo.dotClass}"></span>
							<span style={statusInfo.textStyle} class={statusInfo.textClass}>{statusInfo.label}</span>
						</div>
					</div>
				</div>

				<!-- 90-Day Uptime Bar Component -->
				<div class="pt-1">
					<UptimeBar dailyStats={monitor.daily_stats} overallUptime={monitor.uptime_percent} accentColor={themeColor} />
				</div>

				<!-- Latency Sparkline Footer -->
				{#if monitor.recent_checks && monitor.recent_checks.length > 0}
					<div
						class="pt-3 border-t border-border/40 flex items-center justify-between text-xs text-muted-foreground"
					>
						<span class="font-mono text-[11px]">Recent Latency</span>
						<div class="w-36 sm:w-48 h-6">
							<Sparkline
								data={monitor.recent_checks.slice().reverse()}
								color={statusInfo.sparklineColor}
								height={24}
							/>
						</div>
					</div>
				{/if}
			</div>
		{/snippet}

		<!-- 6. Incident History Archive (Browsable Timeline) -->
		<section class="space-y-4">
			<div class="flex flex-wrap items-center justify-between gap-3">
				<div class="flex items-center gap-2">
					<Archive class="size-4 text-muted-foreground" />
					<h2 class="text-lg font-semibold tracking-tight text-foreground">
						Past Incidents Archive
					</h2>
				</div>

				<!-- Date Range Filter Tabs -->
				<div class="flex items-center rounded-lg bg-white/5 p-1 border border-border/60 text-xs">
					<button
						type="button"
						style={historyRange === '7d' ? `background-color: ${themeColor}; color: #ffffff;` : ''}
						class="px-2.5 py-1 rounded-md transition-colors font-medium {historyRange === '7d'
							? 'shadow-sm'
							: 'text-muted-foreground hover:text-foreground'}"
						onclick={() => (historyRange = '7d')}
					>
						7 Days
					</button>
					<button
						type="button"
						style={historyRange === '14d' ? `background-color: ${themeColor}; color: #ffffff;` : ''}
						class="px-2.5 py-1 rounded-md transition-colors font-medium {historyRange === '14d'
							? 'shadow-sm'
							: 'text-muted-foreground hover:text-foreground'}"
						onclick={() => (historyRange = '14d')}
					>
						14 Days
					</button>
					<button
						type="button"
						style={historyRange === 'all' ? `background-color: ${themeColor}; color: #ffffff;` : ''}
						class="px-2.5 py-1 rounded-md transition-colors font-medium {historyRange === 'all'
							? 'shadow-sm'
							: 'text-muted-foreground hover:text-foreground'}"
						onclick={() => (historyRange = 'all')}
					>
						90 Days Archive
					</button>
				</div>
			</div>

			<!-- Daily Grouped Incidents Timeline -->
			{#if displayedIncidentHistory.length > 0}
				<div class="space-y-4">
					{#each displayedIncidentHistory as dayGroup (dayGroup.date)}
						<div
							class="rounded-xl border border-border/70 bg-white/3 backdrop-blur-md p-5 transition-all duration-200"
						>
							<div class="flex items-center justify-between pb-3 border-b border-border/40 mb-3">
								<h3 class="text-xs sm:text-sm font-semibold text-foreground font-mono">
									{formatDateHeader(dayGroup.date)}
								</h3>
								{#if dayGroup.incidents.length === 0}
									<span
										class="text-[11px] text-muted-foreground flex items-center gap-1.5 font-medium"
									>
										<ShieldCheck class="size-3.5" style="color: {themeColor};" />
										<span style="color: {themeColor};">No incidents</span>
									</span>
								{:else}
									<span
										class="text-[11px] font-mono px-2 py-0.5 rounded-full bg-amber-500/20 text-amber-300 border border-amber-500/30"
									>
										{dayGroup.incidents.length} incident{dayGroup.incidents.length > 1 ? 's' : ''}
									</span>
								{/if}
							</div>

							{#if dayGroup.incidents.length > 0}
								<div class="space-y-4 pt-1">
									{#each dayGroup.incidents as incident}
										<div class="rounded-lg border border-border/60 bg-black/10 p-4 space-y-2.5">
											<div class="flex flex-wrap items-center justify-between gap-2">
												<div class="flex items-center gap-2">
													<span
														style="background-color: color-mix(in srgb, {themeColor} 12%, transparent); border-color: color-mix(in srgb, {themeColor} 25%, transparent); color: {themeColor};"
														class="inline-flex items-center gap-1 text-[11px] font-mono uppercase px-2 py-0.5 rounded border"
													>
														<CheckCircle2 class="size-3" />
														Resolved
													</span>
													<span
														class="text-[11px] font-mono uppercase px-1.5 py-0.5 rounded border {getSeverityBadge(
															incident.severity
														)}"
													>
														{incident.severity}
													</span>
													{#if incident.monitor_name}
														<span class="text-xs text-muted-foreground">
															• {incident.monitor_name}
														</span>
													{/if}
												</div>

												<div
													class="flex items-center gap-2 text-xs text-muted-foreground font-mono"
												>
													{#if incident.started_at && incident.resolved_at}
														<span
															>Duration: {formatDuration(
																incident.started_at,
																incident.resolved_at
															)}</span
														>
														<span>•</span>
													{/if}
													<span>{formatDateTime(incident.resolved_at || incident.started_at)}</span>
												</div>
											</div>

											<h4 class="text-sm font-semibold text-foreground">
												{incident.summary}
											</h4>

											{#if incident.updates && incident.updates.length > 0}
												<p
													class="text-xs text-muted-foreground leading-relaxed pl-3 border-l-2 border-border/60"
												>
													{incident.updates[0].message}
												</p>
											{/if}
										</div>
									{/each}
								</div>
							{:else}
								<p class="text-xs text-muted-foreground leading-relaxed">
									No incidents reported today. All monitored systems operated normally.
								</p>
							{/if}
						</div>
					{/each}
				</div>
			{:else}
				<div
					class="rounded-xl border border-border/60 bg-white/2 backdrop-blur-md p-6 flex items-center gap-4 text-muted-foreground"
				>
					<div
						style="background-color: color-mix(in srgb, {themeColor} 12%, transparent); border-color: color-mix(in srgb, {themeColor} 25%, transparent);"
						class="size-9 rounded-full border flex items-center justify-center shrink-0"
					>
						<ShieldCheck class="size-5" style="color: {themeColor};" />
					</div>
					<div>
						<p class="text-sm font-medium text-foreground">No past incidents on record.</p>
						<p class="text-xs text-muted-foreground mt-0.5">
							All monitored systems have maintained normal operational status.
						</p>
					</div>
				</div>
			{/if}
		</section>
	</main>

	<!-- 7. Footer -->
	<footer class="w-full mt-auto border-t border-border/60 bg-background/50 backdrop-blur-sm">
		<div
			class="max-w-5xl mx-auto flex flex-col sm:flex-row justify-between items-center px-4 sm:px-6 py-6 w-full text-xs text-muted-foreground gap-3"
		>
			<span>
				{#if overallStatus === 'operational'}
					© 2026 {statusPage.title}. All systems operational.
				{:else if overallStatus === 'major_outage'}
					© 2026 {statusPage.title}. Systems currently experiencing a major outage.
				{:else}
					© 2026 {statusPage.title}. Some systems are experiencing issues.
				{/if}
			</span>
			<div class="flex items-center gap-1.5 font-mono">
				<span>Powered by</span>
				<span class="text-foreground font-semibold flex items-center gap-1">
					<img class="h-3.5 w-3.5 object-contain" src={logo} alt="" />
					PulseWatch
				</span>
			</div>
		</div>
	</footer>
</div>

<style>
	.pattern-bg {
		background-color: #0b0c0e;
		background-image: radial-gradient(rgba(255, 255, 255, 0.05) 1px, transparent 1px);
		background-size: 24px 24px;
	}
	:global(.light) .pattern-bg {
		background-color: #f8fafc;
		background-image: radial-gradient(#e2e8f0 1px, transparent 1px);
		background-size: 24px 24px;
	}
</style>
