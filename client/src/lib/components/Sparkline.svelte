<script lang="ts">
	let {
		data = [],
		color = '#4edea3',
		width = 50,
		height = 32
	}: {
		data: number[];
		color?: string;
		width?: number | string;
		height?: number;
	} = $props();

	// Calculate points for the SVG based on data
	let min = $derived(data.length ? Math.min(...data) : 0);
	let max = $derived(data.length ? Math.max(...data) : 100);
	let range = $derived(max - min === 0 ? 1 : max - min);

	// Generate the coordinate string for polyline
	let points = $derived(
		data
			.map((val: number, i: number) => {
				const x = (i / (data.length - 1 || 1)) * 100; // use 100 for viewBox percentage
				const y = height - ((val - min) / range) * height;
				return `${x},${y}`;
			})
			.join(' ')
	);

	// Generate a unique ID for the gradient to prevent collisions if multiple sparklines exist
	let gradientId = $derived(`sparkline-grad-${Math.random().toString(36).substring(2, 9)}`);
</script>

<!-- width: 100% makes it responsive to its container, viewBox maps 0-100 x coords -->
<svg
	{width}
	{height}
	viewBox="0 0 100 {height}"
	class="overflow-visible block w-full"
	preserveAspectRatio="none"
>
	<defs>
		<linearGradient id={gradientId} x1="0" x2="0" y1="0" y2="1">
			<stop offset="0%" stop-color={color} stop-opacity="0.15" />
			<stop offset="100%" stop-color={color} stop-opacity="0.0" />
		</linearGradient>
	</defs>

	<!-- Gradient Fill -->
	{#if points}
		<polygon points={`0,${height} ${points} 100,${height}`} fill={`url(#${gradientId})`} />
	{/if}

	<!-- Line Stroke -->
	<polyline
		{points}
		fill="none"
		stroke={color}
		stroke-width="1"
		stroke-linecap="round"
		stroke-linejoin="round"
	/>
</svg>
