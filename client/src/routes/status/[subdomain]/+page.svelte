<script lang="ts">
	import type { PageData } from './$types';
	import type { PublicStatusPage, PublicMonitor, PublicIncident, PublicMaintenance } from '$lib/api/statuspage';
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
		Calendar
	} from 'lucide-svelte';
	import logo from '$lib/assets/Logo.svg';
	import Sparkline from '$lib/components/Sparkline.svelte';
	import UptimeBar from '$lib/components/statuspage/UptimeBar.svelte';
	import MaintenanceCard from '$lib/components/statuspage/MaintenanceCard.svelte';

	let { data } = $props<{ data: PageData }>();
	let statusPage = $derived(data.statusPage as PublicStatusPage);

	let overallStatus = $derived(statusPage?.overall_status || 'operational');

	const statusConfig = $derived.by(() => {
		switch (overallStatus) {
			case 'major_outage':
				return {
					title: 'Major System Outage',
					description: 'We are currently experiencing severe disruption across our services.',
					bannerClass: 'bg-red-500/10 border-red-500/30 text-red-400 shadow-[0_0_30px_rgba(239,68,68,0.1)]',
					dotClass: 'bg-red-500 shadow-[0_0_14px_rgba(239,68,68,0.8)]',
					icon: XCircle,
					accentColor: '#ef4444'
				};
			case 'partial_outage':
				return {
					title: 'Partial System Outage',
					description: 'Some of our services are currently experiencing disruptions.',
					bannerClass: 'bg-orange-500/10 border-orange-500/30 text-orange-400 shadow-[0_0_30px_rgba(249,115,22,0.1)]',
					dotClass: 'bg-orange-500 shadow-[0_0_14px_rgba(249,115,22,0.8)]',
					icon: AlertTriangle,
					accentColor: '#f97316'
				};
			case 'degraded':
				return {
					title: 'Degraded Performance',
					description: 'We are experiencing elevated latency or minor service issues.',
					bannerClass: 'bg-amber-500/10 border-amber-500/30 text-amber-400 shadow-[0_0_30px_rgba(245,158,11,0.1)]',
					dotClass: 'bg-amber-500 shadow-[0_0_14px_rgba(245,158,11,0.8)]',
					icon: AlertCircle,
					accentColor: '#f59e0b'
				};
			default:
				return {
					title: 'All Systems Operational',
					description: 'All services are online and operating within optimal parameters.',
					bannerClass: 'bg-emerald-500/10 border-emerald-500/30 text-emerald-400 shadow-[0_0_30px_rgba(16,185,129,0.1)]',
					dotClass: 'bg-emerald-500 shadow-[0_0_14px_rgba(16,185,129,0.8)]',
					icon: CheckCircle2,
					accentColor: '#10b981'
				};
		}
	});

	function getMonitorStatusInfo(status: string) {
		switch (status) {
			case 'up':
				return {
					label: 'Operational',
					dotClass: 'bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.7)]',
					textClass: 'text-emerald-400',
					sparklineColor: '#10b981'
				};
			case 'degraded':
				return {
					label: 'Degraded',
					dotClass: 'bg-amber-500 shadow-[0_0_8px_rgba(245,158,11,0.7)]',
					textClass: 'text-amber-400',
					sparklineColor: '#f59e0b'
				};
			case 'paused':
				return {
					label: 'Paused',
					dotClass: 'bg-blue-400 shadow-[0_0_8px_rgba(96,165,250,0.7)]',
					textClass: 'text-blue-400',
					sparklineColor: '#60a5fa'
				};
			default:
				return {
					label: 'Down',
					dotClass: 'bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.7)]',
					textClass: 'text-red-400',
					sparklineColor: '#ef4444'
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
</svelte:head>

<div class="min-h-screen relative overflow-x-hidden text-foreground bg-background pattern-bg flex flex-col selection:bg-primary/20">
	<!-- Top Navigation Bar -->
	<header class="sticky top-0 w-full z-50 bg-background/80 backdrop-blur-md border-b border-border/60 transition-colors">
		<div class="max-w-5xl mx-auto flex justify-between items-center px-4 sm:px-6 py-4 w-full">
			<div class="flex items-center gap-3">
				<img class="h-8 w-8 object-contain" src={logo} alt="PulseWatch Logo" />
				<div class="flex flex-col">
					<div class="flex items-center gap-2">
						<span class="font-headline-md text-headline-md font-bold text-foreground">
							{statusPage.title}
						</span>
						<span class="text-[11px] font-mono px-2 py-0.5 rounded-full bg-surface-dim text-muted-foreground border border-outline-variant">
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
			<div class="flex items-center gap-3">
				<span class="inline-flex items-center gap-1.5 text-xs text-muted-foreground bg-white/5 border border-border/80 px-2.5 py-1 rounded-full">
					<span class="size-2 rounded-full {statusConfig.dotClass} animate-pulse"></span>
					Live
				</span>
			</div>
		</div>
	</header>

	<!-- Main Content Container -->
	<main class="flex-1 pb-16 px-4 sm:px-6 max-w-5xl mx-auto w-full space-y-10 mt-6">
		<!-- 1. Overall Status Banner -->
		<section class="rounded-xl border p-6 sm:p-7 backdrop-blur-md transition-all duration-300 {statusConfig.bannerClass}">
			<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
				<div class="flex items-center gap-4">
					<div class="relative flex items-center justify-center">
						<span class="size-3.5 rounded-full {statusConfig.dotClass} animate-ping absolute opacity-75"></span>
						<span class="size-3.5 rounded-full {statusConfig.dotClass} relative"></span>
					</div>
					<div>
						<h1 class="text-xl sm:text-2xl font-bold tracking-tight text-foreground flex items-center gap-2.5">
							{statusConfig.title}
						</h1>
						<p class="text-xs sm:text-sm text-muted-foreground mt-1">
							{statusConfig.description}
						</p>
					</div>
				</div>
				<div class="flex items-center gap-1.5 text-xs text-muted-foreground self-start sm:self-auto shrink-0 bg-black/20 sm:bg-transparent px-3 py-1.5 sm:p-0 rounded-lg">
					<Clock class="size-3.5 text-muted-foreground" />
					<span>Updated just now</span>
				</div>
			</div>
		</section>

		<!-- 2. Ongoing Maintenance Section (if active) -->
		{#if statusPage.ongoing_maintenances && statusPage.ongoing_maintenances.length > 0}
			<section class="space-y-4">
				<div class="flex items-center gap-2">
					<Wrench class="size-4 text-blue-400" />
					<h2 class="text-lg font-semibold tracking-tight text-foreground">
						Ongoing Maintenance
					</h2>
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
						<h2 class="text-lg font-semibold tracking-tight text-foreground">
							Active Incidents
						</h2>
						<span class="px-2 py-0.5 rounded-full text-xs font-mono font-medium bg-amber-500/20 text-amber-300 border border-amber-500/30">
							{statusPage.active_incidents.length}
						</span>
					</div>
				</div>

				<div class="space-y-4">
					{#each statusPage.active_incidents as incident}
						<div class="rounded-xl border border-border/80 bg-white/3 backdrop-blur-md overflow-hidden transition-all duration-200 hover:border-border">
							<!-- Incident Header -->
							<div class="p-5 sm:p-6 border-b border-border/40 space-y-3">
								<div class="flex flex-wrap items-center justify-between gap-2">
									<div class="flex items-center gap-2">
										<span class="text-[11px] font-mono uppercase tracking-wider px-2 py-0.5 rounded border {getSeverityBadge(incident.severity)}">
											{incident.severity}
										</span>
										<span class="text-[11px] font-mono uppercase tracking-wider px-2 py-0.5 rounded bg-amber-500/10 text-amber-400 border border-amber-500/20">
											Ongoing
										</span>
										{#if incident.monitor_name}
											<span class="text-xs text-muted-foreground flex items-center gap-1">
												• <span class="text-foreground/90 font-medium">{incident.monitor_name}</span>
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
									<div class="space-y-4 relative before:absolute before:left-2 before:top-2 before:bottom-2 before:w-[1px] before:bg-border/60">
										{#each incident.updates as update}
											<div class="relative pl-6 space-y-1">
												<span class="absolute left-1.5 top-1.5 size-1.5 rounded-full bg-amber-400 -translate-x-1/2"></span>
												<div class="flex items-center gap-2 text-xs text-muted-foreground font-mono">
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

		<!-- 5. Services / Systems Section (List Layout with 90-Day Uptime Bars) -->
		<section class="space-y-4">
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-2">
					<Radio class="size-4 text-primary" />
					<h2 class="text-lg font-semibold tracking-tight text-foreground">
						Services & Systems
					</h2>
				</div>
				<span class="text-xs text-muted-foreground">
					{statusPage.monitors?.length || 0} component{statusPage.monitors?.length === 1 ? '' : 's'}
				</span>
			</div>

			{#if statusPage.monitors && statusPage.monitors.length > 0}
				<div class="space-y-4">
					{#each statusPage.monitors as monitor}
						{@const statusInfo = getMonitorStatusInfo(monitor.status)}

						<div class="group rounded-xl border border-border/70 hover:border-border/90 bg-white/3 p-5 sm:p-6 backdrop-blur-md transition-all duration-200 space-y-4">
							<!-- Monitor Top Info Row -->
							<div class="flex flex-wrap items-center justify-between gap-3">
								<div class="flex items-center gap-2.5 min-w-0">
									<p class="text-base sm:text-lg font-semibold text-foreground truncate group-hover:text-primary transition-colors">
										{monitor.name}
									</p>
									<span class="text-[10px] font-mono px-2 py-0.5 bg-surface-dim text-muted-foreground uppercase border border-outline-variant rounded-sm shrink-0">
										{monitor.type}
									</span>
								</div>

								<div class="flex items-center gap-3 shrink-0">
									{#if monitor.recent_checks && monitor.recent_checks.length > 0}
										<span class="text-xs font-mono text-muted-foreground hidden sm:inline-flex items-center gap-1">
											<Activity class="size-3 text-muted-foreground" />
											{monitor.recent_checks[0]}ms
										</span>
									{/if}
									<div class="flex items-center gap-1.5 text-xs font-medium px-2.5 py-1 rounded-full bg-white/5 border border-border/60">
										<span class="size-2 rounded-full {statusInfo.dotClass}"></span>
										<span class={statusInfo.textClass}>{statusInfo.label}</span>
									</div>
								</div>
							</div>

							<!-- 90-Day Uptime Bar Component -->
							<div class="pt-1">
								<UptimeBar dailyStats={monitor.daily_stats} overallUptime={monitor.uptime_percent} />
							</div>

							<!-- Latency Sparkline Footer -->
							{#if monitor.recent_checks && monitor.recent_checks.length > 0}
								<div class="pt-3 border-t border-border/40 flex items-center justify-between text-xs text-muted-foreground">
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
					{/each}
				</div>
			{:else}
				<div class="bg-surface-container border-dashed border border-border rounded-xl p-8 text-center text-muted-foreground font-body-md">
					No services are currently being monitored for this status page.
				</div>
			{/if}
		</section>

		<!-- 6. Recent Incidents (Past 7 Days) Section -->
		<section class="space-y-4">
			<div class="flex items-center gap-2">
				<Clock class="size-4 text-muted-foreground" />
				<h2 class="text-lg font-semibold tracking-tight text-foreground">
					Incident History (Past 7 Days)
				</h2>
			</div>

			{#if statusPage.recent_incidents && statusPage.recent_incidents.length > 0}
				<div class="space-y-3">
					{#each statusPage.recent_incidents as incident}
						<div class="rounded-xl border border-border/70 bg-white/3 backdrop-blur-md p-5 transition-all duration-200 hover:border-border">
							<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 mb-2">
								<div class="flex items-center gap-2">
									<span class="inline-flex items-center gap-1 text-[11px] font-mono uppercase px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
										<CheckCircle2 class="size-3" />
										Resolved
									</span>
									<span class="text-[11px] font-mono uppercase px-1.5 py-0.5 rounded border {getSeverityBadge(incident.severity)}">
										{incident.severity}
									</span>
									{#if incident.monitor_name}
										<span class="text-xs text-muted-foreground">
											• {incident.monitor_name}
										</span>
									{/if}
								</div>
								<div class="flex items-center gap-3 text-xs text-muted-foreground font-mono">
									{#if incident.started_at && incident.resolved_at}
										<span>Downtime: {formatDuration(incident.started_at, incident.resolved_at)}</span>
										<span>•</span>
									{/if}
									<span>{formatDateTime(incident.resolved_at || incident.started_at)}</span>
								</div>
							</div>

							<h4 class="text-base font-semibold text-foreground">
								{incident.summary}
							</h4>

							{#if incident.updates && incident.updates.length > 0}
								<p class="text-xs sm:text-sm text-muted-foreground mt-2 leading-relaxed">
									{incident.updates[0].message}
								</p>
							{/if}
						</div>
					{/each}
				</div>
			{:else}
				<div class="rounded-xl border border-border/60 bg-white/2 backdrop-blur-md p-6 flex items-center gap-4 text-muted-foreground">
					<div class="size-9 rounded-full bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center shrink-0">
						<ShieldCheck class="size-5 text-emerald-400" />
					</div>
					<div>
						<p class="text-sm font-medium text-foreground">No incidents reported in the past 7 days.</p>
						<p class="text-xs text-muted-foreground mt-0.5">All monitored systems have maintained normal operational status.</p>
					</div>
				</div>
			{/if}
		</section>
	</main>

	<!-- 7. Footer -->
	<footer class="w-full mt-auto border-t border-border/60 bg-background/50 backdrop-blur-sm">
		<div class="max-w-5xl mx-auto flex flex-col sm:flex-row justify-between items-center px-4 sm:px-6 py-6 w-full text-xs text-muted-foreground gap-3">
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
