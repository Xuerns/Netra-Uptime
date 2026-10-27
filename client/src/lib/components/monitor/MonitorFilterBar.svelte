<script lang="ts">
	import { LayoutGrid, List } from 'lucide-svelte';

	export type StatusFilter = 'all' | 'up' | 'down' | 'flagged' | 'paused';
	export type ViewMode = 'grid' | 'list';

	interface Props {
		statusFilter: StatusFilter;
		viewMode: ViewMode;
		counts: Record<string, number>;
		onFilterChange: (s: StatusFilter) => void;
		onViewModeChange: (m: ViewMode) => void;
	}

	let { statusFilter, viewMode, counts, onFilterChange, onViewModeChange }: Props = $props();
</script>

<div
	class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 pt-2"
	data-purpose="monitors-filter-bar"
>
	<!-- Status Filter Pills -->
	<div
		class="flex items-center gap-1 bg-surface-sidebar p-1 rounded-lg border border-border text-xs flex-wrap"
		role="group"
		aria-label="Filter by status"
	>
		<button
			type="button"
			class="px-3 py-1 font-medium rounded-md transition-colors cursor-pointer {statusFilter ===
			'all'
				? 'bg-muted text-foreground'
				: 'text-muted-foreground hover:text-foreground'}"
			onclick={() => onFilterChange('all')}
		>
			All <span class="ml-1 text-muted-foreground font-mono text-[11px]">{counts.total ?? 0}</span>
		</button>

		<button
			type="button"
			class="px-3 py-1 font-medium rounded-md transition-colors cursor-pointer {statusFilter ===
			'up'
				? 'bg-muted text-foreground'
				: 'text-muted-foreground hover:text-foreground'}"
			onclick={() => onFilterChange('up')}
		>
			Operational <span class="ml-1 text-emerald-400 font-mono text-[11px]">{counts.up ?? 0}</span>
		</button>

		<button
			type="button"
			class="px-3 py-1 font-medium rounded-md transition-colors cursor-pointer {statusFilter ===
			'down'
				? 'bg-muted text-foreground'
				: 'text-muted-foreground hover:text-foreground'}"
			onclick={() => onFilterChange('down')}
		>
			Down <span class="ml-1 text-rose-400 font-mono text-[11px]">{counts.down ?? 0}</span>
		</button>

		<button
			type="button"
			class="px-3 py-1 font-medium rounded-md transition-colors cursor-pointer {statusFilter ===
			'flagged'
				? 'bg-muted text-foreground'
				: 'text-muted-foreground hover:text-foreground'}"
			onclick={() => onFilterChange('flagged')}
		>
			DNS / SSL Flagged <span class="ml-1 text-amber-400 font-mono text-[11px]"
				>{counts.flagged ?? 0}</span
			>
		</button>

		<button
			type="button"
			class="px-3 py-1 font-medium rounded-md transition-colors cursor-pointer {statusFilter ===
			'paused'
				? 'bg-muted text-foreground'
				: 'text-muted-foreground hover:text-foreground'}"
			onclick={() => onFilterChange('paused')}
		>
			Paused <span class="ml-1 text-muted-foreground font-mono text-[11px]"
				>{counts.paused ?? 0}</span
			>
		</button>
	</div>

	<!-- View Controls -->
	<div class="flex items-center gap-2 self-stretch sm:self-auto justify-end">
		<!-- View Mode Toggle -->
		<div class="flex items-center bg-surface-sidebar border border-border rounded-lg p-0.5">
			<button
				type="button"
				class="p-1.5 rounded transition-colors cursor-pointer {viewMode === 'grid'
					? 'bg-muted text-foreground'
					: 'text-muted-foreground hover:text-foreground'}"
				onclick={() => onViewModeChange('grid')}
				aria-label="Grid view"
				aria-pressed={viewMode === 'grid'}
			>
				<LayoutGrid class="size-3.5" />
			</button>
			<button
				type="button"
				class="p-1.5 rounded transition-colors cursor-pointer {viewMode === 'list'
					? 'bg-muted text-foreground'
					: 'text-muted-foreground hover:text-foreground'}"
				onclick={() => onViewModeChange('list')}
				aria-label="List view"
				aria-pressed={viewMode === 'list'}
			>
				<List class="size-3.5" />
			</button>
		</div>
	</div>
</div>
