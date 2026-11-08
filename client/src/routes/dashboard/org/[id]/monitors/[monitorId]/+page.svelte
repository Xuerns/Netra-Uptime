<script lang="ts">
	import { onMount, untrack } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import {
		MonitorAPI,
		type Monitor,
		type MonitorDailyStat,
		type RecentCheck
	} from '$lib/api/monitor';
	import { IncidentAPI, type Incident } from '$lib/api/incident';
	import MonitorDetailHeader from '$lib/components/monitor/MonitorDetailHeader.svelte';
	import MonitorModal from '$lib/components/monitor/MonitorModal.svelte';
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
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	// ── Params ────────────────────────────────────────────────────────────────
	const orgId = $derived(data.orgId || ($page.params.id as string));
	const monitorId = $derived(($page.params.monitorId as string) || '');

	// ── State ─────────────────────────────────────────────────────────────────
	let monitor = $state<Monitor | null>(null);
	let stats = $state<MonitorDailyStat[]>([]);
	let incidents = $state<Incident[]>([]);
	let activePeriod = $state<Period>('30d');
	let isLoading = $state(true);
	let fetchError = $state('');
	let isEditModalOpen = $state(false);

	let tz = $derived($userTimezone);

	function isAvailableCheck(c: RecentCheck): boolean {
		if (c.Status === 'up') return true;
		if (c.Status === 'degraded') {
			if (c.StatusCode > 0 && c.StatusCode < 500) return true;
			if (c.StatusCode === 0 && c.ResponseTimeMs > 0 && c.ErrorMessage?.startsWith('Response time'))
				return true;
		}
		return false;
	}

	function calculateDowntimeFormatted(incidentList: Incident[], daysWindow: number): string {
		if (!incidentList || incidentList.length === 0) return '0m';
		const nowMs = Date.now();
		const windowStartMs = nowMs - daysWindow * 86_400_000;
		let totalMs = 0;

		for (const inc of incidentList) {
			const startMs = new Date(inc.StartedAt).getTime();
			if (Number.isNaN(startMs)) continue;
			const rawEndMs = inc.ResolvedAt ? new Date(inc.ResolvedAt).getTime() : nowMs;
			const endMs = Number.isNaN(rawEndMs) ? nowMs : rawEndMs;
			if (endMs < windowStartMs) continue;

			const effectiveStart = Math.max(startMs, windowStartMs);
			if (endMs > effectiveStart) {
				totalMs += endMs - effectiveStart;
			}
		}

		if (totalMs <= 0) return '0m';
		const totalSeconds = Math.floor(totalMs / 1000);
		if (totalSeconds < 60) return '< 1m';
		const totalMinutes = Math.floor(totalSeconds / 60);
		if (totalMinutes < 60) return `${totalMinutes}m`;
		const hours = Math.floor(totalMinutes / 60);
		const mins = totalMinutes % 60;
		if (hours < 24) {
			return mins > 0 ? `${hours}h ${mins}m` : `${hours}h`;
		}
		const days = Math.floor(hours / 24);
		const remHours = hours % 24;
		return remHours > 0 ? `${days}d ${remHours}h` : `${days}d`;
	}

	function computeRecentCheckMetrics(checks: RecentCheck[], downtimeFormatted: string) {
		const activeChecks = checks.filter((c) => c.Status !== 'maintenance');
		const upCount = activeChecks.filter(isAvailableCheck).length;
		const uptimePercent =
			activeChecks.length > 0 ? Math.round((upCount / activeChecks.length) * 10000) / 100 : 100;
		const healthyLatencies = activeChecks
			.filter((c) => isAvailableCheck(c) && c.ResponseTimeMs > 0)
			.map((c) => c.ResponseTimeMs);
		const sumLatency = healthyLatencies.reduce((a, b) => a + b, 0);

		return {
			uptimePercent,
			avgLatencyMs:
				healthyLatencies.length > 0 ? Math.round(sumLatency / healthyLatencies.length) : 0,
			minLatencyMs: healthyLatencies.length > 0 ? Math.min(...healthyLatencies) : 0,
			maxLatencyMs: healthyLatencies.length > 0 ? Math.max(...healthyLatencies) : 0,
			totalIncidents: incidents.length,
			downtimeFormatted
		};
	}

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
		const downtimeFormatted = calculateDowntimeFormatted(incidents, periodDays);

		if (activePeriod === '24h' && monitor?.recent_checks && monitor.recent_checks.length > 0) {
			return computeRecentCheckMetrics(monitor.recent_checks, downtimeFormatted);
		}

		if (stats.length === 0) {
			if (monitor?.recent_checks && monitor.recent_checks.length > 0) {
				return computeRecentCheckMetrics(monitor.recent_checks, downtimeFormatted);
			}
			return {
				uptimePercent: monitor?.UptimePercent ?? 100,
				avgLatencyMs: monitor?.ResponseTimeMs ?? 0,
				minLatencyMs: 0,
				maxLatencyMs: monitor?.ResponseTimeMs ?? 0,
				totalIncidents: incidents.length,
				downtimeFormatted
			};
		}

		const totalChecks = stats.reduce((acc, s) => acc + s.TotalChecks, 0);
		const upChecks = stats.reduce((acc, s) => acc + s.UpChecks, 0);
		const uptimePercent = totalChecks > 0 ? (upChecks / totalChecks) * 100 : 100;
		const totalLatencyWeighted = stats.reduce(
			(acc, s) => acc + s.AvgLatencyMs * (s.UpChecks > 0 ? s.UpChecks : 0),
			0
		);
		const avgLatency = upChecks > 0 ? Math.round(totalLatencyWeighted / upChecks) : 0;
		const validMin = stats.map((s) => s.MinLatencyMs).filter((v) => v > 0);
		const minLatency = validMin.length > 0 ? Math.min(...validMin) : 0;
		const maxLatency = Math.max(...stats.map((s) => s.MaxLatencyMs), 0);

		return {
			uptimePercent: Number.isNaN(uptimePercent) ? 100 : uptimePercent,
			avgLatencyMs: Number.isNaN(avgLatency) ? 0 : avgLatency,
			minLatencyMs: Number.isFinite(minLatency) ? minLatency : 0,
			maxLatencyMs: Number.isFinite(maxLatency) ? maxLatency : 0,
			totalIncidents: incidents.length,
			downtimeFormatted
		};
	});

	// Chart Data Points
	let chartPoints = $derived.by<DataPoint[]>(() => {
		if (activePeriod === '24h' && monitor?.recent_checks && monitor.recent_checks.length > 0) {
			return [...monitor.recent_checks].reverse().map((c) => ({
				label: formatDateTime(c.CheckedAt, tz),
				value: c.ResponseTimeMs,
				status: c.Status,
				statusCode: c.StatusCode,
				responseSnippet: c.ResponseSnippet,
				errorMessage: c.ErrorMessage
			}));
		}

		if (stats.length === 0 && monitor?.recent_checks && monitor.recent_checks.length > 0) {
			return [...monitor.recent_checks].reverse().map((c) => ({
				label: formatDateTime(c.CheckedAt, tz),
				value: c.ResponseTimeMs,
				status: c.Status,
				statusCode: c.StatusCode,
				responseSnippet: c.ResponseSnippet,
				errorMessage: c.ErrorMessage
			}));
		}

		// Build a lookup map of existing daily stats by canonical UTC date key (YYYY-MM-DD)
		const statMap = new Map<string, MonitorDailyStat>();
		for (const s of stats) {
			const dateKey = s.Date.slice(0, 10);
			statMap.set(dateKey, s);
		}

		// Generate contiguous timeline of effectiveDays aligned with UTC daily buckets
		const points: DataPoint[] = [];
		const now = new Date();
		const baseUtcMs = Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate(), 12, 0, 0);

		for (let i = effectiveDays - 1; i >= 0; i--) {
			const d = new Date(baseUtcMs - i * 86_400_000);
			const dateKey = d.toISOString().slice(0, 10);
			const label = formatDateOnly(d, 'UTC');
			const s = statMap.get(dateKey);

			if (s && s.TotalChecks > 0) {
				const uptimePct =
					typeof s.UptimePercentage === 'number'
						? s.UptimePercentage
						: (s.UpChecks / s.TotalChecks) * 100;
				let calculatedStatus: 'up' | 'down' | 'degraded' = 'up';
				if (uptimePct < 95) {
					calculatedStatus = 'down';
				} else if (uptimePct < 99) {
					calculatedStatus = 'degraded';
				}

				points.push({
					label,
					value: s.AvgLatencyMs,
					status: calculatedStatus,
					details: {
						count2xx: s.Count2xx,
						count3xx: s.Count3xx,
						count4xx: s.Count4xx,
						count5xx: s.Count5xx,
						countError: s.CountError ?? s.DownChecks,
						total: s.TotalChecks,
						uptimePercent: uptimePct
					}
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
		} catch (err: unknown) {
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
		} catch (err: unknown) {
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
		} catch (err: unknown) {
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
		} catch (err: unknown) {
			toast.error(parseApiError(err, 'Failed to delete monitor'));
		}
	}
</script>

<svelte:head>
	<title>{monitor ? `${monitor.Name} • Analytics` : 'Monitor Analytics'} | Netra Uptime</title>
</svelte:head>

<div class="max-w-7xl mx-auto text-foreground antialiased">
	<!-- Navigation breadcrumb / back button -->
	<div class="mb-6">
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
		<!-- 1. Header & Actions -->
		<header class="pb-6 border-b border-border/60">
			<MonitorDetailHeader
				{monitor}
				onPauseResume={handlePauseResume}
				onEdit={() => (isEditModalOpen = true)}
				onDelete={handleDelete}
				onCheckNow={handleTriggerCheck}
				isChecking={isTriggeringCheck}
			/>
		</header>

		<MonitorModal
			bind:isOpen={isEditModalOpen}
			{monitor}
			{orgId}
			onSuccess={() => loadMonitorData(true)}
		/>

		<!-- 2. Overview: Time Range Selector + Key Metrics Strip -->
		<section class="py-6 border-b border-border/60 space-y-6" aria-labelledby="overview-heading">
			<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
				<h2 id="overview-heading" class="text-base font-semibold tracking-tight">Overview</h2>
				<MonitorPeriodSelector
					{activePeriod}
					{effectiveDays}
					onSelectPeriod={(p) => (activePeriod = p)}
				/>
			</div>

			<MonitorStatsGrid
				uptimePercent={computedMetrics.uptimePercent}
				avgLatencyMs={computedMetrics.avgLatencyMs}
				minLatencyMs={computedMetrics.minLatencyMs}
				maxLatencyMs={computedMetrics.maxLatencyMs}
				totalIncidents={computedMetrics.totalIncidents}
				downtimeFormatted={computedMetrics.downtimeFormatted}
			/>
		</section>

		<!-- 3. Charts: Latency (2/3) + HTTP Status Codes (1/3) -->
		<section
			class="grid grid-cols-1 border-b border-border/60 {monitor.Type === 'http'
				? 'lg:grid-cols-3'
				: ''}"
			aria-label="Performance charts"
		>
			<div class="py-6 min-w-0 {monitor.Type === 'http' ? 'lg:col-span-2 lg:pr-6' : ''}">
				<MonitorLatencyChart
					data={chartPoints}
					title={`Response Time (${activePeriod.toUpperCase()})`}
					height={240}
					monitorType={monitor.Type}
					isAggregated={activePeriod !== '24h'}
				/>
			</div>

			{#if monitor.Type === 'http'}
				<div class="py-6 min-w-0 border-t border-border/60 lg:border-t-0 lg:border-l lg:pl-6">
					<MonitorStatusCodeChart
						count2xx={statusCodeMetrics.count2xx}
						count3xx={statusCodeMetrics.count3xx}
						count4xx={statusCodeMetrics.count4xx}
						count5xx={statusCodeMetrics.count5xx}
						countError={statusCodeMetrics.countError}
						totalChecks={statusCodeMetrics.total}
						periodLabel={activePeriod.toUpperCase()}
					/>
				</div>
			{/if}
		</section>

		<!-- 4. Incidents (2/3) + Alert Policy (1/3) -->
		<section class="grid grid-cols-1 lg:grid-cols-3" aria-label="Incidents and alert policy">
			<div class="py-6 min-w-0 lg:col-span-2 lg:pr-6">
				<MonitorIncidentHistory {orgId} {incidents} {tz} />
			</div>

			<div class="py-6 min-w-0 border-t border-border/60 lg:border-t-0 lg:border-l lg:pl-6">
				<MonitorAlertRulesSection
					{orgId}
					{monitorId}
					monitorType={monitor.Type}
					alertTemplateId={monitor.AlertTemplateID}
				/>
			</div>
		</section>
	{/if}
</div>
