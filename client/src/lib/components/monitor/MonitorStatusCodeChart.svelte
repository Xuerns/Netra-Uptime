<script lang="ts">
	import { onDestroy } from 'svelte';
	import Chart from 'chart.js/auto';

	interface Props {
		count2xx: number;
		count3xx: number;
		count4xx: number;
		count5xx: number;
		countError: number;
		totalChecks: number;
		periodLabel?: string;
	}

	let {
		count2xx = 0,
		count3xx = 0,
		count4xx = 0,
		count5xx = 0,
		countError = 0,
		totalChecks = 0,
		periodLabel = '30D'
	}: Props = $props();

	type ViewMode = 'percentage' | 'count';
	let viewMode = $state<ViewMode>('percentage');

	let canvasEl = $state<HTMLCanvasElement | null>(null);
	let chartInstance: Chart | null = null;

	let effectiveTotal = $derived(
		totalChecks > 0 ? totalChecks : count2xx + count3xx + count4xx + count5xx + countError
	);

	let rawCounts = $derived([count2xx, count3xx, count4xx, count5xx, countError]);

	let percentages = $derived.by(() => {
		const total = effectiveTotal > 0 ? effectiveTotal : 1;
		return rawCounts.map((cnt) => (cnt / total) * 100);
	});

	const categories = [
		{
			id: '2xx',
			label: '2xx Success',
			short: '2xx',
			subtext: 'OK / Created',
			desc: 'Status 200–299: Successful responses',
			color: '#10B981',
			hoverColor: '#34D399'
		},
		{
			id: '3xx',
			label: '3xx Redirect',
			short: '3xx',
			subtext: 'Moved / Found',
			desc: 'Status 300–399: Redirections',
			color: '#3B82F6',
			hoverColor: '#60A5FA'
		},
		{
			id: '4xx',
			label: '4xx Client Err',
			short: '4xx',
			subtext: 'Bad Req / 404',
			desc: 'Status 400–499: Client errors / Not found',
			color: '#F59E0B',
			hoverColor: '#FBBF24'
		},
		{
			id: '5xx',
			label: '5xx Server Err',
			short: '5xx',
			subtext: '500 / 502 / 503',
			desc: 'Status 500–599: Internal server errors',
			color: '#EF4444',
			hoverColor: '#F87171'
		},
		{
			id: 'error',
			label: 'Timeout / Err',
			short: 'Err',
			subtext: 'No Response',
			desc: 'Connection timeouts or dropped requests',
			color: '#A855F7',
			hoverColor: '#C084FC'
		}
	];

	function createOrUpdateChart() {
		if (!canvasEl) return;

		const currentData = viewMode === 'percentage' ? percentages : rawCounts;
		const isDark =
			typeof document !== 'undefined' && document.documentElement.classList.contains('dark');
		const gridColor = isDark ? 'rgba(255, 255, 255, 0.06)' : 'rgba(15, 23, 42, 0.06)';
		const xTickColor = isDark ? '#94a3b8' : '#475569';

		if (!chartInstance) {
			chartInstance = new Chart(canvasEl, {
				type: 'bar',
				data: {
					labels: categories.map((c) => c.short),
					datasets: [
						{
							data: currentData,
							backgroundColor: categories.map((c) => c.color),
							hoverBackgroundColor: categories.map((c) => c.hoverColor),
							borderRadius: 6,
							borderSkipped: 'bottom',
							maxBarThickness: 40
						}
					]
				},
				options: {
					responsive: true,
					maintainAspectRatio: false,
					animation: {
						duration: 400
					},
					plugins: {
						legend: {
							display: false
						},
						tooltip: {
							backgroundColor: 'rgba(15, 23, 42, 0.95)',
							titleColor: '#f8fafc',
							bodyColor: '#cbd5e1',
							borderColor: 'rgba(255, 255, 255, 0.12)',
							borderWidth: 1,
							padding: 12,
							cornerRadius: 8,
							displayColors: true,
							boxWidth: 8,
							boxHeight: 8,
							titleFont: {
								family: 'monospace',
								size: 13,
								weight: 'bold'
							},
							bodyFont: {
								family: 'monospace',
								size: 12
							},
							callbacks: {
								title: (items) => {
									const idx = items[0]?.dataIndex ?? 0;
									return `${categories[idx].label} (${categories[idx].subtext})`;
								},
								label: (item) => {
									const idx = item.dataIndex;
									const count = rawCounts[idx];
									const pct = percentages[idx];
									return `Checks: ${count.toLocaleString()} (${pct.toFixed(2)}%)`;
								},
								afterLabel: (item) => {
									const idx = item.dataIndex;
									return categories[idx].desc;
								}
							}
						}
					},
					scales: {
						x: {
							grid: {
								display: false
							},
							ticks: {
								color: xTickColor,
								font: {
									family: 'monospace',
									size: 11,
									weight: 'bold'
								}
							}
						},
						y: {
							min: 0,
							max: viewMode === 'percentage' ? 100 : undefined,
							grid: {
								color: gridColor
							},
							ticks: {
								color: '#64748b',
								font: {
									family: 'monospace',
									size: 11
								},
								callback: (val) => {
									return viewMode === 'percentage' ? `${val}%` : Number(val).toLocaleString();
								}
							}
						}
					}
				}
			});
		} else {
			chartInstance.data.datasets[0].data = currentData;
			if (chartInstance.options.scales?.y) {
				chartInstance.options.scales.y.max = viewMode === 'percentage' ? 100 : undefined;
			}
			chartInstance.update();
		}
	}

	$effect(() => {
		// Reactive trigger on canvas element, viewMode, or counts change
		if (canvasEl && effectiveTotal >= 0) {
			createOrUpdateChart();
		}
	});

	onDestroy(() => {
		if (chartInstance) {
			chartInstance.destroy();
			chartInstance = null;
		}
	});
</script>

<div class="flex flex-col gap-4 min-w-0">
	<!-- Header Bar -->
	<div class="flex items-center justify-between gap-3 min-h-7">
		<div class="min-w-0">
			<h3 class="text-sm font-semibold text-foreground tracking-tight truncate">
				Status Codes ({periodLabel})
			</h3>
		</div>

		<!-- View Switcher -->
		<div class="flex items-center gap-2 shrink-0">
			<!-- Toggle View Mode: % or Count -->
			<div class="bg-muted/60 rounded-lg p-0.5 flex items-center text-[11px] font-mono">
				<button
					type="button"
					class="px-2.5 py-1 rounded-md transition-all {viewMode === 'percentage'
						? 'bg-primary text-primary-foreground font-bold shadow-sm'
						: 'text-muted-foreground hover:text-foreground'}"
					onclick={() => (viewMode = 'percentage')}
				>
					% Ratio
				</button>
				<button
					type="button"
					class="px-2.5 py-1 rounded-md transition-all {viewMode === 'count'
						? 'bg-primary text-primary-foreground font-bold shadow-sm'
						: 'text-muted-foreground hover:text-foreground'}"
					onclick={() => (viewMode = 'count')}
				>
					Count
				</button>
			</div>
		</div>
	</div>

	{#if effectiveTotal === 0}
		<div
			class="h-[240px] flex items-center justify-center text-center text-muted-foreground text-xs font-mono"
		>
			No check data available for the selected timeframe
		</div>
	{:else}
		<!-- Chart.js Canvas (drawn directly on the page surface) -->
		<div class="h-[240px] w-full relative">
			<canvas bind:this={canvasEl}></canvas>
		</div>
	{/if}
</div>
