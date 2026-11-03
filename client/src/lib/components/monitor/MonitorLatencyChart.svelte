<script lang="ts">
	import { onDestroy } from 'svelte';
	import Chart from 'chart.js/auto';

	export interface DataPoint {
		label: string;
		value: number | null; // response time in ms (null if no checks on this date)
		status?: 'up' | 'down' | 'degraded' | 'unresponsive' | 'no_data';
		statusCode?: number;
		details?: {
			count2xx?: number;
			count3xx?: number;
			count4xx?: number;
			count5xx?: number;
			countError?: number;
			total?: number;
			uptimePercent?: number;
		};
	}

	interface Props {
		data: DataPoint[];
		title?: string;
		height?: number;
		color?: string;
		monitorType?: string;
		isAggregated?: boolean;
	}

	let {
		data = [],
		title = 'Response Time (ms)',
		height = 240,
		color = '#10B981',
		monitorType = 'http',
		isAggregated = false
	}: Props = $props();

	let canvasEl = $state<HTMLCanvasElement | null>(null);
	let chartInstance: Chart<'line'> | null = null;

	function getPointColor(pt: DataPoint): string {
		if (pt.status === 'no_data') return 'transparent';
		if (monitorType === 'http') {
			// 1. Single check realtime mode (24h)
			if (pt.statusCode && pt.statusCode > 0) {
				if (pt.statusCode >= 200 && pt.statusCode < 300) return '#10B981'; // Emerald 2xx
				if (pt.statusCode >= 300 && pt.statusCode < 400) return '#3B82F6'; // Blue 3xx
				if (pt.statusCode >= 400 && pt.statusCode < 500) return '#F59E0B'; // Amber 4xx
				if (pt.statusCode >= 500 && pt.statusCode < 600) return '#EF4444'; // Rose 5xx
			}
			// 2. Daily aggregate mode (> 24h) with status breakdown
			if (pt.details) {
				if (
					(pt.details.count5xx ?? 0) > 0 ||
					(pt.details.countError ?? 0) > 0 ||
					pt.status === 'down'
				) {
					return '#EF4444'; // Rose 5xx / Outage
				}
				if ((pt.details.count4xx ?? 0) > 0) {
					return '#F59E0B'; // Amber 4xx / Client Errors
				}
				if ((pt.details.count3xx ?? 0) > 0) {
					return '#3B82F6'; // Blue 3xx / Redirects
				}
				return '#10B981'; // Emerald 2xx / Healthy
			}
		}
		if (pt.status === 'down') return '#EF4444';
		if (pt.status === 'degraded' || pt.status === 'unresponsive') return '#F59E0B';
		return '#10B981';
	}

	function createOrUpdateChart() {
		if (!canvasEl || data.length === 0) return;

		const labels = data.map((d) => d.label);
		const values = data.map((d) => d.value);
		const pointColors = data.map((d) => getPointColor(d));
		const isDark =
			typeof document !== 'undefined' && document.documentElement.classList.contains('dark');
		const gridColor = isDark ? 'rgba(255, 255, 255, 0.06)' : 'rgba(15, 23, 42, 0.06)';
		const pointBorder = isDark ? '#0f172a' : '#ffffff';

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
			return !pt || pt.status === 'no_data' || pt.value === null;
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
							pointBorderColor: (c: any) =>
								isNoDataPoint(c.dataIndex) ? 'transparent' : pointBorder,
							pointBorderWidth: 1.5,
							pointRadius: (c: any) => {
								if (isNoDataPoint(c.dataIndex)) return 0;
								return data.length > 50 ? 2 : 3.5;
							},
							pointHoverRadius: (c: any) => (isNoDataPoint(c.dataIndex) ? 0 : 6),
							pointHoverBackgroundColor: pointColors,
							pointHoverBorderColor: (c: any) =>
								isNoDataPoint(c.dataIndex) ? 'transparent' : '#ffffff',
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
									if (isNoDataPoint(idx)) {
										return '0 ms (No Data Recorded)';
									}
									return `Latency: ${pt.value} ms`;
								},
								afterLabel: (item) => {
									const idx = item.dataIndex;
									const pt = data[idx];
									if (isNoDataPoint(idx)) {
										return 'Status: Monitor Offline / No checks';
									}
									if (pt.details) {
										const lines: string[] = [];
										if (pt.details.uptimePercent !== undefined) {
											const val = Number(pt.details.uptimePercent);
											if (!Number.isNaN(val)) {
												const formatted = val === 100 ? '100%' : `${val.toFixed(2)}%`;
												lines.push(`Uptime: ${formatted}`);
											}
										}
										const counts: string[] = [];
										if (pt.details.count2xx) counts.push(`2xx: ${pt.details.count2xx}`);
										if (pt.details.count3xx) counts.push(`3xx: ${pt.details.count3xx}`);
										if (pt.details.count4xx) counts.push(`4xx: ${pt.details.count4xx}`);
										if (pt.details.count5xx) counts.push(`5xx: ${pt.details.count5xx}`);
										if (pt.details.countError) counts.push(`Err: ${pt.details.countError}`);
										if (counts.length > 0) {
											lines.push(`Checks: ${counts.join(' | ')}`);
										} else if (pt.details.total) {
											lines.push(`Checks: ${pt.details.total} total`);
										}
										return lines.length > 0 ? (lines as any) : 'Status: OK';
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
								color: gridColor
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
			chartInstance.data.datasets[0].pointBorderColor = ((c: any) =>
				isNoDataPoint(c.dataIndex) ? 'transparent' : pointBorder) as any;
			chartInstance.data.datasets[0].pointHoverBorderColor = ((c: any) =>
				isNoDataPoint(c.dataIndex) ? 'transparent' : '#ffffff') as any;
			chartInstance.data.datasets[0].spanGaps = true;
			chartInstance.data.datasets[0].segment = {
				borderColor: (c: any) => (isGapSegment(c) ? 'rgba(148, 163, 184, 0.45)' : color),
				borderDash: (c: any) => (isGapSegment(c) ? [5, 5] : undefined)
			} as any;
			chartInstance.data.datasets[0].pointRadius = ((c: any) => {
				if (isNoDataPoint(c.dataIndex)) return 0;
				return data.length > 50 ? 2 : 3.5;
			}) as any;
			chartInstance.data.datasets[0].pointHoverRadius = ((c: any) =>
				isNoDataPoint(c.dataIndex) ? 0 : 6) as any;
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

<div class="flex flex-col gap-4 min-w-0">
	<!-- Header Bar -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 min-h-7">
		<h3 class="text-sm font-semibold text-foreground tracking-tight">
			{title}
		</h3>

		<!-- Status Code Color Legend (Only for HTTP) or Status Legend (TCP / Ping) -->
		{#if monitorType === 'http'}
			<div
				class="flex items-center gap-x-3 gap-y-1 flex-wrap text-[11px] font-mono text-muted-foreground"
			>
				<span class="flex items-center gap-1"
					><span class="size-2 rounded-full bg-emerald-400"></span>
					{isAggregated ? 'All 2xx' : '2xx'}</span
				>
				<span class="flex items-center gap-1"
					><span class="size-2 rounded-full bg-blue-400"></span>
					{isAggregated ? 'Has 3xx' : '3xx'}</span
				>
				<span class="flex items-center gap-1"
					><span class="size-2 rounded-full bg-amber-400"></span>
					{isAggregated ? 'Has 4xx' : '4xx'}</span
				>
				<span class="flex items-center gap-1"
					><span class="size-2 rounded-full bg-rose-500"></span>
					{isAggregated ? '5xx / Outage' : '5xx/Err'}</span
				>
				<span class="flex items-center gap-1.5"
					><span class="w-3 border-t border-dashed border-slate-400"></span> No Data</span
				>
			</div>
		{:else}
			<div
				class="flex items-center gap-x-3 gap-y-1 flex-wrap text-[11px] font-mono text-muted-foreground"
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
				<span class="flex items-center gap-1.5"
					><span class="w-3 border-t border-dashed border-slate-400"></span> No Data</span
				>
			</div>
		{/if}
	</div>

	{#if data.length === 0}
		<div
			class="flex items-center justify-center text-muted-foreground text-xs font-mono"
			style="height: {height}px;"
		>
			No latency data available for the selected period
		</div>
	{:else}
		<!-- Chart.js Canvas (drawn directly on the page surface) -->
		<div class="w-full relative" style="height: {height}px;">
			<canvas bind:this={canvasEl}></canvas>
		</div>
	{/if}
</div>
