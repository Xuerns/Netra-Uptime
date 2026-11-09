<script lang="ts">
	import { AlertTriangle, Clock } from 'lucide-svelte';
	import type { PublicIncident } from '$lib/api/statuspage';

	interface Props {
		incidents: PublicIncident[];
		getSeverityBadge: (severity: string) => string;
		timeAgo: (dateString?: string) => string;
		formatDateTime: (dateString?: string) => string;
	}

	let { incidents, getSeverityBadge, timeAgo, formatDateTime }: Props = $props();
</script>

{#if incidents && incidents.length > 0}
	<section class="space-y-3 sm:space-y-4">
		<div class="flex items-center justify-between">
			<div class="flex items-center gap-2">
				<AlertTriangle class="size-4 text-amber-400" />
				<h2 class="text-base sm:text-lg font-semibold tracking-tight text-foreground">
					Active Incidents
				</h2>
				<span
					class="px-2 py-0.5 rounded-full text-xs font-mono font-medium bg-amber-500/20 text-amber-300 border border-amber-500/30"
				>
					{incidents.length}
				</span>
			</div>
		</div>

		<div class="space-y-3">
			{#each incidents as incident (incident.id)}
				<div
					class="rounded-md border border-border/80 bg-card overflow-hidden transition-all duration-200 hover:border-border"
				>
					<!-- Incident Header -->
					<div class="p-3.5 sm:p-4 border-b border-border/40 space-y-2">
						<div class="flex flex-wrap items-center justify-between gap-2">
							<div class="flex items-center gap-1.5 min-w-0 flex-wrap">
								<span
									class="text-[10px] font-mono uppercase px-1.5 py-0.5 rounded border {getSeverityBadge(
										incident.severity
									)}"
								>
									{incident.severity}
								</span>
								<span
									class="text-[10px] font-mono uppercase px-1.5 py-0.5 rounded bg-amber-500/10 text-amber-400 border border-amber-500/20"
								>
									Ongoing
								</span>
								{#if incident.monitor_name}
									<span class="text-xs text-muted-foreground flex items-center gap-1 truncate">
										• <span class="text-foreground/90 font-medium">{incident.monitor_name}</span>
									</span>
								{/if}
							</div>
							<span
								class="text-xs text-muted-foreground font-mono flex items-center gap-1 shrink-0"
							>
								<Clock class="size-3" />
								Started {timeAgo(incident.started_at)}
							</span>
						</div>

						<h3 class="text-sm sm:text-base font-semibold text-foreground">
							{incident.summary}
						</h3>
					</div>

					<!-- Incident Timeline Updates -->
					{#if incident.updates && incident.updates.length > 0}
						<div class="p-3.5 sm:p-4 bg-black/10">
							<div
								class="space-y-3 relative before:absolute before:left-2 before:top-2 before:bottom-2 before:w-px before:bg-border/60"
							>
								{#each incident.updates as update (update.created_at)}
									<div class="relative pl-5 space-y-0.5">
										<span
											class="absolute left-1.5 top-1.5 size-1.5 rounded-full bg-amber-400 -translate-x-1/2"
										></span>
										<div
											class="flex items-center gap-2 text-[11px] text-muted-foreground font-mono"
										>
											<span>{formatDateTime(update.created_at)}</span>
										</div>
										<p class="text-xs sm:text-sm text-foreground/90 leading-relaxed">
											{update.message}
										</p>
									</div>
								{/each}
							</div>
						</div>
					{/if}
				</div>
			{/each}
		</div>
	</section>
{/if}
