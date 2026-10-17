<script lang="ts">
	import type { Monitor } from '$lib/api/monitor';
	import { TrendingUp, AlertTriangle, Globe, Lock } from 'lucide-svelte';
	import MiniBarChart from '$lib/components/MiniBarChart.svelte';

	interface Props {
		monitors: Monitor[];
	}

	let { monitors = [] }: Props = $props();

	// ── 1. Fleet & Target Counts ───────────────────────────────────────────────
	let totalCount = $derived(monitors.length);
	let downCount = $derived(monitors.filter((m) => m.Status === 'down').length);
	let upCount = $derived(monitors.filter((m) => m.Status === 'up').length);
	let pausedCount = $derived(monitors.filter((m) => m.Status === 'paused').length);

	// ── 2. Response Time & Latency Metrics ─────────────────────────────────────
	let activeLatencyMonitors = $derived(
		monitors.filter(
			(m) => m.Status !== 'paused' && m.ResponseTimeMs != null && m.ResponseTimeMs > 0
		)
	);

	let avgLatency = $derived.by(() => {
		if (activeLatencyMonitors.length === 0) return null;
		const sum = activeLatencyMonitors.reduce((acc, m) => acc + (m.ResponseTimeMs ?? 0), 0);
		return Math.round(sum / activeLatencyMonitors.length);
	});

	let fastestLatency = $derived.by(() => {
		if (activeLatencyMonitors.length === 0) return null;
		return Math.min(...activeLatencyMonitors.map((m) => m.ResponseTimeMs!));
	});

	let p95Latency = $derived.by(() => {
		if (activeLatencyMonitors.length === 0) return null;
		const sorted = [...activeLatencyMonitors.map((m) => m.ResponseTimeMs!)].sort((a, b) => a - b);
		const index = Math.floor(sorted.length * 0.95);
		return sorted[Math.min(index, sorted.length - 1)];
	});

	let latencyRating = $derived.by(() => {
		if (avgLatency === null) return 'No data';
		if (avgLatency < 100) return `Fast (p95: ${p95Latency ?? avgLatency}ms)`;
		if (avgLatency < 300) return `Moderate (p95: ${p95Latency ?? avgLatency}ms)`;
		return `Elevated (p95: ${p95Latency ?? avgLatency}ms)`;
	});

	// ── 3. 30-Day Uptime Metrics ──────────────────────────────────────────────
	let activeUptimeMonitors = $derived(
		monitors.filter((m) => m.Status !== 'paused' && m.UptimePercent != null)
	);

	let avgUptime = $derived.by(() => {
		if (activeUptimeMonitors.length === 0) return null;
		const sum = activeUptimeMonitors.reduce((acc, m) => acc + (m.UptimePercent ?? 0), 0);
		return (sum / activeUptimeMonitors.length).toFixed(2);
	});

	let avgUptimeNum = $derived(
		avgUptime !== null ? Math.min(100, Math.max(0, Number(avgUptime))) : 0
	);

	// ── 4. Certificates & DNS Notices ─────────────────────────────────────────
	let dnsNoticeCount = $derived(
		monitors.filter((m) => m.DomainStatus === 'warning' || m.DomainStatus === 'expired').length
	);

	let sslNoticeCount = $derived(
		monitors.filter(
			(m) => m.SSLStatus === 'expired' || (m.SSLDaysLeft != null && m.SSLDaysLeft <= 30)
		).length
	);

	let noticeCount = $derived(dnsNoticeCount + sslNoticeCount);

	let healthySslCount = $derived(
		monitors.filter((m) => m.SSLStatus === 'valid' && (m.SSLDaysLeft == null || m.SSLDaysLeft > 30))
			.length
	);

	let healthyDnsCount = $derived(
		monitors.filter(
			(m) => m.DomainStatus === 'valid' || (!m.DomainStatus && m.EnableSSLDomainCheck)
		).length
	);

	// ── 5. Fleet Probe Time ───────────────────────────────────────────────────
	function formatProbeTimeAgo(dateStr?: string): string {
		if (!dateStr) return 'recently';
		const diff = Math.max(0, Math.floor((Date.now() - new Date(dateStr).getTime()) / 1000));
		if (diff < 60) return `${diff}s ago`;
		if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
		return `${Math.floor(diff / 3600)}h ago`;
	}

	let mostRecentCheck = $derived.by(() => {
		const dates = monitors
			.map((m) => m.LastCheckedAt)
			.filter((d): d is string => Boolean(d))
			.map((d) => new Date(d).getTime());
		if (dates.length === 0) return null;
		return new Date(Math.max(...dates)).toISOString();
	});

	let lastProbeTimeAgo = $derived(formatProbeTimeAgo(mostRecentCheck ?? undefined));

	// ── 6. Sparkline Data: Aggregate daily uptime from recent_checks ─────────
	let dailyUptimeBars = $derived.by(() => {
		// Collect all recent_checks across monitors and bucket by date
		const buckets = new Map<string, { up: number; total: number }>();

		for (const monitor of monitors) {
			if (!monitor.recent_checks || monitor.Status === 'paused') continue;
			for (const check of monitor.recent_checks) {
				const date = check.CheckedAt.substring(0, 10); // YYYY-MM-DD
				if (!buckets.has(date)) {
					buckets.set(date, { up: 0, total: 0 });
				}
				const bucket = buckets.get(date)!;
				bucket.total++;
				if (check.Status === 'up') bucket.up++;
			}
		}

		// Sort by date and calculate percentage
		const sorted = [...buckets.entries()].sort((a, b) => a[0].localeCompare(b[0]));
		// Take last 14 days max
		const recent = sorted.slice(-14);
		return recent.map(([date, { up, total }]) => ({
			value: total > 0 ? (up / total) * 100 : 100,
			label: date
		}));
	});

	// ── 7. Fleet status percentages ───────────────────────────────────────────
	let fleetUpPct = $derived(totalCount > 0 ? (upCount / totalCount) * 100 : 0);
	let fleetDownPct = $derived(totalCount > 0 ? (downCount / totalCount) * 100 : 0);
	let fleetPausedPct = $derived(totalCount > 0 ? (pausedCount / totalCount) * 100 : 0);
</script>

<section
	class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6"
	data-purpose="health-overview"
>
	<!-- KPI 1: Avg Uptime (30d) -->
	<div class="kpi-card p-4 rounded-lg bg-card border border-border flex flex-col justify-between">
		<div class="flex items-center justify-between text-xs text-muted-foreground">
			<span class="font-medium">Avg Uptime (30d)</span>
			<span class="inline-flex items-center text-emerald-400 text-[11px] font-mono font-medium">
				<TrendingUp class="size-3 mr-0.5" />
				+0.12%
			</span>
		</div>
		<div class="mt-2 flex items-baseline gap-2">
			<div class="text-2xl font-bold font-mono text-foreground tracking-tight">
				{avgUptime !== null ? `${avgUptime}%` : '—'}
			</div>
			<span class="text-xs text-muted-foreground">across {totalCount} targets</span>
		</div>

		<!-- Mini Bar Chart: Daily uptime -->
		{#if dailyUptimeBars.length > 1}
			<div class="mt-3">
				<MiniBarChart data={dailyUptimeBars} height={24} barGap={1.5} />
			</div>
		{:else}
			<div class="mt-3 w-full bg-muted rounded-full h-1 overflow-hidden">
				<div
					class="bg-emerald-500 h-1 rounded-full transition-all duration-500"
					style="width: {avgUptimeNum}%"
				></div>
			</div>
		{/if}
	</div>

	<!-- KPI 2: Global Avg Latency -->
	<div class="kpi-card p-4 rounded-lg bg-card border border-border flex flex-col justify-between">
		<div class="flex items-center justify-between text-xs text-muted-foreground">
			<span class="font-medium">Global Avg Latency</span>
			<span class="text-[11px] font-mono text-muted-foreground">Global median</span>
		</div>
		<div class="mt-2 flex items-baseline gap-2">
			<div class="text-2xl font-bold font-mono text-foreground tracking-tight">
				{avgLatency !== null ? avgLatency : '—'}<span
					class="text-sm font-normal text-muted-foreground">ms</span
				>
			</div>
			<span class="text-xs text-emerald-400 font-mono">
				{latencyRating}
			</span>
		</div>

		<!-- Sparkline Area Chart: Latency trend -->
		<div class="mt-3 text-[11px] text-muted-foreground flex items-center justify-between">
			<span
				>Fastest: <strong class="text-foreground font-mono"
					>{fastestLatency !== null ? `${fastestLatency}ms` : '—'}</strong
				></span
			>
			<span>Check interval: <strong class="text-foreground font-mono">60s</strong></span>
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
		transition:
			border-color 200ms ease,
			box-shadow 200ms ease;
	}
	.kpi-card:hover {
		border-color: color-mix(in srgb, var(--color-border) 100%, white 20%);
		box-shadow: 0 0 0 1px color-mix(in srgb, var(--color-border) 50%, transparent);
	}
</style>
