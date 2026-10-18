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

<div
	class="flex items-center justify-between bg-card border border-border rounded-lg p-1.5 flex-wrap gap-2"
>
	<Tabs.Root
		value={activePeriod}
		onValueChange={(val) => {
			if (val) onSelectPeriod(val as Period);
		}}
	>
		<Tabs.List class="h-8">
			{#each periods as period (period)}
				<Tabs.Trigger value={period} class="px-3 py-1 text-xs font-semibold">
					{period.toUpperCase()}
				</Tabs.Trigger>
			{/each}
		</Tabs.List>
	</Tabs.Root>

	<div class="text-xs text-muted-foreground font-mono px-3 hidden sm:block">
		Showing analytics for last {effectiveDays} day{effectiveDays > 1 ? 's' : ''}
	</div>
</div>
