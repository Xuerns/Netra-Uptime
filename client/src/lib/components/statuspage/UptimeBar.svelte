<script lang="ts">
	import type { PublicUptimeDay } from '$lib/api/statuspage';

	let {
		dailyStats = [],
		overallUptime = null,
		accentColor = null
	} = $props<{
		dailyStats?: PublicUptimeDay[];
		overallUptime?: number | null;
		accentColor?: string | null;
	}>();

	let hoveredIndex = $state<number | null>(null);
	let tooltipPos = $state<{ x: number } | null>(null);

	// Ensure we have exactly 90 days array
	let paddedStats = $derived.by(() => {
		if (dailyStats.length === 90) return dailyStats;
		const result: PublicUptimeDay[] = [];
		const now = new Date();
		const statMap = new Map<string, PublicUptimeDay>();
		for (const s of dailyStats) {
			statMap.set(s.date, s);
		}

		for (let i = 0; i < 90; i++) {
			const d = new Date(now);
			d.setDate(d.getDate() - (89 - i));
			const dateStr = d.toISOString().split('T')[0];
			if (statMap.has(dateStr)) {
				result.push(statMap.get(dateStr)!);
			} else {
				result.push({
					date: dateStr,
					uptime_percentage: 100,
					down_checks: 0,
					total_checks: 0,
					has_data: false
				});
			}
		}
		return result;
	});

	function getBarColor(day: PublicUptimeDay): string {
		if (!day.has_data) {
			return 'bg-white/10 hover:bg-white/20';
		}
		if (day.uptime_percentage >= 99.0) {
			return accentColor
				? ''
				: 'bg-emerald-500 hover:bg-emerald-400 shadow-[0_0_6px_rgba(16,185,129,0.25)]';
		}
		if (day.uptime_percentage >= 95.0) {
			return 'bg-amber-400 hover:bg-amber-300 shadow-[0_0_6px_rgba(251,191,36,0.3)]';
		}
		if (day.uptime_percentage >= 90.0) {
			return 'bg-orange-500 hover:bg-orange-400 shadow-[0_0_6px_rgba(249,115,22,0.3)]';
		}
		return 'bg-red-500 hover:bg-red-400 shadow-[0_0_6px_rgba(239,68,68,0.3)]';
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '';
		const [year, month, day] = dateStr.split('-');
		const d = new Date(Number(year), Number(month) - 1, Number(day));
		return d.toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	function handleMouseEnter(index: number, event: MouseEvent) {
		hoveredIndex = index;
		const target = event.currentTarget as HTMLElement;
		const parent = target.parentElement;
		if (parent) {
			const parentRect = parent.getBoundingClientRect();
			const targetRect = target.getBoundingClientRect();
			tooltipPos = {
				x: targetRect.left - parentRect.left + targetRect.width / 2
			};
		}
	}

	function handleMouseLeave() {
		hoveredIndex = null;
		tooltipPos = null;
	}
</script>

<div class="w-full space-y-2 select-none">
	<!-- Top Bar stats & Tooltip Container -->
	<div class="relative w-full">
		{#if hoveredIndex !== null && paddedStats[hoveredIndex] && tooltipPos}
			{@const day = paddedStats[hoveredIndex]}
			<div
				class="absolute -top-12 z-30 -translate-x-1/2 pointer-events-none transition-all duration-75"
				style="left: {Math.max(60, Math.min(tooltipPos.x, 700))}px"
			>
				<div
					class="bg-card/95 backdrop-blur-md border border-border px-2.5 py-1.5 rounded-lg shadow-xl text-center text-xs whitespace-nowrap"
				>
					<div class="font-semibold text-foreground">{formatDate(day.date)}</div>
					{#if day.has_data}
						<div
							class="text-[11px] font-mono text-muted-foreground flex items-center gap-1.5 justify-center mt-0.5"
						>
							<span
								style={day.uptime_percentage >= 99.0 && accentColor ? `color: ${accentColor};` : ''}
								class={day.uptime_percentage >= 99.0
									? (accentColor ? 'font-medium' : 'text-emerald-400')
									: day.uptime_percentage >= 95.0
										? 'text-amber-400'
										: 'text-red-400'}
							>
								{day.uptime_percentage.toFixed(2)}% uptime
							</span>
							{#if day.down_checks > 0}
								<span>•</span>
								<span class="text-red-400">{day.down_checks} failed checks</span>
							{/if}
						</div>
					{:else}
						<div class="text-[11px] text-muted-foreground mt-0.5">No checks recorded</div>
					{/if}
				</div>
			</div>
		{/if}

		<!-- The 90-Day Segments -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="flex items-center gap-[2px] sm:gap-[3px] w-full h-8 sm:h-9 py-1"
			onmouseleave={handleMouseLeave}
		>
			{#each paddedStats as day, i}
				<button
					type="button"
					aria-label={`Uptime for ${day.date}`}
					style={day.has_data && day.uptime_percentage >= 99.0 && accentColor
						? `background-color: ${accentColor}; box-shadow: 0 0 6px ${accentColor}40;`
						: ''}
					class="flex-1 h-full rounded-[2px] sm:rounded-sm transition-all duration-150 {getBarColor(
						day
					)} {hoveredIndex === i ? 'scale-y-125 z-10 brightness-125' : 'opacity-90'}"
					onmouseenter={(e) => handleMouseEnter(i, e)}
					tabindex="-1"
				></button>
			{/each}
		</div>
	</div>

	<!-- Bottom Axis / Scale -->
	<div
		class="flex justify-between items-center text-[11px] font-mono text-muted-foreground/80 px-0.5"
	>
		<span>90 days ago</span>
		{#if overallUptime !== null && overallUptime !== undefined}
			<span
				style={accentColor ? `color: ${accentColor};` : ''}
				class={accentColor ? 'font-medium' : 'text-emerald-400/90 font-medium'}
			>
				{overallUptime.toFixed(2)}% uptime
			</span>
		{:else}
			<span class="text-muted-foreground/50">—</span>
		{/if}
		<span>Today</span>
	</div>
</div>
