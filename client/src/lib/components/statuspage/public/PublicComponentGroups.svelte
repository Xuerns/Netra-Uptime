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
				group: { id: string; name: string; is_expanded?: boolean };
				monitors: PublicMonitor[];
				status: string;
			}>;
			ungrouped: PublicMonitor[];
			ungroupedStatus: string;
		};
		collapsedGroups: Record<string, boolean>;
		onToggleGroup: (groupId: string, currentlyCollapsed?: boolean) => void;
	}

	let { statusPage, themeColor, monitorGrouping, collapsedGroups, onToggleGroup }: Props = $props();

	function isGroupCollapsed(
		groupId: string,
		isExpandedDefault: boolean | undefined,
		status: string
	): boolean {
		if (collapsedGroups[groupId] !== undefined) {
			return collapsedGroups[groupId];
		}
		// Auto-expand groups with active disruptions so outages are never hidden
		if (status === 'major_outage' || status === 'degraded') {
			return false;
		}
		return isExpandedDefault === false;
	}
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
			<div class="space-y-4">
				{#each monitorGrouping.groups as gInfo (gInfo.group.id)}
					{@const isCollapsed = isGroupCollapsed(
						gInfo.group.id,
						gInfo.group.is_expanded,
						gInfo.status
					)}

					<Collapsible.Root
						open={!isCollapsed}
						onOpenChange={() => onToggleGroup(gInfo.group.id, isCollapsed)}
						class="rounded-md border border-border/80 bg-card overflow-hidden"
					>
						<!-- Group Header (Collapsible Trigger) -->
						<Collapsible.Trigger
							class="w-full px-5 py-3.5 flex items-center justify-between gap-3 bg-white/2 hover:bg-white/5 transition-colors text-left cursor-pointer data-[state=open]:border-b data-[state=open]:border-border/50"
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
						</Collapsible.Trigger>

						<!-- Group Monitor Rows -->
						<Collapsible.Content
							class="overflow-hidden transition-all data-[state=closed]:animate-collapsible-up data-[state=open]:animate-collapsible-down"
						>
							<div class="p-4 sm:p-5 space-y-3">
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

				<!-- Ungrouped Monitors (Collapsible just like groups) -->
				{#if monitorGrouping.ungrouped.length > 0}
					{@const ungroupedId = '__ungrouped__'}
					{@const isUngroupedCollapsed = isGroupCollapsed(
						ungroupedId,
						true,
						monitorGrouping.ungroupedStatus
					)}

					<Collapsible.Root
						open={!isUngroupedCollapsed}
						onOpenChange={() => onToggleGroup(ungroupedId, isUngroupedCollapsed)}
						class="rounded-md border border-border/80 bg-card overflow-hidden"
					>
						<Collapsible.Trigger
							class="w-full px-5 py-3.5 flex items-center justify-between gap-3 bg-white/2 hover:bg-white/5 transition-colors text-left cursor-pointer data-[state=open]:border-b data-[state=open]:border-border/50"
						>
							<div class="flex items-center gap-3 min-w-0">
								<ChevronDown
									class="size-4 text-muted-foreground transition-transform duration-200 shrink-0 {isUngroupedCollapsed
										? '-rotate-90'
										: ''}"
								/>
								<div class="flex items-center gap-2 min-w-0">
									<FolderTree class="size-4 text-muted-foreground shrink-0" />
									<h3 class="text-base font-semibold text-foreground truncate">Other Components</h3>
									<span
										class="text-[11px] font-mono text-muted-foreground px-2 py-0.5 rounded-full bg-white/5 border border-border/60"
									>
										{monitorGrouping.ungrouped.length}
									</span>
								</div>
							</div>
						</Collapsible.Trigger>

						<Collapsible.Content
							class="overflow-hidden transition-all data-[state=closed]:animate-collapsible-up data-[state=open]:animate-collapsible-down"
						>
							<div class="p-4 sm:p-5 space-y-3">
								{#each monitorGrouping.ungrouped as monitor (monitor.id)}
									<PublicMonitorRow {monitor} {themeColor} />
								{/each}
							</div>
						</Collapsible.Content>
					</Collapsible.Root>
				{/if}
			</div>
		{:else}
			<!-- Flat List Layout (Fallback when no groups exist) -->
			<div class="space-y-3">
				{#each statusPage.monitors as monitor (monitor.id)}
					<PublicMonitorRow {monitor} {themeColor} />
				{/each}
			</div>
		{/if}
	{:else}
		<div
			class="bg-surface-container border-dashed border border-border rounded-md p-8 text-center text-muted-foreground font-body-md"
		>
			No services are currently being monitored for this status page.
		</div>
	{/if}
</section>
