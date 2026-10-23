<script lang="ts">
	import type { IncidentUpdateResponse } from '$lib/api/incident';
	import { formatDateTime } from '$lib/utils/datetime';
	import { Clock, Loader2, User } from 'lucide-svelte';

	interface Props {
		updates: IncidentUpdateResponse[];
		isLoading: boolean;
		tz: string;
	}

	let { updates = [], isLoading, tz }: Props = $props();

	function fmtDateTime(d: string) {
		return formatDateTime(d, tz);
	}
</script>

<div class="flex flex-col gap-4 min-w-0">
	<div class="flex items-center justify-between gap-3 min-h-7">
		<h2 class="text-base font-semibold text-foreground tracking-tight">Timeline</h2>
		<span class="text-xs font-mono text-muted-foreground shrink-0">
			{updates.length} update{updates.length === 1 ? '' : 's'}
		</span>
	</div>

	{#if isLoading}
		<div
			class="flex items-center justify-center py-8 text-xs text-muted-foreground gap-2 font-mono"
		>
			<Loader2 class="size-4 animate-spin text-primary" />
			<span>Loading timeline updates...</span>
		</div>
	{:else if updates.length === 0}
		<div class="py-10 text-center space-y-1">
			<p class="text-xs text-foreground font-medium">No updates recorded yet</p>
			<p class="text-[11.5px] text-muted-foreground">
				Responders will post status findings and resolution updates here.
			</p>
		</div>
	{:else}
		<div class="relative pt-2">
			{#each updates as update, idx}
				{@const isLast = idx === updates.length - 1}
				<div class="relative pl-7 pb-6 last:pb-1">
					<!-- Timeline vertical line connecting events -->
					{#if !isLast}
						<div class="absolute left-2.5 top-3.5 bottom-0 w-px bg-border/60"></div>
					{/if}
					<!-- Timeline Dot Node -->
					<div
						class="absolute left-2.5 top-1.5 -translate-x-1/2 size-2 rounded-full bg-primary ring-4 ring-background"
					></div>

					<div class="space-y-1.5">
						<div class="flex items-center justify-between text-xs font-mono flex-wrap gap-2">
							<div class="flex items-center gap-2">
								<div
									class="size-5 rounded-full bg-primary/10 flex items-center justify-center overflow-hidden shrink-0"
								>
									{#if update.avatarUrl}
										<img
											src={update.avatarUrl}
											alt={update.username}
											class="size-full object-cover"
										/>
									{:else}
										<User class="size-3 text-primary" />
									{/if}
								</div>
								<span class="font-semibold text-foreground text-xs"
									>{update.username || 'System Responder'}</span
								>
							</div>

							<span class="text-[11px] text-muted-foreground flex items-center gap-1 font-mono">
								<Clock class="size-3" />
								<span>{fmtDateTime(update.createAt)}</span>
							</span>
						</div>

						<p class="text-xs text-foreground/90 leading-relaxed whitespace-pre-line pl-7">
							{update.message}
						</p>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
