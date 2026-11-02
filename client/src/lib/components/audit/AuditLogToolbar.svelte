<script lang="ts">
	import {
		ChevronDown,
		CheckCircle2,
		RotateCcw,
		X,
		Search,
		ListFilter,
		ScrollText
	} from 'lucide-svelte';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import type { ComponentType } from 'svelte';

	export interface AuditCategory {
		id: string;
		label: string;
		icon: ComponentType;
		entityTypes: string;
	}

	interface Props {
		categories: AuditCategory[];
		selectedCategories: string[];
		activeCategories: AuditCategory[];
		selectedActions: Array<{ value: string; label: string }>;
		searchQuery: string;
		hasActiveFilters: boolean;
		totalCount: number;
		onCategoryToggle: (catId: string) => void;
		onActionRemove: (actionVal: string) => void;
		onResetFilters: () => void;
		onClearSearch: () => void;
	}

	let {
		categories,
		selectedCategories,
		activeCategories,
		selectedActions,
		searchQuery,
		hasActiveFilters,
		totalCount,
		onCategoryToggle,
		onActionRemove,
		onResetFilters,
		onClearSearch
	}: Props = $props();

	let TriggerIcon = $derived(activeCategories.length === 1 ? activeCategories[0].icon : ScrollText);

	let triggerLabel = $derived(
		activeCategories.length === 0
			? 'All Categories'
			: activeCategories.length === 1
				? activeCategories[0].label
				: `${activeCategories.length} Categories`
	);
</script>

<div class="flex flex-col gap-2.5 mb-6">
	<div
		class="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 p-2.5 bg-card border border-border rounded-lg"
	>
		<!-- Left: Category & Action Selectors -->
		<div class="flex items-center gap-2 flex-wrap">
			<!-- Category Dropdown Trigger -->
			<DropdownMenu.Root>
				<DropdownMenu.Trigger class="focus:outline-none">
					<button
						type="button"
						class="flex items-center gap-2 h-9 px-3 rounded-lg text-xs font-medium border transition-colors cursor-pointer {selectedCategories.length >
						0
							? 'bg-primary/10 border-primary/30 text-primary'
							: 'bg-surface-sidebar border-border text-foreground hover:bg-white/5'}"
					>
						<TriggerIcon
							class="size-3.5 {selectedCategories.length > 0
								? 'text-primary'
								: 'text-muted-foreground'}"
						/>
						<span>{triggerLabel}</span>
						<ChevronDown class="size-3 opacity-60 ml-0.5" />
					</button>
				</DropdownMenu.Trigger>
				<DropdownMenu.Content
					data-lenis-prevent
					class="w-64 max-h-95 overflow-y-auto bg-card border-border text-foreground rounded-lg p-1.5 shadow-xl z-50"
					align="start"
					sideOffset={6}
				>
					<DropdownMenu.Label
						class="text-[10px] text-muted-foreground font-medium px-2 py-1 uppercase tracking-wider"
					>
						Filter by Category
					</DropdownMenu.Label>
					{#each categories as cat (cat.id)}
						{@const isSelected =
							cat.id === 'ALL'
								? selectedCategories.length === 0
								: selectedCategories.includes(cat.id)}
						<DropdownMenu.Item
							closeOnSelect={cat.id === 'ALL'}
							class="cursor-pointer flex items-center justify-between gap-2 px-2.5 py-1.5 rounded-md hover:bg-white/8 text-xs transition-colors {isSelected
								? 'text-primary bg-white/5 font-semibold'
								: 'text-foreground'}"
							onclick={() => onCategoryToggle(cat.id)}
						>
							<div class="flex items-center gap-2">
								<cat.icon
									class="size-3.5 {isSelected ? 'text-primary' : 'text-muted-foreground'}"
								/>
								<span>{cat.label}</span>
							</div>
							{#if isSelected}
								<CheckCircle2 class="size-3 text-primary shrink-0" />
							{/if}
						</DropdownMenu.Item>
					{/each}
				</DropdownMenu.Content>
			</DropdownMenu.Root>

			<!-- Reset Filters button -->
			{#if hasActiveFilters}
				<button
					type="button"
					class="flex items-center gap-1.5 h-9 px-2.5 text-xs text-muted-foreground hover:text-foreground hover:bg-white/5 rounded-lg transition-colors cursor-pointer"
					onclick={onResetFilters}
				>
					<RotateCcw class="size-3" />
					<span>Reset</span>
				</button>
			{/if}
		</div>

		<!-- Right: Total Event Counter -->
		<div class="flex items-center gap-2 self-end sm:self-auto">
			<span
				class="text-xs font-mono text-muted-foreground px-2.5 py-1 bg-surface-sidebar border border-border rounded-md"
			>
				{totalCount} event{totalCount !== 1 ? 's' : ''}
			</span>
		</div>
	</div>

	<!-- Active Filter Badges Bar (shown only when a filter is applied) -->
	{#if hasActiveFilters}
		<div class="flex items-center gap-2 flex-wrap text-xs px-1">
			<span class="text-[11px] text-muted-foreground font-medium">Active:</span>

			{#each activeCategories as cat (cat.id)}
				<span
					class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-primary/10 border border-primary/20 text-primary text-[11px] font-medium"
				>
					<cat.icon class="size-3" />
					<span>Category: {cat.label}</span>
					<button
						type="button"
						class="hover:text-foreground transition-colors cursor-pointer ml-0.5"
						onclick={() => onCategoryToggle(cat.id)}
						aria-label="Remove {cat.label} category filter"
					>
						<X class="size-3" />
					</button>
				</span>
			{/each}

			{#each selectedActions as act (act.value)}
				<span
					class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-primary/10 border border-primary/20 text-primary text-[11px] font-medium"
				>
					<ListFilter class="size-3" />
					<span>Action: {act.label}</span>
					<button
						type="button"
						class="hover:text-foreground transition-colors cursor-pointer ml-0.5"
						onclick={() => onActionRemove(act.value)}
						aria-label="Remove {act.label} action filter"
					>
						<X class="size-3" />
					</button>
				</span>
			{/each}

			{#if searchQuery}
				<span
					class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-white/5 border border-border text-foreground text-[11px] font-medium"
				>
					<Search class="size-3 text-muted-foreground" />
					<span>"{searchQuery}"</span>
					<button
						type="button"
						class="text-muted-foreground hover:text-foreground transition-colors cursor-pointer ml-0.5"
						onclick={onClearSearch}
						aria-label="Clear search"
					>
						<X class="size-3" />
					</button>
				</span>
			{/if}
		</div>
	{/if}
</div>
