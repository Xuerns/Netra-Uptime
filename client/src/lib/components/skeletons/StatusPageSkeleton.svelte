<script lang="ts">
	let { viewMode = 'grid' }: { viewMode?: 'grid' | 'list' } = $props();
</script>

{#if viewMode === 'grid'}
	<div class="grid grid-cols-[repeat(auto-fill,minmax(450px,1fr))] gap-6">
		{#each Array(6) as _}
			<div
				class="bg-card border border-border rounded-xl p-5 flex flex-col justify-between shadow-sm min-h-[210px] animate-pulse"
			>
				<!-- Top row: icon + title + badge + 3-dot menu -->
				<div class="flex items-start justify-between gap-3 mb-3">
					<div class="flex items-center gap-3 min-w-0">
						<div
							class="size-9 rounded-lg bg-white/5 border border-border shrink-0"
						></div>
						<div class="space-y-1.5 min-w-0 flex-1">
							<div class="h-4 w-40 bg-white/10 rounded"></div>
							<div class="h-3 w-28 bg-white/5 rounded font-mono"></div>
						</div>
					</div>

					<div class="flex items-center gap-2 shrink-0">
						<div class="h-5 w-16 bg-white/10 rounded-full"></div>
						<div class="size-8 rounded-lg bg-white/5 border border-border"></div>
					</div>
				</div>

				<!-- Organization Pill -->
				<div class="h-6 w-32 rounded-md bg-white/5 mb-3"></div>

				<!-- Description lines -->
				<div class="space-y-1.5 mb-4 flex-1">
					<div class="h-3.5 w-full bg-white/5 rounded"></div>
					<div class="h-3.5 w-2/3 bg-white/5 rounded"></div>
				</div>

				<!-- Footer: monitor count + view link -->
				<div class="flex items-center justify-between mt-auto pt-4 border-t border-border">
					<div class="h-3.5 w-20 bg-white/5 rounded"></div>
					<div class="h-3.5 w-24 bg-white/10 rounded"></div>
				</div>
			</div>
		{/each}
	</div>
{:else}
	<div
		class="flex flex-col divide-y divide-border border border-border rounded-lg bg-white/2 overflow-hidden"
	>
		<!-- Header row -->
		<div class="skeleton-status-row bg-white/3 text-xs text-muted-foreground font-medium" aria-hidden="true">
			<div class="h-3.5 w-20 bg-white/10 rounded animate-pulse"></div>
			<div class="h-3.5 w-20 bg-white/10 rounded animate-pulse"></div>
			<div class="h-3.5 w-16 bg-white/10 rounded animate-pulse"></div>
			<div class="h-3.5 w-16 bg-white/10 rounded animate-pulse"></div>
			<div class="h-3.5 w-12 bg-white/10 rounded ml-auto animate-pulse"></div>
		</div>

		<!-- Rows -->
		{#each Array(5) as _}
			<div class="skeleton-status-row hover:bg-white/3 transition-colors animate-pulse">
				<!-- 1. Title + Org -->
				<div class="flex items-center gap-3 min-w-0">
					<div class="size-4 rounded bg-white/10 shrink-0"></div>
					<div class="min-w-0 flex-1 space-y-1">
						<div class="h-3.5 w-32 bg-white/10 rounded"></div>
						<div class="h-2.5 w-24 bg-white/5 rounded"></div>
					</div>
				</div>

				<!-- 2. Subdomain -->
				<div class="h-3.5 w-28 bg-white/5 rounded font-mono"></div>

				<!-- 3. Visibility -->
				<div>
					<div class="h-5 w-16 bg-white/10 rounded-full"></div>
				</div>

				<!-- 4. Monitors -->
				<div class="h-3.5 w-16 bg-white/5 rounded"></div>

				<!-- 5. Actions -->
				<div class="flex items-center justify-end gap-1.5">
					<div class="size-7 rounded bg-white/5 border border-border"></div>
					<div class="size-7 rounded bg-white/5 border border-border"></div>
				</div>
			</div>
		{/each}
	</div>
{/if}

<style>
	.skeleton-status-row {
		display: grid;
		grid-template-columns: 1fr 1fr 100px 100px 74px;
		align-items: center;
		gap: 16px;
		padding: 12px 20px;
	}
</style>
