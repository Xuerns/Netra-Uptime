<script lang="ts">
	import type { Monitor } from '$lib/api/monitor';
	import { typeIcon, statusLabel } from '$lib/utils/monitor';
	import { Monitor as MonitorIcon, ArrowUpRight } from 'lucide-svelte';

	interface Props {
		monitors: Monitor[];
		orgId: string;
	}

	let { monitors, orgId }: Props = $props();
</script>

<section class="bg-card border border-border rounded-xl p-5 md:p-6 space-y-4">
	<div>
		<h2 class="text-base font-semibold text-foreground tracking-tight">Linked Monitors</h2>
		<p class="text-xs text-muted-foreground mt-0.5">
			Monitors currently subscribed to this template's alert rules and thresholds.
		</p>
	</div>

	{#if monitors.length === 0}
		<div
			class="rounded-xl border border-dashed border-border bg-white/2 p-6 text-center text-xs text-muted-foreground space-y-1.5"
		>
			<MonitorIcon class="size-6 mx-auto opacity-50 mb-1" />
			<p class="font-medium text-foreground">No monitors linked to this template yet.</p>
			<p class="text-[11px] max-w-[48ch] mx-auto leading-relaxed">
				You can link monitors to this template from each monitor's settings or during monitor
				creation.
			</p>
		</div>
	{:else}
		<div class="divide-y divide-border/60 border border-border/80 rounded-xl overflow-hidden">
			{#each monitors as m (m.ID)}
				{@const MonitorTypeIcon = typeIcon(m.Type)}
				<div
					class="flex items-center justify-between gap-4 p-4 bg-white/2 hover:bg-white/4 transition-colors"
				>
					<div class="flex items-center gap-3.5 min-w-0">
						<div class="p-2 rounded-lg bg-white/5 border border-border text-foreground shrink-0">
							<MonitorTypeIcon class="size-4" />
						</div>
						<div class="min-w-0 space-y-0.5">
							<div class="flex items-center gap-2">
								<a
									href={`/dashboard/org/${orgId}/monitors/${m.ID}`}
									class="font-semibold text-sm text-foreground hover:text-primary transition-colors truncate"
								>
									{m.Name}
								</a>
								<span
									class="status-badge status-badge--{m.Status} text-[10px] py-0.5 px-2 rounded-full inline-flex items-center gap-1 font-mono uppercase"
								>
									{statusLabel(m.Status)}
								</span>
							</div>
							<div class="text-xs font-mono text-muted-foreground/70 truncate">
								{m.Target}
							</div>
						</div>
					</div>

					<a
						href={`/dashboard/org/${orgId}/monitors/${m.ID}`}
						class="inline-flex items-center gap-1 text-xs text-muted-foreground hover:text-primary font-medium transition-colors shrink-0"
					>
						<span>View Monitor</span>
						<ArrowUpRight class="size-3.5" />
					</a>
				</div>
			{/each}
		</div>
	{/if}
</section>
