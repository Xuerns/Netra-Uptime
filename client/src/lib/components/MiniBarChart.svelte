<script lang="ts">
	interface BarData {
		value: number;
		label?: string;
	}

	let {
		data = [],
		height = 32,
		barGap = 1.5,
		baseColor = '#10b981',
		warningColor = '#f59e0b',
		dangerColor = '#ef4444'
	} = $props<{
		data: BarData[];
		height?: number;
		barGap?: number;
		baseColor?: string;
		warningColor?: string;
		dangerColor?: string;
	}>();

	function getBarColor(value: number): string {
		if (value >= 99) return baseColor;
		if (value >= 95) return warningColor;
		return dangerColor;
	}

	let barWidth = $derived(
		data.length > 0 ? Math.max(2, (100 - barGap * (data.length - 1)) / data.length) : 4
	);

	let bars = $derived(
		data.map((d: BarData, i: number) => {
			const x = i * (barWidth + barGap);
			const barH = Math.max(1.5, (d.value / 100) * height);
			const y = height - barH;
			return {
				x,
				y,
				width: barWidth,
				height: barH,
				color: getBarColor(d.value),
				value: d.value,
				label: d.label
			};
		})
	);

	let totalWidth = $derived(
		data.length > 0 ? data.length * barWidth + (data.length - 1) * barGap : 100
	);
</script>

<svg
	width="100%"
	{height}
	viewBox="0 0 {totalWidth} {height}"
	class="block w-full"
	preserveAspectRatio="none"
	role="img"
	aria-label="Daily uptime chart"
>
	{#each bars as bar, i (i)}
		<rect
			x={bar.x}
			y={bar.y}
			width={bar.width}
			height={bar.height}
			fill={bar.color}
			rx="0.8"
			opacity="0.85"
			class="mini-bar"
		>
			{#if bar.label}
				<title>{bar.label}: {bar.value.toFixed(1)}%</title>
			{/if}
		</rect>
	{/each}
</svg>

<style>
	.mini-bar {
		transition: opacity 150ms ease;
	}
	.mini-bar:hover {
		opacity: 1;
	}
</style>
