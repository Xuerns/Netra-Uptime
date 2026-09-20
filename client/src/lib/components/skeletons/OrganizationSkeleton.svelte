<script lang="ts">
	let { viewMode = 'grid' }: { viewMode?: 'grid' | 'list' } = $props();
</script>

{#if viewMode === 'grid'}
	<div class="grid grid-cols-[repeat(auto-fill,minmax(320px,1fr))] gap-5">
		{#each Array(6) as _}
			<article class="relative flex flex-col gap-2 p-4 bg-white/3 border border-border rounded-lg">
				<div class="flex items-start gap-3">
					<div class="size-11 rounded-md bg-white/5 animate-pulse shrink-0"></div>
					<div class="min-w-0 flex-1 flex flex-col gap-2 mt-1">
						<div class="h-5 w-3/4 bg-white/10 rounded animate-pulse"></div>
						<div class="h-3 w-1/2 bg-white/5 rounded animate-pulse"></div>
					</div>
					<div class="size-2 rounded-full bg-white/10 mt-1.5 shrink-0 animate-pulse"></div>
				</div>

				<div class="h-px bg-white/8 my-1"></div>

				<div class="flex gap-6">
					<div class="h-4 w-20 bg-white/5 rounded animate-pulse"></div>
					<div class="h-4 w-24 bg-white/5 rounded animate-pulse"></div>
				</div>

				<div class="h-[22px] w-full bg-white/5 rounded-[1px] animate-pulse mt-1"></div>

				<div class="flex items-center justify-between mt-2">
					<div class="h-6 w-16 bg-white/5 rounded-full animate-pulse"></div>
					<div class="h-8 w-32 bg-white/5 rounded-md animate-pulse"></div>
				</div>
			</article>
		{/each}
	</div>
{:else}
	<div class="flex flex-col gap-2">
		{#each Array(5) as _}
			<article
				class="flex flex-col lg:flex-row lg:items-center gap-3 lg:gap-5 p-3 lg:py-2 lg:px-5 bg-white/3 border border-border rounded-lg"
			>
				<div class="flex items-start lg:items-center gap-3 flex-1 min-w-[240px]">
					<div class="size-11 rounded-md bg-white/5 animate-pulse shrink-0"></div>
					<div class="min-w-0 flex-1 flex flex-col gap-2 lg:gap-1.5 mt-1 lg:mt-0">
						<div class="h-5 w-40 bg-white/10 rounded animate-pulse"></div>
						<div class="h-3 w-28 bg-white/5 rounded animate-pulse"></div>
					</div>
					<div
						class="size-2 rounded-full bg-white/10 mt-1.5 lg:mt-0 lg:ml-2 shrink-0 animate-pulse"
					></div>
				</div>

				<div class="h-px bg-white/8 w-full lg:hidden"></div>

				<div class="flex gap-5 lg:w-[180px] lg:pl-5 lg:border-l lg:border-border shrink-0 py-1">
					<div class="h-4 w-20 bg-white/5 rounded animate-pulse"></div>
					<div class="h-4 w-24 bg-white/5 rounded animate-pulse"></div>
				</div>

				<div class="h-px bg-white/8 w-full lg:hidden my-0.5"></div>

				<div
					class="flex items-center justify-between lg:w-[320px] lg:pl-5 lg:border-l lg:border-border shrink-0"
				>
					<div class="h-[22px] w-[140px] bg-white/5 rounded-[1px] animate-pulse"></div>
					<div class="flex items-center gap-2">
						<div class="h-6 w-16 bg-white/5 rounded-full animate-pulse hidden md:block"></div>
						<div class="h-8 w-24 bg-white/5 rounded-md animate-pulse"></div>
					</div>
				</div>
			</article>
		{/each}
	</div>
{/if}
