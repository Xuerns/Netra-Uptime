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
</script>

<button
	class="flex flex-col lg:flex-row lg:items-center gap-3 lg:gap-5 p-3 lg:py-2 lg:px-5 bg-white/3 border border-border rounded-lg transition-all hover:-translate-y-px hover:border-primary/30 group text-left w-full cursor-pointer"
	onclick={() => onSelect(org.ID)}
>
	<div class="flex items-start lg:items-center gap-3 flex-1 min-w-60">
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

	<div class="h-px bg-white/8 w-full lg:hidden"></div>

	<div class="flex gap-5 lg:w-45 lg:pl-5 lg:border-l lg:border-border shrink-0">
		<div class="flex items-center gap-1.5">
			<Users class="size-3.75 text-muted-foreground shrink-0" />
			<span class="font-mono font-medium text-body-md text-foreground">
				{org.OrganizationMembers?.length || 0}
			</span>
			<span class="text-[12px] text-muted-foreground">
				member{org.OrganizationMembers?.length === 1 ? '' : 's'}
			</span>
		</div>
		<div class="flex items-center gap-1.5">
			<Activity class="size-3.75 text-muted-foreground shrink-0" />
			<span class="font-mono font-medium text-body-md text-foreground">
				{org.Monitors?.length || 0}
			</span>
			<span class="text-[12px] text-muted-foreground">
				monitor{org.Monitors?.length === 1 ? '' : 's'}
			</span>
		</div>
	</div>

	<div class="h-px bg-white/8 w-full lg:hidden my-0.5"></div>

	<div
		class="flex items-center justify-between lg:w-50 lg:pl-5 lg:border-l lg:border-border shrink-0"
	>
		<div class="flex items-end gap-0.5 h-5.5 w-35" aria-hidden="true">
			{#if org.Monitors && org.Monitors.length > 0}
				{#each org.Monitors as monitor, i (monitor.ID || i)}
					<span
						class="flex-1 h-full rounded-[1px] {monitor.Status === 'down'
							? 'bg-destructive'
							: 'bg-primary'} opacity-85 min-w-0.5"
					></span>
				{/each}
			{:else}
				<span class="flex-1 h-full rounded-[1px] bg-white/10"></span>
			{/if}
		</div>
	</div>
</button>
