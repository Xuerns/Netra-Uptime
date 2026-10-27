<script lang="ts">
	import * as Pagination from '$lib/components/ui/pagination/index.js';
	import { cn } from '$lib/utils';

	let {
		currentPage = $bindable(1),
		totalPages = 1,
		totalItems,
		itemName = 'items',
		class: className = '',
		onPageChange
	}: {
		currentPage: number;
		totalPages: number;
		totalItems?: number;
		itemName?: string;
		class?: string;
		onPageChange?: (page: number) => void;
	} = $props();

	function handlePageChange(p: number) {
		const target = Math.max(1, Math.min(totalPages, p));
		if (target !== currentPage) {
			currentPage = target;
		}
		onPageChange?.(target);
	}
</script>

{#if totalPages > 1}
	<div
		class={cn(
			'flex flex-col sm:flex-row items-center justify-between gap-3 pt-4 border-t border-border/50 text-xs text-muted-foreground w-full',
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

		<Pagination.Root
			count={totalPages}
			perPage={1}
			bind:page={currentPage}
			onPageChange={handlePageChange}
			siblingCount={1}
			class="mx-0 w-auto justify-end"
		>
			{#snippet children({ pages, currentPage: current })}
				<Pagination.Content class="gap-1">
					<Pagination.Item>
						<Pagination.Previous
							class="h-8 px-2 text-xs cursor-pointer disabled:cursor-not-allowed"
						/>
					</Pagination.Item>
					{#each pages as page (page.key)}
						{#if page.type === 'ellipsis'}
							<Pagination.Item>
								<Pagination.Ellipsis class="size-8" />
							</Pagination.Item>
						{:else}
							<Pagination.Item>
								<Pagination.Link
									{page}
									isActive={current === page.value}
									class="size-8 text-xs cursor-pointer font-mono"
								>
									{page.value}
								</Pagination.Link>
							</Pagination.Item>
						{/if}
					{/each}
					<Pagination.Item>
						<Pagination.Next class="h-8 px-2 text-xs cursor-pointer disabled:cursor-not-allowed" />
					</Pagination.Item>
				</Pagination.Content>
			{/snippet}
		</Pagination.Root>
	</div>
{/if}
