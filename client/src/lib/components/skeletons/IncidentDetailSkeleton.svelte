<div class="space-y-6">
	<!-- 1. Header Section Skeleton -->
	<div
		class="bg-card border border-border rounded-lg p-5 md:p-6 shadow-sm flex flex-col md:flex-row md:items-center justify-between gap-4"
	>
		<div class="space-y-3 flex-1 min-w-0">
			<div class="flex items-center gap-2.5 flex-wrap">
				<!-- Status Badge -->
				<div class="h-6 w-24 bg-white/10 rounded-full animate-pulse"></div>
				<!-- Severity Badge -->
				<div class="h-6 w-20 bg-white/10 rounded-md font-mono animate-pulse"></div>
				<!-- Incident ID -->
				<div class="h-4 w-20 bg-white/5 rounded font-mono animate-pulse"></div>
			</div>
			<!-- Title / Summary -->
			<div class="h-7 w-3/4 max-w-xl bg-white/10 rounded-md animate-pulse"></div>
			<!-- Metadata row (Started, Monitor, Target) -->
			<div class="flex items-center gap-4 text-xs pt-1 flex-wrap">
				<div class="h-4 w-44 bg-white/5 rounded animate-pulse"></div>
				<div class="h-4 w-36 bg-white/5 rounded animate-pulse"></div>
			</div>
		</div>

		<!-- Right Actions (Join / Leave Bridge button) -->
		<div class="flex items-center gap-2 shrink-0 self-start md:self-center">
			<div class="h-9 w-36 bg-white/10 rounded-lg animate-pulse"></div>
		</div>
	</div>

	<!-- 2. Communication Feed & Responders Side by Side -->
	<div class="grid grid-cols-1 lg:grid-cols-4 gap-6">
		<!-- Timeline Feed (75% / lg:col-span-3) -->
		<div class="lg:col-span-3">
			<div class="bg-card border border-border rounded-lg p-5 space-y-4">
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-2">
						<div class="size-4 rounded bg-white/10 animate-pulse"></div>
						<div class="h-5 w-56 bg-white/10 rounded animate-pulse"></div>
					</div>
					<div class="h-4 w-20 bg-white/5 rounded animate-pulse font-mono"></div>
				</div>

				<!-- Timeline Feed Items with vertical line -->
				<div
					class="space-y-4 relative before:absolute before:left-4 before:top-3 before:bottom-3 before:w-0.5 before:bg-white/10"
				>
					{#each Array(3) as _}
						<div class="relative pl-9 space-y-1">
							<!-- Node dot -->
							<div
								class="absolute left-2.5 top-3 -translate-x-1/2 size-3 rounded-full bg-card border-2 border-white/20 animate-pulse"
							></div>
							<!-- Message Box -->
							<div class="bg-foreground/10 border border-border rounded-lg p-3.5 space-y-2">
								<div class="flex items-center justify-between">
									<div class="flex items-center gap-2">
										<div class="size-5 rounded-full bg-white/10 animate-pulse"></div>
										<div class="h-3.5 w-24 bg-white/10 rounded animate-pulse"></div>
									</div>
									<div class="h-3 w-16 bg-white/5 rounded animate-pulse font-mono"></div>
								</div>
								<div class="space-y-1 pt-1">
									<div class="h-3.5 w-full bg-white/5 rounded animate-pulse"></div>
									<div class="h-3.5 w-4/5 bg-white/5 rounded animate-pulse"></div>
								</div>
							</div>
						</div>
					{/each}
				</div>
			</div>
		</div>

		<!-- Command Responders & PIC (25% / lg:col-span-1) -->
		<div class="lg:col-span-1">
			<div class="bg-card border border-border rounded-lg p-5 space-y-4">
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-2">
						<div class="size-4 rounded bg-white/10 animate-pulse"></div>
						<div class="h-5 w-28 bg-white/10 rounded animate-pulse"></div>
					</div>
					<div class="h-4 w-12 bg-white/5 rounded animate-pulse font-mono"></div>
				</div>

				<div class="space-y-3">
					{#each Array(3) as _}
						<div class="flex items-center gap-2.5 p-2 rounded-lg bg-white/3 border border-border/50">
							<div class="size-7 rounded-full bg-white/10 shrink-0 animate-pulse"></div>
							<div class="space-y-1 min-w-0 flex-1">
								<div class="h-3.5 w-24 bg-white/10 rounded animate-pulse"></div>
								<div class="h-3 w-16 bg-white/5 rounded animate-pulse"></div>
							</div>
						</div>
					{/each}
				</div>
			</div>
		</div>
	</div>

	<!-- 3. Post Status Update Form Skeleton -->
	<div class="bg-card border border-border rounded-lg p-5 space-y-3">
		<div class="h-4 w-40 bg-white/10 rounded animate-pulse"></div>
		<div class="h-24 w-full rounded-lg bg-white/4 border border-border animate-pulse"></div>
		<div class="flex justify-end">
			<div class="h-9 w-28 rounded-lg bg-white/10 animate-pulse"></div>
		</div>
	</div>
</div>
