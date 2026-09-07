<div class="space-y-4">
	<!-- Search & Filter Control Bar Skeleton -->
	<div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-3 mb-4">
		<!-- Search Input Skeleton -->
		<div class="h-9 flex-1 bg-card border border-border rounded-xl animate-pulse"></div>
		<!-- Severity Filter Select Skeleton -->
		<div class="h-9 w-36 bg-card border border-border rounded-xl animate-pulse shrink-0"></div>
	</div>

	<!-- GitHub-Style Table List Box Skeleton -->
	<div class="bg-card border border-border rounded-2xl overflow-hidden shadow-xl">
		<!-- Table Header Bar Skeleton -->
		<div class="bg-white/3 border-b border-border px-5 py-3.5 flex items-center justify-between">
			<!-- Filter Tabs Skeleton (Active / Resolved) -->
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
			<!-- Right Stats Skeleton -->
			<div class="h-4 w-28 bg-white/5 rounded animate-pulse"></div>
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
								<div class="h-5 w-12 bg-white/10 rounded animate-pulse"></div>
								<!-- Title / Summary -->
								<div class="h-5 w-56 bg-white/10 rounded animate-pulse"></div>
								<!-- Target Monitor Tag -->
								<div class="h-5 w-32 bg-white/5 rounded animate-pulse"></div>
							</div>

							<!-- Meta Details Row -->
							<div class="flex items-center gap-3 flex-wrap">
								<div class="h-3.5 w-16 bg-white/5 rounded animate-pulse"></div>
								<div class="h-3.5 w-44 bg-white/5 rounded animate-pulse"></div>
								<div class="h-3.5 w-20 bg-white/5 rounded animate-pulse"></div>
							</div>
						</div>
					</div>

					<!-- Right Actions Group -->
					<div class="flex items-center gap-2.5 shrink-0 self-end md:self-center">
						<div class="h-8 w-28 bg-white/10 rounded-lg animate-pulse"></div>
						<div class="h-8 w-24 bg-white/5 rounded-lg animate-pulse"></div>
					</div>
				</div>
			{/each}
		</div>
	</div>
</div>
