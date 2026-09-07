<script lang="ts">
	let { viewMode = 'grid' }: { viewMode?: 'grid' | 'list' } = $props();
</script>

<div class="h-9 w-full max-w-80 rounded-lg bg-white/6 animate-pulse mb-6"></div>

{#if viewMode === 'grid'}
	<div class="grid grid-cols-[repeat(auto-fill,minmax(450px,1fr))] gap-6">
		{#each Array(6) as _}
			<div
				class="h-[148px] rounded-xl bg-white/4 border border-border animate-pulse p-5 flex flex-col justify-between"
			>
				<div class="flex justify-between items-start">
					<div class="space-y-2">
						<div class="h-5 w-40 bg-white/10 rounded animate-pulse"></div>
						<div class="h-4 w-56 bg-white/5 rounded animate-pulse"></div>
					</div>
					<div class="h-6 w-16 bg-white/5 rounded-full animate-pulse"></div>
				</div>
				<div class="flex items-center justify-between pt-3 border-t border-border">
					<div class="h-4 w-24 bg-white/5 rounded animate-pulse"></div>
					<div class="h-4 w-20 bg-white/10 rounded animate-pulse"></div>
				</div>
			</div>
		{/each}
	</div>
{:else}
	<div
		class="flex flex-col divide-y divide-border border border-border rounded-xl overflow-hidden bg-white/2"
	>
		<div class="h-[42px] bg-white/5 animate-pulse border-b border-border"></div>
		{#each Array(5) as _}
			<div
				class="h-[56px] border-b border-border bg-white/4 animate-pulse px-4 flex items-center justify-between"
			>
				<div class="flex items-center gap-3">
					<div class="size-8 rounded-lg bg-white/6 animate-pulse"></div>
					<div class="h-4 w-36 bg-white/10 rounded animate-pulse"></div>
				</div>
				<div class="h-4 w-32 bg-white/6 rounded animate-pulse"></div>
				<div class="h-5 w-16 bg-white/6 rounded-full animate-pulse"></div>
				<div class="h-4 w-12 bg-white/6 rounded animate-pulse"></div>
				<div class="h-7 w-20 bg-white/6 rounded-md animate-pulse"></div>
			</div>
		{/each}
	</div>
{/if}
