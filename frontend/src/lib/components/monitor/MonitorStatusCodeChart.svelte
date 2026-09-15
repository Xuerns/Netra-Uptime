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
			subtext: 'OK / Created',
			desc: 'Status 200–299: Successful responses',
			color: '#10B981',
			hoverColor: '#34D399'
		},
		{
			id: '3xx',
			label: '3xx Redirect',
			subtext: 'Moved / Found',
			desc: 'Status 300–399: Redirections',
			color: '#3B82F6',
			hoverColor: '#60A5FA'
		},
		{
			id: '4xx',
			label: '4xx Client Err',
			subtext: 'Bad Req / 404',
			desc: 'Status 400–499: Client errors / Not found',
			color: '#F59E0B',
			hoverColor: '#FBBF24'
		},
		{
			id: '5xx',
			label: '5xx Server Err',
			subtext: '500 / 502 / 503',
			desc: 'Status 500–599: Internal server errors',
			color: '#EF4444',
			hoverColor: '#F87171'
		},
		{
			id: 'error',
			label: 'Timeout / Err',
			subtext: 'No Response',
			desc: 'Connection timeouts or dropped requests',
			color: '#A855F7',
			hoverColor: '#C084FC'
		}
	];

	function createOrUpdateChart() {
		if (!canvasEl) return;

		const currentData = viewMode === 'percentage' ? percentages : rawCounts;

		if (!chartInstance) {
			chartInstance = new Chart(canvasEl, {
				type: 'bar',
				data: {
					labels: categories.map((c) => c.label),
					datasets: [
						{
							data: currentData,
							backgroundColor: categories.map((c) => c.color),
							hoverBackgroundColor: categories.map((c) => c.hoverColor),
							borderRadius: 8,
							borderSkipped: 'bottom',
							maxBarThickness: 54
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
								color: '#94a3b8',
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
								color: 'rgba(255, 255, 255, 0.05)'
							},
							ticks: {
								color: '#64748b',
								font: {
									family: 'monospace',
									size: 11
								},
								callback: (val) => {
									return viewMode === 'percentage'
										? `${val}%`
										: Number(val).toLocaleString();
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

<div class="bg-card border border-border rounded-2xl p-5 md:p-6 shadow-xl flex flex-col space-y-4">
	<!-- Header Bar -->
	<div class="flex items-center justify-between flex-wrap gap-3">
		<div class="space-y-0.5">
			<h3 class="text-sm font-semibold text-foreground tracking-tight">
				HTTP Status Code Distribution
			</h3>
			<p class="text-xs text-muted-foreground">
				Visual breakdown of HTTP status responses and network errors for ({periodLabel}).
			</p>
		</div>

		<!-- View Switcher & Total Count -->
		<div class="flex items-center gap-2">
			<!-- Toggle View Mode: % or Count -->
			<div class="bg-white/5 border border-border rounded-lg p-0.5 flex items-center text-[11px] font-mono">
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
		<div class="h-48 flex items-center justify-center text-muted-foreground text-xs font-mono">
			No check data available for the selected timeframe
		</div>
	{:else}
		<!-- Chart.js Canvas Container (Clean, Zero Bleeding, Zero Ghost Backgrounds) -->
		<div class="w-full bg-white/2 border border-border/60 rounded-xl p-3 sm:p-4">
			<div class="h-[220px] sm:h-[240px] w-full relative">
				<canvas bind:this={canvasEl}></canvas>
			</div>
		</div>
	{/if}
</div>
