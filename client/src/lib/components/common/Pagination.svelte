<script lang="ts">
	import { ChevronLeft, ChevronRight } from 'lucide-svelte';
	import { cn } from '$lib/helper/utils';

	let {
		currentPage = $bindable(1),
		totalPages = 1,
		totalItems,
		itemName = 'items',
		class: className = '',
		onPageChange
	} = $props<{
		currentPage: number;
		totalPages: number;
		totalItems?: number;
		itemName?: string;
		class?: string;
		onPageChange?: (page: number) => void;
	}>();

	function goToPage(p: number) {
		const target = Math.max(1, Math.min(totalPages, p));
		if (target === currentPage) return;
		currentPage = target;
		onPageChange?.(target);
	}
</script>

{#if totalPages > 1}
	<div
		class={cn(
			'flex items-center justify-between pt-4 border-t border-border/50 text-xs text-muted-foreground',
			className
		)}
	>
		<div>
			{#if totalItems !== undefined}
				Total <span class="text-foreground font-mono font-medium">{totalItems}</span> {itemName}
			{:else}
				<span></span>
			{/if}
		</div>

		<div class="flex items-center gap-2">
			<button
				type="button"
				class="p-1.5 rounded-lg border border-border bg-white/5 hover:bg-white/10 hover:text-foreground disabled:opacity-30 disabled:cursor-not-allowed cursor-pointer transition-all active:scale-95"
				disabled={currentPage <= 1}
				onclick={() => goToPage(currentPage - 1)}
				aria-label="Previous page"
			>
				<ChevronLeft class="size-4" />
			</button>

			<span class="font-mono text-xs text-foreground px-2 select-none">
				Page {currentPage} of {totalPages}
			</span>

			<button
				type="button"
				class="p-1.5 rounded-lg border border-border bg-white/5 hover:bg-white/10 hover:text-foreground disabled:opacity-30 disabled:cursor-not-allowed cursor-pointer transition-all active:scale-95"
				disabled={currentPage >= totalPages}
				onclick={() => goToPage(currentPage + 1)}
				aria-label="Next page"
			>
				<ChevronRight class="size-4" />
			</button>
		</div>
	</div>
{/if}
