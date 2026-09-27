<div class="bg-card border border-border rounded-lg overflow-hidden">
	<!-- Table Header Bar with GitHub-Style Active/Resolved Toggles -->
	<div
		class="bg-white/3 border-b border-border px-5 py-3.5 flex items-center justify-between text-xs text-muted-foreground"
	>
		<!-- Left: Filter Tabs Skeleton (Active / Resolved) -->
		<div class="flex items-center gap-6">
			<div class="flex items-center gap-2">
				<div class="size-3.5 rounded-full bg-white/10 animate-pulse"></div>
				<div class="h-4 w-12 bg-white/10 rounded animate-pulse"></div>
				<div class="h-5 w-7 bg-white/5 rounded-full animate-pulse"></div>
			</div>
			<div class="flex items-center gap-2">
				<div class="size-3.5 rounded-full bg-white/10 animate-pulse"></div>
				<div class="h-4 w-16 bg-white/10 rounded animate-pulse"></div>
				<div class="h-5 w-7 bg-white/5 rounded-full animate-pulse"></div>
			</div>
		</div>

		<!-- Right: Severity Filter Dropdown & Stats -->
		<div class="flex items-center gap-3.5 text-xs">
			<div class="h-7 w-32 bg-white/5 rounded-md border border-border/60 animate-pulse"></div>
			<div class="h-4 w-px bg-border/60"></div>
			<div class="h-4 w-28 bg-white/5 rounded animate-pulse hidden sm:block"></div>
		</div>
	</div>

	<!-- Incident Rows Skeleton Container -->
	<div class="divide-y divide-white/5">
		{#each Array(5) as _}
			<div class="p-5 flex flex-col md:flex-row md:items-center justify-between gap-4">
				<!-- Left: Status Icon + Info -->
				<div class="flex items-start gap-3.5 min-w-0 flex-1">
					<!-- Status Icon Dot -->
					<div class="pt-1 shrink-0">
						<div class="size-3 rounded-full bg-white/10 animate-pulse"></div>
					</div>

					<!-- Incident Title, Severity & Meta -->
					<div class="space-y-2 min-w-0 flex-1">
						<div class="flex items-center gap-2.5 flex-wrap">
							<!-- Severity Pill -->
							<div class="h-5 w-14 bg-white/10 rounded font-mono animate-pulse"></div>
							<!-- Title / Summary -->
							<div class="h-5 w-64 bg-white/10 rounded animate-pulse"></div>
							<!-- Target Monitor Tag -->
							<div class="h-5 w-28 bg-white/5 border border-border/50 rounded animate-pulse"></div>
						</div>

						<!-- Meta Details Row -->
						<div class="flex items-center gap-3 flex-wrap">
							<div class="h-3.5 w-16 bg-white/5 rounded font-mono animate-pulse"></div>
							<div class="size-1 rounded-full bg-white/10"></div>
							<div class="h-3.5 w-44 bg-white/5 rounded animate-pulse"></div>
							<div class="size-1 rounded-full bg-white/10"></div>
							<div class="h-3.5 w-14 bg-white/5 rounded animate-pulse"></div>
						</div>
					</div>
				</div>

				<!-- Right Actions Group -->
				<div class="flex items-center gap-2.5 shrink-0 self-end md:self-center">
					<div class="h-7 w-28 bg-white/5 rounded-md animate-pulse"></div>
				</div>
			</div>
		{/each}
	</div>

	<!-- Table Footer / Pagination Skeleton -->
	<div class="px-5 py-3 border-t border-border bg-white/2 flex items-center justify-between">
		<div class="h-4 w-32 bg-white/5 rounded animate-pulse"></div>
		<div class="flex items-center gap-2">
			<div class="size-7 rounded-md bg-white/5 border border-border animate-pulse"></div>
			<div class="size-7 rounded-md bg-white/5 border border-border animate-pulse"></div>
		</div>
	</div>
</div>
