<script lang="ts">
	import * as Tabs from '$lib/components/ui/tabs/index.js';

	export type Period = '24h' | '7d' | '30d' | '90d' | '1y';

	interface Props {
		activePeriod: Period;
		effectiveDays: number;
		onSelectPeriod: (period: Period) => void;
	}

	let { activePeriod, effectiveDays, onSelectPeriod }: Props = $props();

	const periods: Period[] = ['24h', '7d', '30d', '90d', '1y'];
</script>

<div class="flex flex-col-reverse sm:flex-row sm:items-center gap-2 sm:gap-3">
	<span class="text-[11px] text-muted-foreground font-mono">
		Last {effectiveDays} day{effectiveDays > 1 ? 's' : ''}
	</span>

	<Tabs.Root
		value={activePeriod}
		onValueChange={(val) => {
			if (val) onSelectPeriod(val as Period);
		}}
	>
		<Tabs.List class="h-8">
			{#each periods as period (period)}
				<Tabs.Trigger value={period} class="px-2.5 sm:px-3 py-1 text-xs font-semibold">
					{period.toUpperCase()}
				</Tabs.Trigger>
			{/each}
		</Tabs.List>
	</Tabs.Root>
</div>
