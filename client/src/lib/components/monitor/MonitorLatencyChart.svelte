<script lang="ts">
	import { onDestroy } from 'svelte';
	import Chart from 'chart.js/auto';

	export interface DataPoint {
		label: string;
		value: number | null; // response time in ms (null if no checks on this date)
		status?: 'up' | 'down' | 'degraded' | 'unresponsive' | 'no_data';
		statusCode?: number;
	}

	interface Props {
		data: DataPoint[];
		title?: string;
		height?: number;
		color?: string;
		monitorType?: string;
	}

	let {
		data = [],
		title = 'Response Time (ms)',
		height = 240,
		color = '#10B981',
		monitorType = 'http'
	}: Props = $props();

	let canvasEl = $state<HTMLCanvasElement | null>(null);
	let chartInstance: Chart<'line'> | null = null;

	function getPointColor(status?: string, statusCode?: number): string {
		if (status === 'no_data') return 'transparent';
		if (monitorType === 'http' && statusCode && statusCode > 0) {
			if (statusCode >= 200 && statusCode < 300) return '#10B981'; // Emerald 2xx
			if (statusCode >= 300 && statusCode < 400) return '#3B82F6'; // Blue 3xx
			if (statusCode >= 400 && statusCode < 500) return '#F59E0B'; // Amber 4xx
			if (statusCode >= 500 && statusCode < 600) return '#EF4444'; // Rose 5xx
		}
		if (status === 'down') return '#EF4444';
		if (status === 'degraded' || status === 'unresponsive') return '#F59E0B';
		return '#10B981';
	}

	function createOrUpdateChart() {
		if (!canvasEl || data.length === 0) return;

		const labels = data.map((d) => d.label);
		const values = data.map((d) => d.value);
		const pointColors = data.map((d) => getPointColor(d.status, d.statusCode));

		const ctx = canvasEl.getContext('2d');
		let gradient: any = 'rgba(16, 185, 129, 0.15)';
		if (ctx) {
			const g = ctx.createLinearGradient(0, 0, 0, height);
			g.addColorStop(0, `${color}40`);
			g.addColorStop(1, `${color}00`);
			gradient = g;
		}

		const isNoDataPoint = (idx: number) => {
			const pt = data[idx];
			return !pt || pt.status === 'no_data' || pt.value === 0;
		};

		const isGapSegment = (c: any) => {
			return (
				c.p0?.skip ||
				c.p1?.skip ||
				isNoDataPoint(c.p0DataIndex) ||
				isNoDataPoint(c.p1DataIndex) ||
				Math.abs(c.p1DataIndex - c.p0DataIndex) > 1
			);
		};

		if (!chartInstance) {
			chartInstance = new Chart(canvasEl, {
				type: 'line',
				data: {
					labels,
					datasets: [
						{
							data: values as any,
							borderColor: color,
							borderWidth: 2,
							tension: 0.3,
							fill: true,
							spanGaps: true,
							segment: {
								borderColor: (c: any) => (isGapSegment(c) ? 'rgba(148, 163, 184, 0.45)' : color),
								borderDash: (c: any) => (isGapSegment(c) ? [5, 5] : undefined)
							},
							backgroundColor: gradient,
							pointBackgroundColor: pointColors,
							pointBorderColor: (c: any) => {
								const idx = c.dataIndex;
								const pt = data[idx];
								return !pt || pt.status === 'no_data' || pt.value === 0 ? 'transparent' : '#0f172a';
							},
							pointBorderWidth: 1.5,
							pointRadius: (c: any) => {
								const idx = c.dataIndex;
								const pt = data[idx];
								if (!pt || pt.status === 'no_data' || pt.value === 0) return 0;
								return data.length > 50 ? 2 : 3.5;
							},
							pointHoverRadius: (c: any) => {
								const idx = c.dataIndex;
								const pt = data[idx];
								if (!pt || pt.status === 'no_data' || pt.value === 0) return 0;
								return 6;
							},
							pointHoverBackgroundColor: pointColors,
							pointHoverBorderColor: (c: any) => {
								const idx = c.dataIndex;
								const pt = data[idx];
								return !pt || pt.status === 'no_data' || pt.value === 0 ? 'transparent' : '#ffffff';
							},
							pointHoverBorderWidth: 2
						}
					]
				},
				options: {
					responsive: true,
					maintainAspectRatio: false,
					interaction: {
						mode: 'index',
						intersect: false
					},
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
									return data[idx]?.label ?? '';
								},
								label: (item) => {
									const idx = item.dataIndex;
									const pt = data[idx];
									if (!pt || pt.status === 'no_data' || pt.value === 0) {
										return '0 ms (No Data Recorded)';
									}
									return `Latency: ${pt.value} ms`;
								},
								afterLabel: (item) => {
									const idx = item.dataIndex;
									const pt = data[idx];
									if (!pt || pt.status === 'no_data' || pt.value === 0) {
										return 'Status: Monitor Offline / No checks';
									}
									if (pt.statusCode && pt.statusCode > 0) {
										return `HTTP Status: ${pt.statusCode}`;
									}
									if (pt.status === 'down') {
										return `Status: Down (Connection Error)`;
									}
									if (pt.status === 'degraded' || pt.status === 'unresponsive') {
										return `Status: Unresponsive / Degraded`;
									}
									return `Status: OK`;
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
								color: '#64748b',
								font: {
									family: 'monospace',
									size: 10
								},
								maxTicksLimit: 10,
								maxRotation: 0
							}
						},
						y: {
							min: 0,
							grid: {
								color: 'rgba(255, 255, 255, 0.05)'
							},
							ticks: {
								color: '#64748b',
								font: {
									family: 'monospace',
									size: 10
								},
								callback: (val) => `${val} ms`
							}
						}
					}
				}
			});
		} else {
			chartInstance.data.labels = labels;
			chartInstance.data.datasets[0].data = values as any;
			chartInstance.data.datasets[0].pointBackgroundColor = pointColors;
			chartInstance.data.datasets[0].pointHoverBackgroundColor = pointColors;
			chartInstance.data.datasets[0].spanGaps = true;
			chartInstance.data.datasets[0].segment = {
				borderColor: (c: any) => (isGapSegment(c) ? 'rgba(148, 163, 184, 0.45)' : color),
				borderDash: (c: any) => (isGapSegment(c) ? [5, 5] : undefined)
			} as any;
			chartInstance.data.datasets[0].pointRadius = ((c: any) => {
				const idx = c.dataIndex;
				const pt = data[idx];
				if (!pt || pt.status === 'no_data' || pt.value === 0) return 0;
				return data.length > 50 ? 2 : 3.5;
			}) as any;
			chartInstance.data.datasets[0].pointHoverRadius = ((c: any) => {
				const idx = c.dataIndex;
				const pt = data[idx];
				if (!pt || pt.status === 'no_data' || pt.value === 0) return 0;
				return 6;
			}) as any;
			chartInstance.update();
		}
	}

	$effect(() => {
		if (canvasEl && data) {
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
	<div class="flex items-center justify-between min-h-7 flex-wrap gap-2">
		<div class="flex items-center gap-3">
			<h3 class="text-sm font-semibold text-foreground tracking-tight flex items-center gap-2">
				<span>{title}</span>
			</h3>

			<!-- Status Code Color Legend (Only for HTTP) or Status Legend (TCP / Ping) -->
			{#if monitorType === 'http'}
				<div
					class="hidden md:flex items-center gap-2.5 text-[11px] font-mono text-muted-foreground pl-2 border-l border-border"
				>
					<span class="flex items-center gap-1"
						><span class="size-2 rounded-full bg-emerald-400"></span> 2xx</span
					>
					<span class="flex items-center gap-1"
						><span class="size-2 rounded-full bg-blue-400"></span> 3xx</span
					>
					<span class="flex items-center gap-1"
						><span class="size-2 rounded-full bg-amber-400"></span> 4xx</span
					>
					<span class="flex items-center gap-1"
						><span class="size-2 rounded-full bg-rose-500"></span> 5xx/Err</span
					>
					<span class="flex items-center gap-1.5 pl-1.5 border-l border-border/80"
						><span class="w-3 border-t border-dashed border-slate-400"></span> No Data</span
					>
				</div>
			{:else}
				<div
					class="hidden md:flex items-center gap-2.5 text-[11px] font-mono text-muted-foreground pl-2 border-l border-border"
				>
					<span class="flex items-center gap-1"
						><span class="size-2 rounded-full bg-emerald-400"></span> Operational</span
					>
					<span class="flex items-center gap-1"
						><span class="size-2 rounded-full bg-amber-400"></span> Degraded</span
					>
					<span class="flex items-center gap-1"
						><span class="size-2 rounded-full bg-rose-500"></span> Down / Timeout</span
					>
					<span class="flex items-center gap-1.5 pl-1.5 border-l border-border/80"
						><span class="w-3 border-t border-dashed border-slate-400"></span> No Data</span
					>
				</div>
			{/if}
		</div>
	</div>

	{#if data.length === 0}
		<div class="h-48 flex items-center justify-center text-muted-foreground text-xs font-mono">
			No latency data available for the selected period
		</div>
	{:else}
		<!-- Chart.js Canvas Container -->
		<div class="w-full bg-white/2 border border-border/60 rounded-xl p-3 sm:p-4">
			<div class="w-full relative" style="height: {height}px;">
				<canvas bind:this={canvasEl}></canvas>
			</div>
		</div>
	{/if}
</div>
