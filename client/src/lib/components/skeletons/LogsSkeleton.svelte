<div class="space-y-4">
	<!-- Table Box Skeleton -->
	<div class="border border-border rounded-lg bg-white/3 backdrop-blur-md overflow-hidden flex flex-col">
		<!-- Table Column Headers -->
		<div class="skeleton-log-header bg-card border-b border-border/80">
			<div class="h-3.5 w-16 bg-white/10 rounded animate-pulse"></div>
			<div class="h-3.5 w-20 bg-white/10 rounded animate-pulse"></div>
			<div class="h-3.5 w-28 bg-white/10 rounded animate-pulse"></div>
			<div class="h-3.5 w-20 bg-white/10 rounded animate-pulse"></div>
			<div class="h-3.5 w-14 bg-white/10 rounded ml-auto animate-pulse"></div>
		</div>

		<!-- Log Rows -->
		<div class="divide-y divide-white/5">
			{#each Array(8) as _}
				<div
					class="text-xs py-3 px-4 flex flex-col md:grid md:grid-cols-[200px_160px_1fr_130px_140px] items-start md:items-center gap-2 md:gap-4 hover:bg-white/3 transition-colors animate-pulse"
				>
					<!-- Col 1: Actor -->
					<div class="flex items-center gap-2.5 min-w-0 w-full md:w-auto">
						<div class="size-7 rounded-lg bg-white/10 shrink-0"></div>
						<div class="space-y-1 min-w-0 flex-1">
							<div class="h-3.5 w-24 bg-white/10 rounded"></div>
							<div class="h-2.5 w-32 bg-white/5 rounded font-mono"></div>
						</div>
					</div>

					<!-- Col 2: Action -->
					<div class="w-full md:w-auto">
						<div class="h-6 w-28 rounded-md bg-white/10"></div>
					</div>

					<!-- Col 3: Description -->
					<div class="min-w-0 w-full md:w-auto space-y-1">
						<div class="h-3.5 w-3/4 bg-white/10 rounded"></div>
						<div class="h-2.5 w-1/2 bg-white/5 rounded font-mono"></div>
					</div>

					<!-- Col 4: IP Address -->
					<div class="w-full md:w-auto">
						<div class="h-4 w-24 rounded bg-white/5 font-mono"></div>
					</div>

					<!-- Col 5: Time -->
					<div class="w-full md:w-auto flex md:justify-end">
						<div class="h-3.5 w-16 bg-white/5 rounded font-mono"></div>
					</div>
				</div>
			{/each}
		</div>
	</div>

	<!-- Pagination Footer Skeleton -->
	<div class="flex items-center justify-between py-2 text-xs">
		<div class="h-3.5 w-32 bg-white/5 rounded animate-pulse"></div>
		<div class="flex items-center gap-2">
			<div class="h-8 w-20 rounded-lg bg-white/5 border border-border animate-pulse"></div>
			<div class="h-3.5 w-20 bg-white/5 rounded animate-pulse font-mono"></div>
			<div class="h-8 w-16 rounded-lg bg-white/5 border border-border animate-pulse"></div>
		</div>
	</div>
</div>

<style>
	.skeleton-log-header {
		display: none;
		padding: 0.625rem 1rem;
	}

	@media (min-width: 768px) {
		.skeleton-log-header {
			display: grid;
			grid-template-columns: 200px 160px 1fr 130px 140px;
			gap: 1rem;
			align-items: center;
		}
	}
</style>
