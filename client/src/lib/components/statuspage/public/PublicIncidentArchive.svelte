<script lang="ts">
	import { Archive, ShieldCheck, CheckCircle2 } from 'lucide-svelte';
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
				onclick={() => onSelectRange('7d')}
			>
				7 Days
			</button>
			<button
				type="button"
				style={historyRange === '14d' ? `background-color: ${themeColor}; color: #ffffff;` : ''}
				class="px-2.5 py-1 rounded-md transition-colors font-medium {historyRange === '14d'
					? 'shadow-sm'
					: 'text-muted-foreground hover:text-foreground'}"
				onclick={() => onSelectRange('14d')}
			>
				14 Days
			</button>
			<button
				type="button"
				style={historyRange === 'all' ? `background-color: ${themeColor}; color: #ffffff;` : ''}
				class="px-2.5 py-1 rounded-md transition-colors font-medium {historyRange === 'all'
					? 'shadow-sm'
					: 'text-muted-foreground hover:text-foreground'}"
				onclick={() => onSelectRange('all')}
			>
				90 Days Archive
			</button>
		</div>
	</div>

	<!-- Daily Grouped Incidents Timeline -->
	{#if displayedIncidentHistory.length > 0}
		<div class="space-y-4">
			{#each displayedIncidentHistory as dayGroup (dayGroup.date)}
				<div
					class="rounded-xl border border-border/70 bg-white/3 backdrop-blur-md p-5 transition-all duration-200"
				>
					<div class="flex items-center justify-between pb-3 border-b border-border/40 mb-3">
						<h3 class="text-xs sm:text-sm font-semibold text-foreground font-mono">
							{formatDateHeader(dayGroup.date)}
						</h3>
						{#if dayGroup.incidents.length === 0}
							<span class="text-[11px] text-muted-foreground flex items-center gap-1.5 font-medium">
								<ShieldCheck class="size-3.5" style="color: {themeColor};" />
								<span style="color: {themeColor};">No incidents</span>
							</span>
						{:else}
							<span
								class="text-[11px] font-mono px-2 py-0.5 rounded-full bg-amber-500/20 text-amber-300 border border-amber-500/30"
							>
								{dayGroup.incidents.length} incident{dayGroup.incidents.length > 1 ? 's' : ''}
							</span>
						{/if}
					</div>

					{#if dayGroup.incidents.length > 0}
						<div class="space-y-4 pt-1">
							{#each dayGroup.incidents as incident (incident.id)}
								<div class="rounded-lg border border-border/60 bg-black/10 p-4 space-y-2.5">
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
													>Duration: {formatDuration(
														incident.started_at,
														incident.resolved_at
													)}</span
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
					{:else}
						<p class="text-xs text-muted-foreground leading-relaxed">
							No incidents reported today. All monitored systems operated normally.
						</p>
					{/if}
				</div>
			{/each}
		</div>
	{:else}
		<div
			class="rounded-xl border border-border/60 bg-white/2 backdrop-blur-md p-6 flex items-center gap-4 text-muted-foreground"
		>
			<div
				style="background-color: color-mix(in srgb, {themeColor} 12%, transparent); border-color: color-mix(in srgb, {themeColor} 25%, transparent);"
				class="size-9 rounded-full border flex items-center justify-center shrink-0"
			>
				<ShieldCheck class="size-5" style="color: {themeColor};" />
			</div>
			<div>
				<p class="text-sm font-medium text-foreground">No past incidents on record.</p>
				<p class="text-xs text-muted-foreground mt-0.5">
					All monitored systems have maintained normal operational status.
				</p>
			</div>
		</div>
	{/if}
</section>
