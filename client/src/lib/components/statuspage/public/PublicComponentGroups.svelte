<script lang="ts">
	import { Radio, ChevronDown, FolderTree } from 'lucide-svelte';
	import type { PublicStatusPage, PublicMonitor } from '$lib/api/statuspage';
	import PublicMonitorRow from '$lib/components/statuspage/PublicMonitorRow.svelte';
	import * as Collapsible from '$lib/components/ui/collapsible/index.js';

	interface Props {
		statusPage: PublicStatusPage;
		themeColor: string;
		monitorGrouping: {
			hasGroups: boolean;
			groups: Array<{
				group: { id: string; name: string };
				monitors: PublicMonitor[];
				status: string;
			}>;
			ungrouped: PublicMonitor[];
		};
		collapsedGroups: Record<string, boolean>;
		onToggleGroup: (groupId: string) => void;
		getGroupStatusBadge: (status: string) => {
			label: string;
			dotClass: string;
			dotStyle: string;
			textClass: string;
			textStyle: string;
		};
	}

	let {
		statusPage,
		themeColor,
		monitorGrouping,
		collapsedGroups,
		onToggleGroup,
		getGroupStatusBadge
	}: Props = $props();
</script>

<section class="space-y-4">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-2">
			<Radio class="size-4 text-primary" />
			<h2 class="text-lg font-semibold tracking-tight text-foreground">Services & Systems</h2>
		</div>
		<span class="text-xs text-muted-foreground font-mono">
			{statusPage.monitors?.length || 0} component{statusPage.monitors?.length === 1 ? '' : 's'}
		</span>
	</div>

	{#if statusPage.monitors && statusPage.monitors.length > 0}
		{#if monitorGrouping.hasGroups}
			<!-- Grouped Component Sections -->
			<div class="space-y-6">
				{#each monitorGrouping.groups as gInfo (gInfo.group.id)}
					{@const isCollapsed = collapsedGroups[gInfo.group.id] ?? false}
					{@const gBadge = getGroupStatusBadge(gInfo.status)}

					<Collapsible.Root
						open={!isCollapsed}
						onOpenChange={() => onToggleGroup(gInfo.group.id)}
						class="rounded-xl border border-border/80 bg-white/3 overflow-hidden shadow-sm"
					>
						<!-- Group Header (Collapsible Trigger) -->
						<Collapsible.Trigger
							class="w-full px-5 py-4 flex items-center justify-between gap-3 bg-white/2 hover:bg-white/5 transition-colors text-left cursor-pointer data-[state=open]:border-b data-[state=open]:border-border/50"
						>
							<div class="flex items-center gap-3 min-w-0">
								<ChevronDown
									class="size-4 text-muted-foreground transition-transform duration-200 shrink-0 {isCollapsed
										? '-rotate-90'
										: ''}"
								/>
								<div class="flex items-center gap-2 min-w-0">
									<h3 class="text-base font-semibold text-foreground truncate">
										{gInfo.group.name}
									</h3>
									<span
										class="text-[11px] font-mono text-muted-foreground px-2 py-0.5 rounded-full bg-white/5 border border-border/60"
									>
										{gInfo.monitors.length}
									</span>
								</div>
							</div>

							<div
								class="flex items-center gap-1.5 text-xs font-medium px-2.5 py-1 rounded-full bg-white/5 border border-border/60 shrink-0"
							>
								<span style={gBadge.dotStyle} class="size-2 rounded-full {gBadge.dotClass}"></span>
								<span style={gBadge.textStyle} class={gBadge.textClass}>{gBadge.label}</span>
							</div>
						</Collapsible.Trigger>

						<!-- Group Monitor Rows -->
						<Collapsible.Content
							class="overflow-hidden transition-all data-[state=closed]:animate-collapsible-up data-[state=open]:animate-collapsible-down"
						>
							<div class="p-4 sm:p-5 space-y-4">
								{#if gInfo.monitors.length === 0}
									<div class="text-xs text-muted-foreground py-3 text-center italic">
										No monitors currently in this group.
									</div>
								{:else}
									{#each gInfo.monitors as monitor (monitor.id)}
										<PublicMonitorRow {monitor} {themeColor} />
									{/each}
								{/if}
							</div>
						</Collapsible.Content>
					</Collapsible.Root>
				{/each}

				<!-- Ungrouped Monitors (if any) -->
				{#if monitorGrouping.ungrouped.length > 0}
					<div class="space-y-4 pt-2">
						<div class="flex items-center gap-2 px-1">
							<FolderTree class="size-3.5 text-muted-foreground" />
							<h3 class="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
								Other Components
							</h3>
						</div>
						<div class="space-y-4">
							{#each monitorGrouping.ungrouped as monitor (monitor.id)}
								<PublicMonitorRow {monitor} {themeColor} />
							{/each}
						</div>
					</div>
				{/if}
			</div>
		{:else}
			<!-- Flat List Layout (Fallback when no groups exist) -->
			<div class="space-y-4">
				{#each statusPage.monitors as monitor (monitor.id)}
					<PublicMonitorRow {monitor} {themeColor} />
				{/each}
			</div>
		{/if}
	{:else}
		<div
			class="bg-surface-container border-dashed border border-border rounded-xl p-8 text-center text-muted-foreground font-body-md"
		>
			No services are currently being monitored for this status page.
		</div>
	{/if}
</section>
