<script lang="ts">
	import {
		Activity,
		Zap,
		AlertTriangle,
		ShieldCheck,
		TrendingUp,
		TrendingDown
	} from 'lucide-svelte';

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

<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
	<!-- 1. Uptime Percentage Card -->
	<div
		class="bg-card border border-border rounded-2xl p-5 shadow-lg flex flex-col justify-between space-y-3"
	>
		<div class="flex items-center justify-between">
			<span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider"
				>Uptime Rate</span
			>
			<div
				class="size-8 rounded-xl bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center"
			>
				<ShieldCheck class="size-4 text-emerald-400" />
			</div>
		</div>
		<div>
			<div class="text-2xl md:text-3xl font-bold text-foreground tracking-tight">
				{uptimePercent.toFixed(2)}%
			</div>
			<div class="flex items-center gap-2 mt-1">
				<span class="px-2 py-0.5 rounded-full text-[11px] font-medium border {uptimeBadgeColor}">
					{uptimePercent >= 99.9 ? 'Excellent' : uptimePercent >= 98 ? 'Good' : 'Degraded'}
				</span>
			</div>
		</div>
	</div>

	<!-- 2. Average Response Time Card -->
	<div
		class="bg-card border border-border rounded-2xl p-5 shadow-lg flex flex-col justify-between space-y-3"
	>
		<div class="flex items-center justify-between">
			<span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider"
				>Avg Latency</span
			>
			<div
				class="size-8 rounded-xl bg-sky-500/10 border border-sky-500/20 flex items-center justify-center"
			>
				<Zap class="size-4 text-sky-400" />
			</div>
		</div>
		<div>
			<div class="text-2xl md:text-3xl font-bold text-foreground tracking-tight">
				{avgLatencyMs} <span class="text-base font-normal text-muted-foreground">ms</span>
			</div>
			<div class="flex items-center gap-1.5 mt-1 text-xs {latencyBadgeColor}">
				<Activity class="size-3.5" />
				<span
					>{avgLatencyMs < 150 ? 'Fast Response' : avgLatencyMs < 400 ? 'Moderate' : 'Slow'}</span
				>
			</div>
		</div>
	</div>

	<!-- 3. Min & Max Peak Latency Card -->
	<div
		class="bg-card border border-border rounded-2xl p-5 shadow-lg flex flex-col justify-between space-y-3"
	>
		<div class="flex items-center justify-between">
			<span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider"
				>Min / Max Peak</span
			>
			<div
				class="size-8 rounded-xl bg-purple-500/10 border border-purple-500/20 flex items-center justify-center"
			>
				<TrendingUp class="size-4 text-purple-400" />
			</div>
		</div>
		<div>
			<div class="flex items-center gap-3 text-lg font-bold text-foreground">
				<div class="flex items-center gap-1 text-emerald-400 text-base">
					<TrendingDown class="size-4" />
					<span>{minLatencyMs}ms</span>
				</div>
				<span class="text-foreground/20">/</span>
				<div class="flex items-center gap-1 text-rose-400 text-base">
					<TrendingUp class="size-4" />
					<span>{maxLatencyMs}ms</span>
				</div>
			</div>
			<span class="text-xs text-muted-foreground block mt-1">Best vs Worst Response Time</span>
		</div>
	</div>

	<!-- 4. Incidents & Outage Duration Card -->
	<div
		class="bg-card border border-border rounded-2xl p-5 shadow-lg flex flex-col justify-between space-y-3"
	>
		<div class="flex items-center justify-between">
			<span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider"
				>Total Outages</span
			>
			<div
				class="size-8 rounded-xl bg-amber-500/10 border border-amber-500/20 flex items-center justify-center"
			>
				<AlertTriangle class="size-4 text-amber-400" />
			</div>
		</div>
		<div>
			<div class="text-2xl md:text-3xl font-bold text-foreground tracking-tight">
				{totalIncidents}
				<span class="text-sm font-normal text-muted-foreground"
					>incident{totalIncidents === 1 ? '' : 's'}</span
				>
			</div>
			<span class="text-xs text-muted-foreground block mt-1">
				Total downtime: <span class="text-foreground font-medium">{downtimeFormatted}</span>
			</span>
		</div>
	</div>
</div>
