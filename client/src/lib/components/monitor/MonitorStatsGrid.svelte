<script lang="ts">
	import { Activity, TrendingUp, TrendingDown } from 'lucide-svelte';

	interface Props {
		uptimePercent: number;
		avgLatencyMs: number;
		minLatencyMs: number;
		maxLatencyMs: number;
		totalIncidents: number;
		downtimeFormatted?: string;
	}

	let {
		uptimePercent = 100,
		avgLatencyMs = 0,
		minLatencyMs = 0,
		maxLatencyMs = 0,
		totalIncidents = 0,
		downtimeFormatted = '0m'
	}: Props = $props();

	let uptimeBadgeColor = $derived.by(() => {
		if (uptimePercent >= 99) return 'text-emerald-400 border-emerald-500/20 bg-emerald-500/10';
		if (uptimePercent >= 95) return 'text-amber-400 border-amber-500/20 bg-amber-500/10';
		return 'text-rose-400 border-rose-500/20 bg-rose-500/10';
	});

	let latencyBadgeColor = $derived.by(() => {
		if (avgLatencyMs < 150) return 'text-emerald-400';
		if (avgLatencyMs < 400) return 'text-amber-400';
		return 'text-rose-400';
	});
</script>

<div class="grid grid-cols-2 md:grid-cols-4 gap-y-6">
	<!-- 1. Uptime Percentage -->
	<div class="flex flex-col gap-2 pr-4 md:pr-6 min-w-0">
		<span class="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider"
			>Uptime Rate</span
		>
		<div class="text-2xl md:text-3xl font-bold text-foreground tracking-tight">
			{uptimePercent.toFixed(2)}%
		</div>
		<div>
			<span class="px-2 py-0.5 rounded-full text-[11px] font-medium border {uptimeBadgeColor}">
				{uptimePercent >= 99.9 ? 'Excellent' : uptimePercent >= 98 ? 'Good' : 'Degraded'}
			</span>
		</div>
	</div>

	<!-- 2. Average Response Time -->
	<div class="flex flex-col gap-2 border-l border-border/60 pl-4 md:pl-6 md:pr-6 min-w-0">
		<span class="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider"
			>Avg Latency</span
		>
		<div class="text-2xl md:text-3xl font-bold text-foreground tracking-tight">
			{avgLatencyMs} <span class="text-base font-normal text-muted-foreground">ms</span>
		</div>
		<div class="flex items-center gap-1.5 text-xs {latencyBadgeColor}">
			<Activity class="size-3.5" />
			<span>{avgLatencyMs < 150 ? 'Fast Response' : avgLatencyMs < 400 ? 'Moderate' : 'Slow'}</span>
		</div>
	</div>

	<!-- 3. Min & Max Peak Latency -->
	<div class="flex flex-col gap-2 pr-4 md:border-l md:border-border/60 md:px-6 min-w-0">
		<span class="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider"
			>Min / Max Peak</span
		>
		<div class="flex items-center gap-2 flex-wrap text-base font-bold md:h-9">
			<div class="flex items-center gap-1 text-emerald-400">
				<TrendingDown class="size-4" />
				<span>{minLatencyMs}ms</span>
			</div>
			<span class="text-foreground/20">/</span>
			<div class="flex items-center gap-1 text-rose-400">
				<TrendingUp class="size-4" />
				<span>{maxLatencyMs}ms</span>
			</div>
		</div>
		<span class="text-xs text-muted-foreground">Best vs Worst Response Time</span>
	</div>

	<!-- 4. Incidents & Outage Duration -->
	<div class="flex flex-col gap-2 border-l border-border/60 pl-4 md:pl-6 min-w-0">
		<span class="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider"
			>Total Outages</span
		>
		<div class="text-2xl md:text-3xl font-bold text-foreground tracking-tight">
			{totalIncidents}
			<span class="text-sm font-normal text-muted-foreground"
				>incident{totalIncidents === 1 ? '' : 's'}</span
			>
		</div>
		<span class="text-xs text-muted-foreground">
			Total downtime: <span class="text-foreground font-medium">{downtimeFormatted}</span>
		</span>
	</div>
</div>
