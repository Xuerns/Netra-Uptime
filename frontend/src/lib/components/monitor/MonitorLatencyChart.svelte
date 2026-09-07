<script lang="ts">
	interface DataPoint {
		label: string;
		value: number; // response time in ms
		status?: 'up' | 'down';
	}

	interface Props {
		data: DataPoint[];
		title?: string;
		height?: number;
		color?: string;
	}

	let {
		data = [],
		title = 'Response Time (ms)',
		height = 220,
		color = '#10B981'
	}: Props = $props();

	let hoverIndex = $state<number | null>(null);

	// Compute SVG path & coordinates
	let points = $derived.by(() => {
		if (!data || data.length === 0) return [];
		const maxVal = Math.max(...data.map((d) => d.value), 10);
		const minVal = 0;
		const padding = 20;
		const svgWidth = 800;
		const svgHeight = height;

		const stepX = (svgWidth - padding * 2) / Math.max(data.length - 1, 1);

		return data.map((d, i) => {
			const x = padding + i * stepX;
			const y =
				svgHeight - padding - ((d.value - minVal) / (maxVal - minVal)) * (svgHeight - padding * 2);
			return { x, y, label: d.label, value: d.value, status: d.status || 'up' };
		});
	});

	let pathD = $derived.by(() => {
		if (points.length === 0) return '';
		return points.reduce(
			(acc, p, i) => (i === 0 ? `M ${p.x} ${p.y}` : `${acc} L ${p.x} ${p.y}`),
			''
		);
	});

	let areaD = $derived.by(() => {
		if (points.length === 0) return '';
		const first = points[0];
		const last = points[points.length - 1];
		return `${pathD} L ${last.x} ${height - 20} L ${first.x} ${height - 20} Z`;
	});

	let activePoint = $derived(hoverIndex !== null && points[hoverIndex] ? points[hoverIndex] : null);
</script>

<div class="bg-card border border-border rounded-2xl p-5 md:p-6 shadow-xl flex flex-col space-y-4">
	<div class="flex items-center justify-between min-h-7">
		<h3 class="text-sm font-semibold text-foreground tracking-tight flex items-center gap-2">
			<span class="size-2 rounded-full bg-emerald-400 shadow-[0_0_8px_rgba(16,185,129,0.5)]"></span>
			<span>{title}</span>
		</h3>

		<div class="min-h-7 flex items-center shrink-0">
			{#if activePoint}
				<div
					class="text-xs font-mono text-emerald-400 bg-emerald-500/10 border border-emerald-500/20 px-2.5 py-1 rounded-md"
				>
					{activePoint.label}: <span class="font-bold">{activePoint.value} ms</span>
				</div>
			{:else}
				<span class="text-xs text-muted-foreground/40 font-mono">Hover a point for info</span>
			{/if}
		</div>
	</div>

	{#if points.length === 0}
		<div class="h-48 flex items-center justify-center text-muted-foreground text-xs">
			No check data available for the selected period
		</div>
	{:else}
		<div class="relative w-full overflow-hidden" style="height: {height}px;">
			<svg
				viewBox="0 0 800 {height}"
				preserveAspectRatio="none"
				class="w-full h-full overflow-visible"
			>
				<defs>
					<linearGradient id="chartGradient" x1="0" y1="0" x2="0" y2="1">
						<stop offset="0%" stop-color={color} stop-opacity="0.35" />
						<stop offset="100%" stop-color={color} stop-opacity="0.0" />
					</linearGradient>
				</defs>

				<!-- Horizontal Grid Lines -->
				<line x1="20" y1="20" x2="780" y2="20" stroke="var(--chart-grid)" stroke-dasharray="4 4" />
				<line
					x1="20"
					y1={height / 2}
					x2="780"
					y2={height / 2}
					stroke="var(--chart-grid)"
					stroke-dasharray="4 4"
				/>
				<line x1="20" y1={height - 20} x2="780" y2={height - 20} stroke="var(--chart-grid)" />

				<!-- Area Fill -->
				<path d={areaD} fill="url(#chartGradient)" />

				<!-- Stroke Line -->
				<path
					d={pathD}
					fill="none"
					stroke={color}
					stroke-width="2.5"
					stroke-linejoin="round"
					stroke-linecap="round"
				/>

				<!-- Hover Circles & Points -->
				{#each points as p, i}
					<circle
						cx={p.x}
						cy={p.y}
						r={hoverIndex === i ? 5 : 2.5}
						class="{p.status === 'down'
							? 'fill-rose-500'
							: 'fill-emerald-400'} stroke-card transition-all cursor-pointer"
						stroke-width="2"
						role="button"
						tabindex="0"
						aria-label={`Data point ${p.label}: ${p.value} ms`}
						onmouseenter={() => (hoverIndex = i)}
						onmouseleave={() => (hoverIndex = null)}
					/>
				{/each}

				<!-- Active Line Pointer -->
				{#if activePoint}
					<line
						x1={activePoint.x}
						y1="10"
						x2={activePoint.x}
						y2={height - 20}
						stroke="rgba(255,255,255,0.2)"
						stroke-dasharray="2 2"
					/>
				{/if}
			</svg>
		</div>
	{/if}
</div>
