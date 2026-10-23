<script lang="ts">
	import type { Monitor } from '$lib/api/monitor';
	import { statusLabel } from '$lib/utils/monitor';
	import { ArrowUpRight } from 'lucide-svelte';

	interface Props {
		monitors: Monitor[];
		orgId: string;
	}

	let { monitors, orgId }: Props = $props();

	function getStatusColor(status: string) {
		switch (status) {
			case 'up':
				return 'bg-emerald-400';
			case 'down':
				return 'bg-rose-500';
			case 'paused':
				return 'bg-amber-400';
			default:
				return 'bg-sky-400';
		}
	}
</script>

<div class="flex flex-col gap-4 min-w-0">
	<div class="flex items-center justify-between gap-3 min-h-7">
		<div>
			<h2 class="text-base font-semibold text-foreground tracking-tight">Linked Monitors</h2>
			<p class="text-xs text-muted-foreground mt-0.5">Monitors subscribed to this template.</p>
		</div>
		<span class="text-xs text-muted-foreground font-mono shrink-0">
			{monitors.length}
		</span>
	</div>

	{#if monitors.length === 0}
		<div class="py-10 text-center space-y-1">
			<p class="text-xs font-medium text-foreground">No monitors linked yet</p>
			<p class="text-[11px] text-muted-foreground">
				Link monitors to this template from their settings or during creation.
			</p>
		</div>
	{:else}
		<ul class="divide-y divide-border/60">
			{#each monitors as m (m.ID)}
				<li>
					<a
						href={`/dashboard/org/${orgId}/monitors/${m.ID}`}
						class="flex items-center justify-between gap-3 py-3 px-2 -mx-2 rounded-md hover:bg-muted/40 transition-colors group"
					>
						<div class="min-w-0 space-y-1 flex-1">
							<div class="flex items-center gap-2">
								<span class="size-2 rounded-full shrink-0 {getStatusColor(m.Status)}"></span>
								<span
									class="font-medium text-sm text-foreground group-hover:text-primary transition-colors truncate"
								>
									{m.Name}
								</span>
								<span
									class="status-badge status-badge--{m.Status} text-[10px] py-0.5 px-2 rounded-full inline-flex items-center font-mono uppercase shrink-0"
								>
									{statusLabel(m.Status)}
								</span>
							</div>
							<div class="text-xs font-mono text-muted-foreground/70 truncate pl-4">
								{m.Target}
							</div>
						</div>

						<ArrowUpRight
							class="size-4 text-muted-foreground group-hover:text-primary transition-colors shrink-0"
						/>
					</a>
				</li>
			{/each}
		</ul>
	{/if}
</div>
