<script lang="ts">
	import type { IncidentAssignee } from '$lib/api/incident';
	import { Users, UserCheck } from 'lucide-svelte';

	interface Props {
		assignees?: IncidentAssignee[];
	}

	let { assignees = [] }: Props = $props();
</script>

<div class="bg-card border border-border rounded-lg p-5 space-y-4">
	<div class="flex items-center justify-between">
		<h2 class="text-sm font-semibold text-foreground tracking-tight flex items-center gap-2">
			<Users class="size-4 text-primary" />
			<span>Command Responders & Assignees</span>
		</h2>
		<span class="text-xs font-mono text-muted-foreground">
			{assignees.length} responder{assignees.length === 1 ? '' : 's'}
		</span>
	</div>

	{#if assignees.length === 0}
		<div
			class="p-4 bg-foreground/10 border border-border rounded-lg text-center text-xs text-muted-foreground"
		>
			No responders yet.
		</div>
	{:else}
		<div class="space-y-2">
			{#each assignees as assignee}
				<div class="bg-foreground/15 border border-border rounded-lg p-3 flex items-center gap-3">
					<div
						class="size-8 rounded-lg bg-primary/20 border border-primary/30 text-primary font-bold text-xs flex items-center justify-center shrink-0 font-mono"
					>
						{#if assignee.User?.Avatar}
							<img
								src={assignee.User.Avatar}
								alt={assignee.User?.Name || 'User'}
								class="size-full object-cover rounded-lg"
							/>
						{:else}
							<span>{(assignee.User?.Name || 'User').substring(0, 2).toUpperCase()}</span>
						{/if}
					</div>

					<div class="min-w-0 flex-1 space-y-0.5">
						<div class="text-xs font-semibold text-foreground truncate">
							{assignee.User?.Name || 'Responder'}
						</div>
						<div class="text-[10.5px] font-mono text-muted-foreground flex items-center gap-1">
							<UserCheck class="size-3 text-emerald-400" />
							<span>{assignee.Role || 'Responder'}</span>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
