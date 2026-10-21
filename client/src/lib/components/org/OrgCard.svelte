<script lang="ts">
	import { Link2, Users, Activity } from 'lucide-svelte';
	import { getInitials as initials } from '$lib/utils/string';
	import type { OrgData } from './types';

	let {
		org,
		onSelect
	}: {
		org: OrgData;
		onSelect: (id: string) => void;
	} = $props();

	let activeMonitorsCount = $derived(
		org.Monitors?.filter((m: { Status: string }) => m.Status !== 'down').length || 0
	);
	let totalMonitorsCount = $derived(org.Monitors?.length || 0);
	let allActive = $derived(totalMonitorsCount > 0 && activeMonitorsCount === totalMonitorsCount);
</script>

<button
	class="relative flex flex-col gap-2 p-4 bg-card border border-border rounded-lg shadow-xs transition-all hover:-translate-y-0.5 hover:border-primary/30 group text-left w-full cursor-pointer"
	onclick={() => onSelect(org.ID)}
>
	<div class="flex items-start gap-3">
		<div
			class="flex items-center justify-center size-11 rounded-md bg-primary/15 text-primary font-mono text-sm font-medium shrink-0"
		>
			{initials(org.Name)}
		</div>
		<div class="min-w-0 flex-1">
			<h3 class="text-[17px] font-semibold text-foreground leading-tight mb-0.5 truncate">
				{org.Name}
			</h3>
			<div class="flex items-center gap-1.5 text-[12.5px] text-muted-foreground font-mono truncate">
				<Link2 class="size-3.5 opacity-70 shrink-0" />
				<span class="truncate">netra.app/o/{org.Slug}</span>
			</div>
		</div>
	</div>

	<div class="h-px bg-white/8 my-1"></div>

	<div class="flex gap-6">
		<div class="flex items-center gap-2">
			<Users class="size-3.75 text-muted-foreground shrink-0" />
			<span class="font-mono font-medium text-body-md text-foreground">
				{org.OrganizationMembers?.length || 0}
			</span>
			<span class="text-[13px] text-muted-foreground">
				member{org.OrganizationMembers?.length === 1 ? '' : 's'}
			</span>
		</div>
		<div class="flex items-center gap-2">
			<Activity class="size-3.75 text-muted-foreground shrink-0" />
			<span class="font-mono font-medium text-body-md text-foreground">
				{totalMonitorsCount}
			</span>
			<span class="text-[13px] text-muted-foreground">
				monitor{totalMonitorsCount === 1 ? '' : 's'}
			</span>
		</div>
	</div>

	<!-- Status Monitor -->
	<div class="mt-4 flex flex-col gap-2">
		<div class="flex items-center justify-between text-[13px]">
			<span class="text-muted-foreground font-medium">Status monitor</span>
			{#if totalMonitorsCount > 0}
				<span
					class="{allActive
						? 'text-primary'
						: 'text-destructive'} font-semibold flex items-center gap-1.5 tracking-wide"
				>
					<Activity class="size-3.5" />
					{activeMonitorsCount}/{totalMonitorsCount} active
				</span>
			{:else}
				<span
					class="text-muted-foreground/50 font-semibold flex items-center gap-1.5 tracking-wide"
				>
					<Activity class="size-3.5" />
					0/0 active
				</span>
			{/if}
		</div>
		<div class="flex items-center gap-1 h-3.5" aria-hidden="true">
			{#if org.Monitors && org.Monitors.length > 0}
				{#each org.Monitors as monitor, i (monitor.ID || i)}
					<span
						class="flex-1 h-full rounded-xs {monitor.Status === 'down'
							? 'bg-destructive'
							: 'bg-primary'} min-w-1"
					></span>
				{/each}
			{:else}
				<span class="flex-1 h-full rounded-xs bg-white/10"></span>
			{/if}
		</div>
	</div>
</button>
