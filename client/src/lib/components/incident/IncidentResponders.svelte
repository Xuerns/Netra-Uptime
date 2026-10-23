<script lang="ts">
	import type { IncidentAssignee } from '$lib/api/incident';
	import { UserCheck } from 'lucide-svelte';

	interface Props {
		assignees?: IncidentAssignee[];
	}

	let { assignees = [] }: Props = $props();
</script>

<div class="flex flex-col gap-4 min-w-0">
	<div class="flex items-center justify-between gap-3 min-h-7">
		<h2 class="text-base font-semibold text-foreground tracking-tight">Command Responders</h2>
		<span class="text-xs font-mono text-muted-foreground shrink-0">
			{assignees.length}
		</span>
	</div>

	{#if assignees.length === 0}
		<div class="py-10 text-center space-y-1">
			<p class="text-xs font-medium text-foreground">No responders assigned yet</p>
			<p class="text-[11px] text-muted-foreground">
				Team members can join this incident command bridge.
			</p>
		</div>
	{:else}
		<ul class="divide-y divide-border/60">
			{#each assignees as assignee (assignee.UserID || assignee.ID)}
				<li class="py-3 flex items-center gap-3">
					<div
						class="size-8 rounded-full bg-primary/10 text-primary font-bold text-xs flex items-center justify-center shrink-0 font-mono overflow-hidden"
					>
						{#if assignee.User?.Avatar}
							<img
								src={assignee.User.Avatar}
								alt={assignee.User?.Name || 'User'}
								class="size-full object-cover"
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
							<UserCheck class="size-3 text-emerald-400 shrink-0" />
							<span>{assignee.Role || 'Responder'}</span>
						</div>
					</div>
				</li>
			{/each}
		</ul>
	{/if}
</div>
