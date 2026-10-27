<script lang="ts">
	import type { Incident } from '$lib/api/incident';
	import { CheckCircle2, Server, Clock, Users, ChevronRight } from 'lucide-svelte';
	import { formatDateTime, formatIncidentTimeAgo } from '$lib/utils/datetime';

	interface Props {
		incident: Incident;
		orgId: string;
		timezone: string;
	}

	let { incident, orgId, timezone }: Props = $props();

	let isOngoing = $derived(incident.Status === 'ongoing');

	function fmtDateTime(d: string) {
		return formatDateTime(d, timezone);
	}
	function fmtTimeAgo(d: string) {
		return formatIncidentTimeAgo(d);
	}
</script>

<div class="transition-colors hover:bg-white/5">
	<!-- Main Row Content -->
	<div class="p-5 flex flex-col md:flex-row md:items-center justify-between gap-4">
		<!-- Left: Status Icon + Info -->
		<div class="flex items-start gap-3.5 min-w-0 flex-1">
			<!-- Status Icon Dot -->
			<div class="pt-1 shrink-0">
				{#if isOngoing}
					<span class="relative flex size-3">
						<span class="inline-flex rounded-full size-3 bg-rose-500"></span>
					</span>
				{:else}
					<CheckCircle2 class="size-4 text-emerald-400 shrink-0" />
				{/if}
			</div>

			<!-- Incident Title, Severity & Meta -->
			<div class="space-y-1.5 min-w-0 flex-1">
				<div class="flex items-center gap-2.5 flex-wrap">
					<!-- Severity Pill -->
					{#if incident.SeverityCode}
						<span
							class="px-2 py-0.5 rounded text-[10.5px] font-semibold uppercase tracking-wider font-mono border {incident.SeverityCode ===
							'sev1'
								? 'bg-rose-500/10 text-rose-400 border-rose-500/20'
								: incident.SeverityCode === 'sev2'
									? 'bg-amber-500/10 text-amber-400 border-amber-500/20'
									: incident.SeverityCode === 'sev3'
										? 'bg-sky-500/10 text-sky-400 border-sky-500/20'
										: 'bg-slate-500/10 text-slate-400 border-slate-500/20'}"
						>
							{incident.SeverityLabel || incident.Severity || incident.SeverityCode}
						</span>
					{/if}

					<!-- Title / Summary -->
					<a
						href="/dashboard/org/{orgId}/incidents/{incident.ID}"
						class="text-sm font-semibold text-foreground hover:text-primary transition-colors truncate"
					>
						{incident.Summary || 'Service Disruption'}
					</a>

					<!-- Target Monitor Tag -->
					{#if incident.Monitor}
						<a
							href="/dashboard/org/{orgId}/monitors/{incident.MonitorID}"
							class="inline-flex items-center gap-1 px-2 py-0.5 rounded bg-white/5 border border-border text-[11px] font-mono text-muted-foreground hover:text-foreground hover:border-border transition-all"
						>
							<Server class="size-3 text-primary" />
							<span class="truncate max-w-36">{incident.Monitor.Name}</span>
						</a>
					{/if}
				</div>

				<!-- Meta Details Row -->
				<div class="flex items-center gap-3 text-xs text-muted-foreground flex-wrap font-mono">
					<span class="text-foreground/40">#{incident.ID.substring(0, 8)}</span>
					<span>&middot;</span>
					<span class="flex items-center gap-1">
						<Clock class="size-3" />
						<span
							>Started: {fmtDateTime(incident.StartedAt)} ({fmtTimeAgo(incident.StartedAt)})</span
						>
					</span>

					{#if !isOngoing && incident.ResolvedAt}
						<span>&middot;</span>
						<span class="text-emerald-400">Resolved: {fmtDateTime(incident.ResolvedAt)}</span>
					{/if}

					{#if incident.Assignees && incident.Assignees.length > 0}
						<span>&middot;</span>
						<span class="flex items-center gap-1">
							<Users class="size-3 text-primary" />
							<span
								>{incident.Assignees.length}
								{incident.Assignees.length === 1 ? 'Assignee' : 'Assignees'}</span
							>
						</span>
					{/if}
				</div>
			</div>
		</div>

		<!-- Right Actions Group: Direct link to Incident Command Room -->
		<div class="flex items-center gap-2.5 shrink-0 self-end md:self-center group transition-colors">
			<a
				href="/dashboard/org/{orgId}/incidents/{incident.ID}"
				class="px-3.5 py-1.5 text-foreground/60 group-hover:text-foreground text-xs font-semibold transition-colors flex items-center gap-1.5 cursor-pointer"
			>
				<span>Command Room</span>
				<ChevronRight
					class="size-3.5 text-foreground/60 group-hover:text-foreground transition-colors"
				/>
			</a>
		</div>
	</div>
</div>
