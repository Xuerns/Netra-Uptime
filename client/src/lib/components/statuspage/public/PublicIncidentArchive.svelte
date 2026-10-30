<script lang="ts">
	import { Archive, ShieldCheck, CheckCircle2, ChevronDown } from 'lucide-svelte';
	import type { PublicIncidentDayGroup } from '$lib/api/statuspage';

	interface Props {
		historyRange: '7d' | '14d' | 'all';
		displayedIncidentHistory: PublicIncidentDayGroup[];
		themeColor: string;
		onSelectRange: (range: '7d' | '14d' | 'all') => void;
		formatDateHeader: (dateStr: string) => string;
		getSeverityBadge: (severity: string) => string;
		formatDuration: (startedAt?: string, resolvedAt?: string | null) => string;
		formatDateTime: (dateString?: string) => string;
	}

	let {
		historyRange,
		displayedIncidentHistory,
		themeColor,
		onSelectRange,
		formatDateHeader,
		getSeverityBadge,
		formatDuration,
		formatDateTime
	}: Props = $props();

	const INITIAL_VISIBLE_DAYS = 5;
	let showAllDays = $state(false);

	let visibleDays = $derived(
		showAllDays ? displayedIncidentHistory : displayedIncidentHistory.slice(0, INITIAL_VISIBLE_DAYS)
	);

	let rangeLabel = $derived(
		historyRange === '7d' ? '7 days' : historyRange === '14d' ? '14 days' : '90 days'
	);

	function handleSelectRange(range: '7d' | '14d' | 'all') {
		showAllDays = false;
		onSelectRange(range);
	}
</script>

<section class="space-y-4">
	<div class="flex flex-wrap items-center justify-between gap-3">
		<div class="flex items-center gap-2">
			<Archive class="size-4 text-muted-foreground" />
			<h2 class="text-lg font-semibold tracking-tight text-foreground">Past Incidents Archive</h2>
		</div>

		<!-- Date Range Filter Tabs -->
		<div class="flex items-center rounded-lg bg-white/5 p-1 border border-border/60 text-xs">
			<button
				type="button"
				style={historyRange === '7d' ? `background-color: ${themeColor}; color: #ffffff;` : ''}
				class="px-2.5 py-1 rounded-md transition-colors font-medium {historyRange === '7d'
					? 'shadow-sm'
					: 'text-muted-foreground hover:text-foreground'}"
				onclick={() => handleSelectRange('7d')}
			>
				7 Days
			</button>
			<button
				type="button"
				style={historyRange === '14d' ? `background-color: ${themeColor}; color: #ffffff;` : ''}
				class="px-2.5 py-1 rounded-md transition-colors font-medium {historyRange === '14d'
					? 'shadow-sm'
					: 'text-muted-foreground hover:text-foreground'}"
				onclick={() => handleSelectRange('14d')}
			>
				14 Days
			</button>
			<button
				type="button"
				style={historyRange === 'all' ? `background-color: ${themeColor}; color: #ffffff;` : ''}
				class="px-2.5 py-1 rounded-md transition-colors font-medium {historyRange === 'all'
					? 'shadow-sm'
					: 'text-muted-foreground hover:text-foreground'}"
				onclick={() => handleSelectRange('all')}
			>
				90 Days Archive
			</button>
		</div>
	</div>

	<!-- Incidents-Only Timeline -->
	{#if displayedIncidentHistory.length > 0}
		<div class="space-y-3">
			{#each visibleDays as dayGroup (dayGroup.date)}
				<div
					class="rounded-md border border-border/70 bg-card p-4 sm:p-5 transition-all duration-200"
				>
					<div class="flex items-center justify-between pb-3 border-b border-border/40 mb-3">
						<h3 class="text-xs sm:text-sm font-semibold text-foreground font-mono">
							{formatDateHeader(dayGroup.date)}
						</h3>
						<span class="text-xs font-mono text-muted-foreground">
							{dayGroup.incidents.length} incident{dayGroup.incidents.length > 1 ? 's' : ''}
						</span>
					</div>

					<div class="space-y-3 pt-1">
						{#each dayGroup.incidents as incident (incident.id)}
							<div class="rounded-lg border border-border/60 bg-black/10 p-3.5 sm:p-4 space-y-2">
								<div class="flex flex-wrap items-center justify-between gap-2">
									<div class="flex items-center gap-2">
										<span
											style="background-color: color-mix(in srgb, {themeColor} 12%, transparent); border-color: color-mix(in srgb, {themeColor} 25%, transparent); color: {themeColor};"
											class="inline-flex items-center gap-1 text-[11px] font-mono uppercase px-2 py-0.5 rounded border"
										>
											<CheckCircle2 class="size-3" />
											Resolved
										</span>
										<span
											class="text-[11px] font-mono uppercase px-1.5 py-0.5 rounded border {getSeverityBadge(
												incident.severity
											)}"
										>
											{incident.severity}
										</span>
										{#if incident.monitor_name}
											<span class="text-xs text-muted-foreground">
												• {incident.monitor_name}
											</span>
										{/if}
									</div>

									<div class="flex items-center gap-2 text-xs text-muted-foreground font-mono">
										{#if incident.started_at && incident.resolved_at}
											<span
												>Duration: {formatDuration(incident.started_at, incident.resolved_at)}</span
											>
											<span>•</span>
										{/if}
										<span>{formatDateTime(incident.resolved_at || incident.started_at)}</span>
									</div>
								</div>

								<h4 class="text-sm font-semibold text-foreground">
									{incident.summary}
								</h4>

								{#if incident.updates && incident.updates.length > 0}
									<p
										class="text-xs text-muted-foreground leading-relaxed pl-3 border-l-2 border-border/60"
									>
										{incident.updates[0].message}
									</p>
								{/if}
							</div>
						{/each}
					</div>
				</div>
			{/each}

			{#if displayedIncidentHistory.length > INITIAL_VISIBLE_DAYS}
				<button
					type="button"
					onclick={() => (showAllDays = !showAllDays)}
					class="w-full py-2.5 px-4 rounded-md border border-border/60 bg-white/3 hover:bg-white/6 text-xs font-medium text-muted-foreground hover:text-foreground transition-colors flex items-center justify-center gap-1.5 cursor-pointer"
				>
					<span>
						{showAllDays
							? 'Show fewer incident days'
							: `Show all ${displayedIncidentHistory.length} incident days (${displayedIncidentHistory.length - INITIAL_VISIBLE_DAYS} more)`}
					</span>
					<ChevronDown
						class="size-3.5 transition-transform duration-200 {showAllDays ? 'rotate-180' : ''}"
					/>
				</button>
			{/if}
		</div>
	{:else}
		<div
			class="rounded-md border border-border/60 bg-card px-5 py-4 flex items-center justify-between gap-4 text-muted-foreground"
		>
			<div class="flex items-center gap-3.5">
				<div
					style="background-color: color-mix(in srgb, {themeColor} 12%, transparent); border-color: color-mix(in srgb, {themeColor} 25%, transparent);"
					class="size-8 rounded-full border flex items-center justify-center shrink-0"
				>
					<ShieldCheck class="size-4" style="color: {themeColor};" />
				</div>
				<div>
					<p class="text-sm font-medium text-foreground">
						No incidents reported in the last {rangeLabel}.
					</p>
					<p class="text-xs text-muted-foreground mt-0.5">
						All monitored systems operated normally during this period.
					</p>
				</div>
			</div>
			<span
				style="color: {themeColor};"
				class="hidden sm:inline-flex items-center gap-1.5 text-xs font-mono font-medium shrink-0"
			>
				100% Incident-Free
			</span>
		</div>
	{/if}
</section>
