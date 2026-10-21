<script lang="ts">
	import { onMount, untrack } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { MonitorAPI, type Monitor, type MonitorDailyStat } from '$lib/api/monitor';
	import { IncidentAPI, type Incident } from '$lib/api/incident';
	import MonitorDetailHeader from '$lib/components/monitor/MonitorDetailHeader.svelte';
	import MonitorStatsGrid from '$lib/components/monitor/MonitorStatsGrid.svelte';
	import MonitorLatencyChart, {
		type DataPoint
	} from '$lib/components/monitor/MonitorLatencyChart.svelte';
	import MonitorStatusCodeChart from '$lib/components/monitor/MonitorStatusCodeChart.svelte';
	import MonitorAlertRulesSection from '$lib/components/monitor/MonitorAlertRulesSection.svelte';
	import MonitorPeriodSelector, {
		type Period
	} from '$lib/components/monitor/MonitorPeriodSelector.svelte';
	import MonitorIncidentHistory from '$lib/components/monitor/MonitorIncidentHistory.svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import { DataSkeleton } from '$lib/components/common';
	import { AlertTriangle, ArrowLeft } from 'lucide-svelte';
	import { formatDateOnly, formatDateTime } from '$lib/utils/datetime';
	import { userTimezone } from '$lib/stores/timezone';
	import { toast } from 'svelte-sonner';
	import { parseApiError } from '$lib/utils/parseApiError';

	// ── Params ────────────────────────────────────────────────────────────────
	const orgId = $derived($page.params.id as string);
	const monitorId = $derived($page.params.monitorId as string);

	// ── State ─────────────────────────────────────────────────────────────────
	let monitor = $state<Monitor | null>(null);
	let stats = $state<MonitorDailyStat[]>([]);
	let incidents = $state<Incident[]>([]);
	let activePeriod = $state<Period>('30d');
	let isLoading = $state(true);
	let fetchError = $state('');

	let tz = $derived($userTimezone);

	// Days count for active period
	let periodDays = $derived.by(() => {
		switch (activePeriod) {
			case '24h':
				return 1;
			case '7d':
				return 7;
			case '30d':
				return 30;
			case '90d':
				return 90;
			case '1y':
				return 365;
		}
	});

	// Effective days to render: avoid rendering empty void before monitor was created
	let effectiveDays = $derived.by(() => {
		if (activePeriod === '24h') return 1;
		if (!monitor?.CreatedAt) return periodDays;
		const createdDate = new Date(monitor.CreatedAt);
		const now = new Date();
		const diffMs = now.getTime() - createdDate.getTime();
		const diffDays = Math.ceil(diffMs / (1000 * 60 * 60 * 24));
		return Math.min(periodDays, Math.max(diffDays, 1));
	});

	// Calculated Metrics from stats / recent checks
	let computedMetrics = $derived.by(() => {
		if (activePeriod === '24h' && monitor?.recent_checks && monitor.recent_checks.length > 0) {
			const checks = monitor.recent_checks;
			const upCount = checks.filter((c) => c.Status === 'up').length;
			const latencies = checks.map((c) => c.ResponseTimeMs);
			const sumLatency = latencies.reduce((a, b) => a + b, 0);
			return {
				uptimePercent: Math.round((upCount / checks.length) * 10000) / 100,
				avgLatencyMs: Math.round(sumLatency / checks.length),
				minLatencyMs: latencies.length > 0 ? Math.min(...latencies) : 0,
				maxLatencyMs: latencies.length > 0 ? Math.max(...latencies) : 0,
				totalIncidents: incidents.length,
				downtimeFormatted: `${incidents.length * 5}m`
			};
		}

		if (stats.length === 0) {
			if (monitor?.recent_checks && monitor.recent_checks.length > 0) {
				const checks = monitor.recent_checks;
				const upCount = checks.filter((c) => c.Status === 'up').length;
				const latencies = checks.map((c) => c.ResponseTimeMs);
				const sumLatency = latencies.reduce((a, b) => a + b, 0);
				return {
					uptimePercent: Math.round((upCount / checks.length) * 10000) / 100,
					avgLatencyMs: Math.round(sumLatency / checks.length),
					minLatencyMs: latencies.length > 0 ? Math.min(...latencies) : 0,
					maxLatencyMs: latencies.length > 0 ? Math.max(...latencies) : 0,
					totalIncidents: incidents.length,
					downtimeFormatted: `${incidents.length * 5}m`
				};
			}
			return {
				uptimePercent: monitor?.UptimePercent ?? 100,
				avgLatencyMs: monitor?.ResponseTimeMs ?? 0,
				minLatencyMs: 0,
				maxLatencyMs: monitor?.ResponseTimeMs ?? 0,
				totalIncidents: incidents.length,
				downtimeFormatted: '0m'
			};
		}

		const totalChecks = stats.reduce((acc, s) => acc + s.TotalChecks, 0);
		const upChecks = stats.reduce((acc, s) => acc + s.UpChecks, 0);
		const uptimePercent = totalChecks > 0 ? (upChecks / totalChecks) * 100 : 100;
		const totalLatencyWeighted = stats.reduce((acc, s) => acc + s.AvgLatencyMs * s.TotalChecks, 0);
		const avgLatency = totalChecks > 0 ? Math.round(totalLatencyWeighted / totalChecks) : 0;
		const validMin = stats.map((s) => s.MinLatencyMs).filter((v) => v > 0);
		const minLatency = validMin.length > 0 ? Math.min(...validMin) : 0;
		const maxLatency = Math.max(...stats.map((s) => s.MaxLatencyMs), 0);

		return {
			uptimePercent: Number.isNaN(uptimePercent) ? 100 : uptimePercent,
			avgLatencyMs: Number.isNaN(avgLatency) ? 0 : avgLatency,
			minLatencyMs: Number.isFinite(minLatency) ? minLatency : 0,
			maxLatencyMs: Number.isFinite(maxLatency) ? maxLatency : 0,
			totalIncidents: incidents.length,
			downtimeFormatted: `${incidents.length * 5}m`
		};
	});

	// Chart Data Points
	let chartPoints = $derived.by<DataPoint[]>(() => {
		if (activePeriod === '24h' && monitor?.recent_checks && monitor.recent_checks.length > 0) {
			return [...monitor.recent_checks].reverse().map((c: any) => ({
				label: formatDateTime(c.CheckedAt, tz),
				value: c.ResponseTimeMs,
				status: c.Status,
				statusCode: c.StatusCode
			}));
		}

		if (stats.length === 0 && monitor?.recent_checks && monitor.recent_checks.length > 0) {
			return [...monitor.recent_checks].reverse().map((c: any) => ({
				label: formatDateTime(c.CheckedAt, tz),
				value: c.ResponseTimeMs,
				status: c.Status,
				statusCode: c.StatusCode
			}));
		}

		// Build a lookup map of existing daily stats by formatted date
		const statMap = new Map<string, MonitorDailyStat>();
		for (const s of stats) {
			const label = formatDateOnly(s.Date, tz);
			statMap.set(label, s);
		}

		// Generate contiguous timeline of effectiveDays (trimmed to monitor creation date)
		const points: DataPoint[] = [];
		const baseDate = new Date();
		baseDate.setHours(12, 0, 0, 0); // Midday avoids DST shifting

		for (let i = effectiveDays - 1; i >= 0; i--) {
			const d = new Date(baseDate);
			d.setDate(baseDate.getDate() - i);
			const label = formatDateOnly(d.toISOString(), tz);
			const s = statMap.get(label);

			if (s && s.TotalChecks > 0) {
				points.push({
					label,
					value: s.AvgLatencyMs,
					status: (s.DownChecks > 0 ? 'down' : 'up') as 'up' | 'down'
				});
			} else {
				points.push({
					label,
					value: 0,
					status: 'no_data'
				});
			}
		}

		return points;
	});

	// Status Code Aggregate Metrics
	let statusCodeMetrics = $derived.by(() => {
		if (activePeriod === '24h' && monitor?.recent_checks && monitor.recent_checks.length > 0) {
			const checks = monitor.recent_checks;
			let c2xx = 0,
				c3xx = 0,
				c4xx = 0,
				c5xx = 0,
				cErr = 0;
			for (const c of checks) {
				const sc = c.StatusCode;
				if (sc >= 200 && sc < 300) c2xx++;
				else if (sc >= 300 && sc < 400) c3xx++;
				else if (sc >= 400 && sc < 500) c4xx++;
				else if (sc >= 500 && sc < 600) c5xx++;
				else cErr++;
			}
			return {
				count2xx: c2xx,
				count3xx: c3xx,
				count4xx: c4xx,
				count5xx: c5xx,
				countError: cErr,
				total: checks.length
			};
		}

		if (stats.length > 0) {
			let c2xx = 0,
				c3xx = 0,
				c4xx = 0,
				c5xx = 0,
				cErr = 0,
				tot = 0;
			for (const s of stats) {
				c2xx += s.Count2xx ?? 0;
				c3xx += s.Count3xx ?? 0;
				c4xx += s.Count4xx ?? 0;
				c5xx += s.Count5xx ?? 0;
				cErr += s.CountError ?? s.DownChecks ?? 0;
				tot += s.TotalChecks ?? 0;
			}
			if (c2xx === 0 && c3xx === 0 && c4xx === 0 && c5xx === 0 && tot > 0) {
				const up = stats.reduce((acc, s) => acc + (s.UpChecks ?? 0), 0);
				const down = stats.reduce((acc, s) => acc + (s.DownChecks ?? 0), 0);
				c2xx = up;
				cErr = down;
			}
			return {
				count2xx: c2xx,
				count3xx: c3xx,
				count4xx: c4xx,
				count5xx: c5xx,
				countError: cErr,
				total: tot
			};
		}

		if (monitor?.recent_checks && monitor.recent_checks.length > 0) {
			const checks = monitor.recent_checks;
			let c2xx = 0,
				c3xx = 0,
				c4xx = 0,
				c5xx = 0,
				cErr = 0;
			for (const c of checks) {
				const sc = c.StatusCode;
				if (sc >= 200 && sc < 300) c2xx++;
				else if (sc >= 300 && sc < 400) c3xx++;
				else if (sc >= 400 && sc < 500) c4xx++;
				else if (sc >= 500 && sc < 600) c5xx++;
				else cErr++;
			}
			return {
				count2xx: c2xx,
				count3xx: c3xx,
				count4xx: c4xx,
				count5xx: c5xx,
				countError: cErr,
				total: checks.length
			};
		}

		return { count2xx: 0, count3xx: 0, count4xx: 0, count5xx: 0, countError: 0, total: 0 };
	});

	// ── Data Fetching ─────────────────────────────────────────────────────────
	let currentReqId = 0;

	async function loadPeriodStats(oid: string, mid: string, days: number) {
		const reqId = ++currentReqId;
		try {
			const statsRes = await MonitorAPI.getDailyStats(oid, mid, days);
			if (reqId === currentReqId) {
				stats = statsRes.data || [];
			}
		} catch (err) {
			if (reqId === currentReqId) {
				toast.error(parseApiError(err, 'Failed to load daily stats'));
				stats = [];
			}
		}
	}

	async function loadMonitorData(silent = false) {
		if (!silent) isLoading = true;
		fetchError = '';

		try {
			if (!orgId || !monitorId) return;
			const res = await MonitorAPI.getById(orgId, monitorId);
			if (res.data) {
				monitor = res.data;
				await loadPeriodStats(orgId, monitorId, periodDays);

				// Fetch incidents for this monitor
				try {
					const incRes = await IncidentAPI.getAll(orgId, monitorId);
					incidents = incRes.data || [];
				} catch (err) {
					toast.error(parseApiError(err, 'Failed to load incidents'));
					incidents = [];
				}
			} else {
				fetchError = 'Monitor not found or access denied';
			}
		} catch (err: any) {
			fetchError = parseApiError(err, 'Failed to load monitor data');
		} finally {
			isLoading = false;
		}
	}

	onMount(() => {
		loadMonitorData();
	});

	$effect(() => {
		const days = periodDays;
		if (orgId && monitorId) {
			untrack(() => {
				loadPeriodStats(orgId, monitorId, days);
			});
		}
	});

	let isTriggeringCheck = $state(false);

	async function handleTriggerCheck() {
		if (!monitor || isTriggeringCheck) return;
		isTriggeringCheck = true;
		try {
			const res = await MonitorAPI.triggerCheck(orgId, monitor.ID);
			const checkResult = res?.data;
			if (checkResult) {
				const statusText = checkResult.status === 'up' ? 'UP' : 'DOWN';
				const latency = checkResult.response_time_ms ?? checkResult.ResponseTimeMs ?? 0;
				toast.success(`Check completed: ${statusText} (${latency}ms)`);
			} else {
				toast.success('Check completed successfully');
			}
			await loadMonitorData(true);
		} catch (err: any) {
			toast.error(parseApiError(err, 'Failed to trigger check'));
		} finally {
			isTriggeringCheck = false;
		}
	}

	async function handlePauseResume() {
		if (!monitor) return;
		try {
			const willResume = monitor.Status === 'paused';
			await MonitorAPI.update(orgId, monitor.ID, {
				is_active: willResume
			});
			toast.success(willResume ? 'Monitor resumed' : 'Monitor paused');
			loadMonitorData(true);
		} catch (err: any) {
			toast.error(parseApiError(err, 'Failed to toggle monitor status'));
		}
	}

	async function handleDelete() {
		if (!monitor) return;
		if (!confirm(`Are you sure you want to delete monitor "${monitor.Name}"?`)) return;
		try {
			await MonitorAPI.delete(orgId, monitor.ID);
			toast.success('Monitor deleted successfully');
			goto(`/dashboard/org/${orgId}/monitors`);
		} catch (err: any) {
			toast.error(parseApiError(err, 'Failed to delete monitor'));
		}
	}
</script>

<svelte:head>
	<title>{monitor ? `${monitor.Name} • Analytics` : 'Monitor Analytics'} | Netra Uptime</title>
</svelte:head>

<div class="space-y-6 max-w-7xl mx-auto text-foreground antialiased">
	<!-- Navigation breadcrumb / back button -->
	<div>
		<a
			href="/dashboard/org/{orgId}/monitors"
			class="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground transition-colors group"
		>
			<ArrowLeft class="size-4 group-hover:-translate-x-1 transition-transform" />
			<span>Back to Monitors</span>
		</a>
	</div>

	{#if isLoading}
		<DataSkeleton variant="detail" />
	{:else if fetchError || !monitor}
		<div
			class="border border-border rounded-lg bg-card p-8 text-center space-y-3 max-w-full mx-auto"
		>
			<AlertTriangle class="size-10 text-primary mx-auto" />
			<h2 class="text-lg font-bold text-foreground">Failed to Load Monitor</h2>
			<p class="text-xs text-muted-foreground">{fetchError || 'Monitor not found'}</p>
			<Button
				href="/dashboard/org/{orgId}/monitors"
				class="bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground active:scale-[0.98] transition-all duration-200 h-9 px-4 gap-1.5 font-medium"
			>
				Back to Monitors
			</Button>
		</div>
	{:else}
		<!-- 1. Header & Actions Banner -->
		<MonitorDetailHeader
			{monitor}
			onPauseResume={handlePauseResume}
			onDelete={handleDelete}
			onCheckNow={handleTriggerCheck}
			isChecking={isTriggeringCheck}
		/>

		<!-- 2. Time Range Selector Control Bar -->
		<MonitorPeriodSelector
			{activePeriod}
			{effectiveDays}
			onSelectPeriod={(p) => (activePeriod = p)}
		/>

		<!-- 3. Key Metrics Summary Grid Cards -->
		<MonitorStatsGrid
			uptimePercent={computedMetrics.uptimePercent}
			avgLatencyMs={computedMetrics.avgLatencyMs}
			minLatencyMs={computedMetrics.minLatencyMs}
			maxLatencyMs={computedMetrics.maxLatencyMs}
			totalIncidents={computedMetrics.totalIncidents}
			downtimeFormatted={computedMetrics.downtimeFormatted}
		/>

		<!-- 4. Interactive Response Time Latency Line Chart -->
		<MonitorLatencyChart
			data={chartPoints}
			title={`Response Time Latency (${activePeriod.toUpperCase()})`}
			height={240}
			monitorType={monitor.Type}
		/>

		<!-- 5. HTTP Status Code Distribution Chart -->
		{#if monitor.Type === 'http'}
			<MonitorStatusCodeChart
				count2xx={statusCodeMetrics.count2xx}
				count3xx={statusCodeMetrics.count3xx}
				count4xx={statusCodeMetrics.count4xx}
				count5xx={statusCodeMetrics.count5xx}
				countError={statusCodeMetrics.countError}
				totalChecks={statusCodeMetrics.total}
				periodLabel={activePeriod.toUpperCase()}
			/>
		{/if}

		<!-- 6. Alert Rules & Incident Policies -->
		<MonitorAlertRulesSection
			{orgId}
			{monitorId}
			monitorType={monitor.Type}
			alertTemplateId={monitor.AlertTemplateID}
		/>

		<!-- 7. Recent Incident Logs History -->
		<MonitorIncidentHistory {orgId} {incidents} {tz} />
	{/if}
</div>
