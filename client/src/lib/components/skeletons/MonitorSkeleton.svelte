<script lang="ts">
	let { viewMode = 'grid' }: { viewMode?: 'grid' | 'list' } = $props();
</script>

<div class="space-y-6">
	<!-- 1. Overview KPI Stats Hero Skeleton (4 Cards) -->
	<section
		class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4"
		data-purpose="health-overview-skeleton"
	>
		<!-- KPI 1: Avg Uptime -->
		<div class="p-4 rounded-lg bg-card border border-border flex flex-col justify-between h-[136px]">
			<div class="flex items-center justify-between">
				<div class="h-3.5 w-24 bg-white/10 rounded animate-pulse"></div>
				<div class="h-3.5 w-14 bg-white/5 rounded animate-pulse"></div>
			</div>
			<div class="space-y-1">
				<div class="h-7 w-28 bg-white/10 rounded font-mono animate-pulse"></div>
				<div class="h-3 w-32 bg-white/5 rounded animate-pulse"></div>
			</div>
			<div class="w-full bg-muted rounded-full h-1 overflow-hidden">
				<div class="h-full bg-white/10 w-4/5 rounded-full animate-pulse"></div>
			</div>
		</div>

		<!-- KPI 2: Global Avg Latency -->
		<div class="p-4 rounded-lg bg-card border border-border flex flex-col justify-between h-[136px]">
			<div class="flex items-center justify-between">
				<div class="h-3.5 w-28 bg-white/10 rounded animate-pulse"></div>
				<div class="h-3.5 w-20 bg-white/5 rounded animate-pulse"></div>
			</div>
			<div class="space-y-1">
				<div class="h-7 w-24 bg-white/10 rounded font-mono animate-pulse"></div>
				<div class="h-3 w-36 bg-white/5 rounded animate-pulse"></div>
			</div>
			<div class="flex items-center justify-between pt-1 text-[11px]">
				<div class="h-3 w-20 bg-white/5 rounded animate-pulse"></div>
				<div class="h-3 w-24 bg-white/5 rounded animate-pulse"></div>
			</div>
		</div>

		<!-- KPI 3: Certificates & DNS -->
		<div class="p-4 rounded-lg bg-card border border-border flex flex-col justify-between h-[136px]">
			<div class="flex items-center justify-between">
				<div class="h-3.5 w-28 bg-white/10 rounded animate-pulse"></div>
				<div class="h-4 w-14 bg-white/5 rounded-full animate-pulse"></div>
			</div>
			<div class="space-y-1">
				<div class="h-7 w-12 bg-white/10 rounded font-mono animate-pulse"></div>
				<div class="h-3 w-32 bg-white/5 rounded animate-pulse"></div>
			</div>
			<div class="flex items-center justify-between pt-1">
				<div class="h-3 w-16 bg-white/5 rounded animate-pulse"></div>
				<div class="h-3 w-16 bg-white/5 rounded animate-pulse"></div>
			</div>
		</div>

		<!-- KPI 4: Fleet Engine -->
		<div class="p-4 rounded-lg bg-card border border-border flex flex-col justify-between h-[136px]">
			<div class="flex items-center justify-between">
				<div class="h-3.5 w-24 bg-white/10 rounded animate-pulse"></div>
				<div class="h-3.5 w-16 bg-white/5 rounded animate-pulse"></div>
			</div>
			<div class="space-y-1">
				<div class="h-7 w-20 bg-white/10 rounded font-mono animate-pulse"></div>
				<div class="h-3 w-24 bg-white/5 rounded animate-pulse"></div>
			</div>
			<div class="w-full h-1.5 rounded-full bg-muted overflow-hidden flex">
				<div class="h-full bg-white/10 w-3/4 rounded-full animate-pulse"></div>
			</div>
		</div>
	</section>

	<!-- 2. Filter Toolbar Skeleton -->
	<div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 pt-2">
		<!-- Status Filter Pills -->
		<div class="flex items-center gap-1 bg-surface-sidebar p-1 rounded-lg border border-border">
			{#each [48, 88, 96, 64] as w}
				<div class="h-7 rounded-md bg-white/6 animate-pulse" style="width: {w}px"></div>
			{/each}
		</div>

		<!-- View Controls & Time Range -->
		<div class="flex items-center gap-2 self-stretch sm:self-auto justify-end">
			<div class="flex items-center bg-surface-sidebar border border-border rounded-lg p-0.5 gap-1">
				<div class="h-6 w-14 bg-white/5 rounded animate-pulse"></div>
				<div class="h-6 w-8 bg-white/10 rounded animate-pulse"></div>
				<div class="h-6 w-8 bg-white/5 rounded animate-pulse"></div>
			</div>
			<div class="flex items-center bg-surface-sidebar border border-border rounded-lg p-0.5 gap-1">
				<div class="size-6 bg-white/10 rounded animate-pulse"></div>
				<div class="size-6 bg-white/5 rounded animate-pulse"></div>
			</div>
		</div>
	</div>

	<!-- 3. Monitors Grid / List View Skeleton -->
	{#if viewMode === 'grid'}
		<div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 min-[1400px]:grid-cols-4 2xl:grid-cols-4 gap-4">
			{#each Array(8) as _}
				<div
					class="p-4 flex flex-col gap-3 rounded-xl bg-card border border-border animate-pulse justify-between min-h-[230px]"
				>
					<!-- Row 1: Icon + Title + Badge -->
					<div class="flex items-start gap-3">
						<div class="size-8 rounded-lg bg-white/10 shrink-0"></div>
						<div class="min-w-0 flex-1 space-y-1.5">
							<div class="h-4 w-32 bg-white/10 rounded"></div>
							<div class="h-3 w-40 bg-white/5 rounded font-mono"></div>
						</div>
						<div class="h-5 w-20 rounded-full bg-white/10 shrink-0"></div>
					</div>

					<!-- Divider -->
					<div class="h-px bg-white/6 my-0.5"></div>

					<!-- Sparkline Placeholder -->
					<div class="h-8 w-full rounded bg-white/4"></div>

					<!-- Row 2: Metrics -->
					<div class="flex items-center justify-between gap-4">
						<div class="space-y-1">
							<div class="h-2.5 w-12 bg-white/5 rounded"></div>
							<div class="h-4 w-14 bg-white/10 rounded font-mono"></div>
						</div>
						<div class="space-y-1">
							<div class="h-2.5 w-14 bg-white/5 rounded"></div>
							<div class="h-4 w-16 bg-white/10 rounded font-mono"></div>
						</div>
						<div class="space-y-1 ml-auto text-right">
							<div class="h-2.5 w-10 bg-white/5 rounded"></div>
							<div class="h-4 w-12 bg-white/10 rounded font-mono"></div>
						</div>
					</div>

					<!-- Row 3: SSL Badge Placeholder -->
					<div class="pt-1 border-t border-border flex items-center gap-2">
						<div class="h-4 w-20 bg-white/5 rounded"></div>
					</div>

					<!-- Row 4: 24 Segment History Bar Placeholder -->
					<div class="flex items-center gap-0.5 h-3 w-full">
						{#each Array(24) as _}
							<div class="flex-1 h-full rounded-[1px] bg-white/8"></div>
						{/each}
					</div>
				</div>
			{/each}
		</div>
	{:else}
		<div class="flex flex-col border border-border rounded-xl overflow-hidden bg-white/2 divide-y divide-border">
			<!-- Header row -->
			<div class="skeleton-list-row bg-white/3 text-xs text-muted-foreground font-medium" aria-hidden="true">
				<div class="h-3.5 w-20 bg-white/10 rounded animate-pulse"></div>
				<div class="h-3.5 w-14 bg-white/10 rounded animate-pulse"></div>
				<div class="h-3.5 w-16 bg-white/10 rounded animate-pulse"></div>
				<div class="h-3.5 w-16 bg-white/10 rounded animate-pulse"></div>
				<div class="h-3.5 w-14 bg-white/10 rounded animate-pulse"></div>
				<div class="h-3.5 w-16 bg-white/10 rounded animate-pulse"></div>
				<div class="h-3.5 w-12 bg-white/10 rounded ml-auto animate-pulse"></div>
			</div>

			{#each Array(6) as _}
				<div class="skeleton-list-row hover:bg-white/3 transition-colors animate-pulse">
					<!-- 1. Name + target -->
					<div class="flex items-center gap-3 min-w-0">
						<div class="size-8 rounded-lg bg-white/10 shrink-0"></div>
						<div class="min-w-0 flex-1 space-y-1">
							<div class="h-3.5 w-32 bg-white/10 rounded"></div>
							<div class="h-2.5 w-40 bg-white/5 rounded"></div>
						</div>
					</div>

					<!-- 2. Status Badge -->
					<div>
						<div class="h-5 w-20 rounded-full bg-white/10"></div>
					</div>

					<!-- 3. Response Time -->
					<div class="h-4 w-12 bg-white/10 rounded font-mono"></div>

					<!-- 4. Uptime 30d -->
					<div class="h-4 w-14 bg-white/10 rounded font-mono"></div>

					<!-- 5. Interval -->
					<div class="h-4 w-10 bg-white/5 rounded font-mono"></div>

					<!-- 6. Last Checked -->
					<div class="h-3 w-16 bg-white/5 rounded"></div>

					<!-- 7. Actions -->
					<div class="flex items-center justify-end gap-1">
						<div class="size-7 rounded-md bg-white/5"></div>
						<div class="size-7 rounded-md bg-white/5"></div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<style>
	.skeleton-list-row {
		display: grid;
		grid-template-columns: 1fr 100px 90px 90px 80px 110px 80px;
		align-items: center;
		gap: 16px;
		padding: 12px 20px;
	}
</style>
