<script lang="ts">
	import type { Monitor, FleetOverviewResponse } from '$lib/api/monitor';
	import { TrendingUp, TrendingDown, Minus, AlertTriangle, Globe, Lock } from 'lucide-svelte';
	import { formatTimeAgo } from '$lib/utils/datetime';

	interface Props {
		monitors: Monitor[];
		overview?: FleetOverviewResponse | null;
	}

	let { monitors = [], overview = null }: Props = $props();

	// ── 1. Fleet & Target Counts ───────────────────────────────────────────────
	let totalCount = $derived(overview?.fleet?.total_targets ?? monitors.length);
	let downCount = $derived(
		overview?.fleet?.down ??
			monitors.filter((m) => m.Status === 'down' || m.Status === 'degraded').length
	);
	let upCount = $derived(overview?.fleet?.up ?? monitors.filter((m) => m.Status === 'up').length);
	let pausedCount = $derived(
		overview?.fleet?.paused ?? monitors.filter((m) => m.Status === 'paused').length
	);

	// ── 2. Response Time & Latency Metrics ─────────────────────────────────────
	let activeLatencyMonitors = $derived(
		monitors.filter(
			(m) => m.Status !== 'paused' && m.ResponseTimeMs != null && m.ResponseTimeMs > 0
		)
	);

	let avgLatency = $derived.by(() => {
		if (overview?.latency?.avg_ms != null) return overview.latency.avg_ms;
		if (activeLatencyMonitors.length === 0) return null;
		const sum = activeLatencyMonitors.reduce((acc, m) => acc + (m.ResponseTimeMs ?? 0), 0);
		return Math.round(sum / activeLatencyMonitors.length);
	});

	let fastestLatency = $derived.by(() => {
		if (overview?.latency?.fastest_ms != null) return overview.latency.fastest_ms;
		if (activeLatencyMonitors.length === 0) return null;
		return Math.min(...activeLatencyMonitors.map((m) => m.ResponseTimeMs!));
	});

	let p95Latency = $derived.by(() => {
		if (overview?.latency?.p95_ms != null) return overview.latency.p95_ms;
		if (activeLatencyMonitors.length === 0) return null;
		const sorted = [...activeLatencyMonitors.map((m) => m.ResponseTimeMs!)].sort((a, b) => a - b);
		const index = Math.floor(sorted.length * 0.95);
		return sorted[Math.min(index, sorted.length - 1)];
	});

	let latencyRating = $derived.by(() => {
		if (avgLatency === null) return 'No data';
		const p95 = p95Latency ?? avgLatency;
		if (avgLatency < 100) return `Fast (p95: ${p95}ms)`;
		if (avgLatency < 300) return `Moderate (p95: ${p95}ms)`;
		return `Elevated (p95: ${p95}ms)`;
	});

	let checkInterval = $derived.by(() => {
		if (overview?.fleet?.intervals) {
			const { is_uniform, mode_sec, min_sec, max_sec } = overview.fleet.intervals;
			return is_uniform ? `${mode_sec}s` : `${min_sec}s–${max_sec}s`;
		}
		const intervals = monitors.map((m) => m.IntervalSeconds).filter(Boolean);
		if (intervals.length === 0) return '—';
		const min = Math.min(...intervals);
		const max = Math.max(...intervals);
		return min === max ? `${min}s` : `${min}s–${max}s`;
	});

	// ── 3. 30-Day Uptime Metrics ──────────────────────────────────────────────
	let activeUptimeMonitors = $derived(
		monitors.filter((m) => m.Status !== 'paused' && m.UptimePercent != null)
	);

	let avgUptime = $derived.by(() => {
		if (overview?.uptime?.current_percent != null) {
			return `${overview.uptime.current_percent.toFixed(2)}%`;
		}
		if (activeUptimeMonitors.length === 0) return null;
		const sum = activeUptimeMonitors.reduce((acc, m) => acc + (m.UptimePercent ?? 0), 0);
		return `${(sum / activeUptimeMonitors.length).toFixed(2)}%`;
	});

	// ── 4. Certificates & DNS Notices ─────────────────────────────────────────
	let dnsNoticeCount = $derived(
		overview?.certificates_and_dns?.dns_warnings ??
			monitors.filter((m) => m.DomainStatus === 'warning' || m.DomainStatus === 'expired').length
	);

	let sslNoticeCount = $derived(
		overview?.certificates_and_dns?.ssl_warnings ??
			monitors.filter(
				(m) => m.SSLStatus === 'expired' || (m.SSLDaysLeft != null && m.SSLDaysLeft <= 30)
			).length
	);

	let noticeCount = $derived(
		overview?.certificates_and_dns?.total_notices ?? dnsNoticeCount + sslNoticeCount
	);

	let healthySslCount = $derived(
		overview?.certificates_and_dns?.ssl_healthy ??
			monitors.filter(
				(m) => m.SSLStatus === 'valid' && (m.SSLDaysLeft == null || m.SSLDaysLeft > 30)
			).length
	);

	let healthyDnsCount = $derived(
		overview?.certificates_and_dns?.dns_healthy ??
			monitors.filter(
				(m) => m.DomainStatus === 'valid' || (!m.DomainStatus && m.EnableSSLDomainCheck)
			).length
	);

	// ── 5. Fleet Probe Time ───────────────────────────────────────────────────
	let mostRecentCheck = $derived.by(() => {
		const dates = monitors
			.map((m) => m.LastCheckedAt)
			.filter((d): d is string => Boolean(d))
			.map((d) => new Date(d).getTime());
		if (dates.length === 0) return null;
		return new Date(Math.max(...dates)).toISOString();
	});

	let lastProbeTimeAgo = $derived(mostRecentCheck ? formatTimeAgo(mostRecentCheck) : 'recently');

	// ── 6. Fleet status percentages ───────────────────────────────────────────
	let fleetUpPct = $derived(totalCount > 0 ? (upCount / totalCount) * 100 : 0);
	let fleetDownPct = $derived(totalCount > 0 ? (downCount / totalCount) * 100 : 0);
	let fleetPausedPct = $derived(totalCount > 0 ? (pausedCount / totalCount) * 100 : 0);
</script>

<section
	class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6"
	data-purpose="health-overview"
>
	<!-- KPI 1: Avg Uptime -->
	<div class="kpi-card p-4 rounded-lg bg-card border border-border flex flex-col justify-between">
		<div class="flex items-center justify-between text-xs text-muted-foreground">
			<span class="font-medium">Avg Uptime ({overview?.window_days ?? 30}d)</span>
			{#if overview?.uptime?.trend_delta != null}
				{#if overview.uptime.trend_delta > 0}
					<span class="inline-flex items-center text-emerald-400 text-[11px] font-mono font-medium">
						<TrendingUp class="size-3 mr-0.5" />
						+{overview.uptime.trend_delta.toFixed(2)}%
					</span>
				{:else if overview.uptime.trend_delta < 0}
					<span class="inline-flex items-center text-rose-400 text-[11px] font-mono font-medium">
						<TrendingDown class="size-3 mr-0.5" />
						{overview.uptime.trend_delta.toFixed(2)}%
					</span>
				{:else}
					<span
						class="inline-flex items-center text-muted-foreground text-[11px] font-mono font-medium"
					>
						<Minus class="size-3 mr-0.5" />
						0.00%
					</span>
				{/if}
			{:else}
				<span class="text-[10px] text-muted-foreground/70 font-mono">No prior period</span>
			{/if}
		</div>
		<div class="mt-2 flex items-baseline gap-2">
			<div class="text-2xl font-bold font-mono text-foreground tracking-tight">
				{avgUptime ?? '—'}
			</div>
			<span class="text-xs text-muted-foreground">across {totalCount} targets</span>
		</div>

		<!-- Summary Info -->
		<div class="mt-3 text-[11px] text-muted-foreground flex items-center justify-between">
			<span
				>Evaluation window: <strong class="text-foreground font-mono"
					>{overview?.window_days ?? 30}d</strong
				></span
			>
			<span
				>Healthy: <strong class="text-emerald-400 font-mono">{upCount}/{totalCount}</strong></span
			>
		</div>
	</div>

	<!-- KPI 2: Global Avg Latency -->
	<div class="kpi-card p-4 rounded-lg bg-card border border-border flex flex-col justify-between">
		<div class="flex items-center justify-between text-xs text-muted-foreground">
			<span class="font-medium">Global Avg Latency</span>
			<span class="text-[11px] font-mono text-muted-foreground">24h P50 / P95 (Live)</span>
		</div>
		<div class="mt-2 flex items-baseline gap-2">
			<div class="text-2xl font-bold font-mono text-foreground tracking-tight">
				{#if avgLatency !== null}
					{avgLatency}<span class="text-sm font-normal text-muted-foreground">ms</span>
				{:else}
					—
				{/if}
			</div>
			<span class="text-xs text-emerald-400 font-mono">
				{latencyRating}
			</span>
		</div>

		<!-- Latency Sub-metrics -->
		<div class="mt-3 text-[11px] text-muted-foreground flex items-center justify-between">
			<span
				>Fastest: <strong class="text-foreground font-mono"
					>{fastestLatency !== null ? `${fastestLatency}ms` : '—'}</strong
				></span
			>
			<span>Check interval: <strong class="text-foreground font-mono">{checkInterval}</strong></span
			>
		</div>
	</div>

	<!-- KPI 3: Certificates & DNS -->
	<div class="kpi-card p-4 rounded-lg bg-card border border-border flex flex-col justify-between">
		<div class="flex items-center justify-between text-xs text-muted-foreground">
			<span class="font-medium">Certificates &amp; DNS</span>
			{#if noticeCount > 0}
				<span
					class="text-[10px] font-mono px-1.5 py-0.5 rounded bg-amber-500/10 text-amber-400 border border-amber-500/20 font-medium inline-flex items-center gap-1"
				>
					<AlertTriangle class="size-2.5" />
					{noticeCount}
					{noticeCount === 1 ? 'Notice' : 'Notices'}
				</span>
			{:else}
				<span
					class="text-[10px] font-mono px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-medium"
				>
					Healthy
				</span>
			{/if}
		</div>
		<div class="mt-2 flex items-baseline gap-2">
			<div
				class="text-2xl font-bold font-mono {noticeCount > 0
					? 'text-amber-400'
					: 'text-foreground'} tracking-tight"
			>
				{noticeCount}
			</div>
			<span class="text-xs text-muted-foreground">
				{noticeCount > 0 ? 'DNS / SSL notices' : 'DNS & SSL verified'}
			</span>
		</div>

		<!-- Detailed breakdown instead of single line -->
		<div class="mt-3 flex flex-col gap-1.5">
			<div class="flex items-center justify-between text-[11px]">
				<span class="inline-flex items-center gap-1.5 text-muted-foreground">
					<Lock class="size-3 text-emerald-400/70" />
					SSL
				</span>
				<span class="font-mono">
					<span class="text-emerald-400">{healthySslCount}</span>
					<span class="text-muted-foreground/60 mx-0.5">·</span>
					{#if sslNoticeCount > 0}
						<span class="text-amber-400">{sslNoticeCount} warning</span>
					{:else}
						<span class="text-muted-foreground">0 issues</span>
					{/if}
				</span>
			</div>
			<div class="flex items-center justify-between text-[11px]">
				<span class="inline-flex items-center gap-1.5 text-muted-foreground">
					<Globe class="size-3 text-emerald-400/70" />
					DNS
				</span>
				<span class="font-mono">
					<span class="text-emerald-400">{healthyDnsCount}</span>
					<span class="text-muted-foreground/60 mx-0.5">·</span>
					{#if dnsNoticeCount > 0}
						<span class="text-amber-400">{dnsNoticeCount} warning</span>
					{:else}
						<span class="text-muted-foreground">0 issues</span>
					{/if}
				</span>
			</div>
		</div>
	</div>

	<!-- KPI 4: Fleet Engine -->
	<div class="kpi-card p-4 rounded-lg bg-card border border-border flex flex-col justify-between">
		<div class="flex items-center justify-between text-xs text-muted-foreground">
			<span class="font-medium">Fleet Engine</span>
			<span
				class="flex items-center gap-1 text-[11px] {downCount > 0
					? 'text-rose-400'
					: 'text-emerald-400'} font-mono font-medium"
			>
				<span class="size-1.5 rounded-full {downCount > 0 ? 'bg-rose-400' : 'bg-emerald-400'}"
				></span>
				{downCount > 0 ? 'Degraded' : 'Synced'}
			</span>
		</div>
		<div class="mt-2 flex items-baseline gap-2">
			<div class="text-2xl font-bold font-mono text-foreground tracking-tight">
				{upCount} / {totalCount}
			</div>
			<span class="text-xs text-muted-foreground">targets alive</span>
		</div>

		<!-- Segmented status bar -->
		<div class="mt-3 flex flex-col gap-1.5">
			<div class="w-full h-1.5 rounded-full bg-muted overflow-hidden flex">
				{#if fleetUpPct > 0}
					<div
						class="h-full bg-emerald-500 transition-all duration-500"
						style="width: {fleetUpPct}%"
					></div>
				{/if}
				{#if fleetDownPct > 0}
					<div
						class="h-full bg-rose-500 transition-all duration-500"
						style="width: {fleetDownPct}%"
					></div>
				{/if}
				{#if fleetPausedPct > 0}
					<div
						class="h-full bg-zinc-500 transition-all duration-500"
						style="width: {fleetPausedPct}%"
					></div>
				{/if}
			</div>
			<div class="flex items-center gap-3 text-[11px] text-muted-foreground">
				<span class="inline-flex items-center gap-1">
					<span class="size-1.5 rounded-full bg-emerald-500"></span>
					<span class="font-mono">{upCount}</span> up
				</span>
				{#if downCount > 0}
					<span class="inline-flex items-center gap-1">
						<span class="size-1.5 rounded-full bg-rose-500"></span>
						<span class="font-mono text-rose-400">{downCount}</span> down
					</span>
				{/if}
				{#if pausedCount > 0}
					<span class="inline-flex items-center gap-1">
						<span class="size-1.5 rounded-full bg-zinc-500"></span>
						<span class="font-mono">{pausedCount}</span> paused
					</span>
				{/if}
				<span class="ml-auto text-muted-foreground/70 font-mono">{lastProbeTimeAgo}</span>
			</div>
		</div>
	</div>
</section>

<style>
	.kpi-card {
		transition: border-color 200ms ease;
	}
	.kpi-card:hover {
		border-color: color-mix(in srgb, var(--primary) 35%, var(--color-border));
	}
</style>
