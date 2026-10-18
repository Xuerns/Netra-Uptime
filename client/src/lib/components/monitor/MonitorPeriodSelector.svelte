<script lang="ts">
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
	<div class="flex items-center gap-1.5">
		{#each periods as period (period)}
			<button
				type="button"
				class="px-3.5 py-1.5 rounded-md text-xs font-semibold transition-all cursor-pointer {activePeriod ===
				period
					? 'bg-primary text-primary-foreground'
					: 'text-muted-foreground hover:text-foreground hover:bg-white/5'}"
				onclick={() => onSelectPeriod(period)}
			>
				{period.toUpperCase()}
			</button>
		{/each}
	</div>

	<div class="text-xs text-muted-foreground font-mono px-3 hidden sm:block">
		Showing analytics for last {effectiveDays} day{effectiveDays > 1 ? 's' : ''}
	</div>
</div>
