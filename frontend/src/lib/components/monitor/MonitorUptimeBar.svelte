<script lang="ts">
	interface DayStat {
		date: string;
		uptimePercentage: number;
		totalChecks: number;
		hasDowntime: boolean;
	}

	interface Props {
		stats?: DayStat[];
		daysCount?: number;
	}

	let { stats = [], daysCount = 90 }: Props = $props();

	// Generate empty days array filled with actual stats when available
	let bars = $derived.by(() => {
		const list: DayStat[] = [];
		const now = new Date();

		for (let i = daysCount - 1; i >= 0; i--) {
			const d = new Date(now);
			d.setDate(d.getDate() - i);
			const dateStr = d.toISOString().split('T')[0];

			const found = stats.find((s) => s.date.startsWith(dateStr));
			if (found) {
				list.push(found);
			} else {
				list.push({
					date: dateStr,
					uptimePercentage: 100,
					totalChecks: 0,
					hasDowntime: false
				});
			}
		}
		return list;
	});

	let hoveredBar = $state<DayStat | null>(null);
</script>

<div class="bg-[#161618] border border-white/8 rounded-2xl p-5 md:p-6 shadow-xl flex flex-col space-y-4">
	<div class="flex items-center justify-between min-h-7">
		<div class="flex items-center gap-2">
			<h3 class="text-sm font-semibold text-white tracking-tight">Daily Status Heatmap</h3>
			<span class="text-xs text-on-surface-variant font-mono">({daysCount} Days)</span>
		</div>

		<div class="min-h-7 flex items-center shrink-0">
			{#if hoveredBar}
				<div class="text-xs font-mono text-white bg-white/5 border border-white/10 px-2.5 py-1 rounded-md transition-all">
					{hoveredBar.date}: 
					{#if hoveredBar.totalChecks === 0}
						<span class="font-bold text-on-surface-variant">No Data</span>
					{:else}
						<span class="font-bold {hoveredBar.uptimePercentage < 99 ? 'text-rose-400' : 'text-emerald-400'}">{hoveredBar.uptimePercentage.toFixed(1)}% Uptime</span>
					{/if}
				</div>
			{:else}
				<span class="text-xs text-on-surface-variant/40 font-mono">Hover a bar for info</span>
			{/if}
		</div>
	</div>

	<!-- Status Bars Grid -->
	<div class="flex items-center gap-1 w-full overflow-x-auto py-2">
		{#each bars as bar}
			<div
				class="flex-1 min-w-1 h-9 rounded-sm transition-all cursor-pointer hover:opacity-80 hover:scale-y-110 {bar.totalChecks === 0
					? 'bg-white/10'
					: bar.uptimePercentage >= 99.5
						? 'bg-emerald-500/80'
						: bar.uptimePercentage >= 95
							? 'bg-amber-500/80'
							: 'bg-rose-500/90'}"
				role="button"
				tabindex="0"
				aria-label={`Status for ${bar.date}: ${bar.totalChecks === 0 ? 'No Data' : bar.uptimePercentage + '%'}`}
				onmouseenter={() => (hoveredBar = bar)}
				onmouseleave={() => (hoveredBar = null)}
			></div>
		{/each}
	</div>

	<!-- Legend -->
	<div class="flex items-center justify-between text-[11px] text-on-surface-variant pt-1 border-t border-white/5">
		<span>{daysCount} days ago</span>
		<div class="flex items-center gap-3">
			<div class="flex items-center gap-1.5">
				<span class="size-2.5 rounded-sm bg-white/10"></span>
				<span>No Data</span>
			</div>
			<div class="flex items-center gap-1.5">
				<span class="size-2.5 rounded-sm bg-emerald-500/80"></span>
				<span>Operational (99.5%+)</span>
			</div>
			<div class="flex items-center gap-1.5">
				<span class="size-2.5 rounded-sm bg-amber-500/80"></span>
				<span>Degraded</span>
			</div>
			<div class="flex items-center gap-1.5">
				<span class="size-2.5 rounded-sm bg-rose-500/90"></span>
				<span>Outage</span>
			</div>
		</div>
		<span>Today</span>
	</div>
</div>
