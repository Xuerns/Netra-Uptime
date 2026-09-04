<script lang="ts">
	import type { IncidentUpdateResponse } from '$lib/api/incident';
	import { formatDateTime } from '$lib/utils/datetime';
	import { MessageSquare, Clock, Loader2, User } from 'lucide-svelte';

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

<div class="bg-[#161618] border border-white/10 rounded-lg p-5 space-y-4">
	<div class="flex items-center justify-between">
		<h2 class="text-sm font-semibold text-white tracking-tight flex items-center gap-2">
			<MessageSquare class="size-4 text-primary" />
			<span>Incident Communication Feed & Timeline</span>
		</h2>
		<span class="text-xs font-mono text-on-surface-variant">
			{updates.length} message{updates.length === 1 ? '' : 's'}
		</span>
	</div>

	{#if isLoading}
		<div
			class="flex items-center justify-center py-8 text-xs text-on-surface-variant gap-2 font-mono"
		>
			<Loader2 class="size-4 animate-spin text-primary" />
			<span>Loading incident timeline updates...</span>
		</div>
	{:else if updates.length === 0}
		<div class="p-6 bg-black/20 border border-white/5 rounded-lg text-center space-y-1">
			<p class="text-xs text-white font-medium">No updates recorded yet</p>
			<p class="text-[11.5px] text-on-surface-variant">
				Responders will post status findings and resolution updates here.
			</p>
		</div>
	{:else}
		<div
			class="space-y-3 relative before:absolute before:left-4 before:top-3 before:bottom-3 before:w-0.5 before:bg-white/10"
		>
			{#each updates as update}
				<div class="relative pl-9 space-y-1">
					<!-- Timeline Dot Node -->
					<div
						class="absolute left-2.5 top-2.5 -translate-x-1/2 size-3 rounded-full bg-[#161618] border-2 border-primary"
					></div>

					<div class="bg-black/30 border border-white/5 rounded-lg p-3.5 space-y-2">
						<div class="flex items-center justify-between text-xs font-mono">
							<div class="flex items-center gap-2">
								<div
									class="size-5 rounded-full bg-white/10 flex items-center justify-center overflow-hidden"
								>
									{#if update.avatarUrl}
										<img
											src={update.avatarUrl}
											alt={update.username}
											class="size-full object-cover"
										/>
									{:else}
										<User class="size-3 text-white/60" />
									{/if}
								</div>
								<span class="font-semibold text-white">{update.username || 'System Responder'}</span
								>
							</div>

							<span class="text-[11px] text-on-surface-variant flex items-center gap-1">
								<Clock class="size-3" />
								<span>{fmtDateTime(update.createAt)}</span>
							</span>
						</div>

						<p class="text-xs text-on-surface leading-relaxed whitespace-pre-line">
							{update.message}
						</p>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
