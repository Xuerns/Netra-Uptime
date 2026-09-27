<div class="space-y-6">
	<!-- 1. Header & Actions Banner Skeleton -->
	<div
		class="bg-card border border-border rounded-2xl p-5 md:p-6 shadow-xl flex flex-col md:flex-row md:items-center justify-between gap-4"
	>
		<div class="flex items-start gap-4 min-w-0 flex-1">
			<div class="size-12 rounded-xl bg-white/5 border border-border animate-pulse shrink-0"></div>
			<div class="space-y-2 min-w-0 flex-1">
				<div class="flex items-center gap-3 flex-wrap">
					<div class="h-7 w-48 bg-white/10 rounded-md animate-pulse"></div>
					<div class="h-6 w-20 rounded-md bg-white/6 animate-pulse"></div>
				</div>
				<div class="flex items-center gap-4 pt-1 flex-wrap text-xs">
					<div class="h-4 w-44 bg-white/5 rounded animate-pulse"></div>
					<div class="h-4 w-24 bg-white/5 rounded animate-pulse"></div>
					<div class="h-4 w-14 bg-white/5 rounded animate-pulse"></div>
					<div class="h-5 w-20 bg-white/5 rounded-md animate-pulse"></div>
					<div class="h-5 w-20 bg-white/5 rounded-md animate-pulse"></div>
				</div>
			</div>
		</div>

		<!-- Header Actions Skeleton -->
		<div class="flex items-center gap-2 shrink-0 self-start md:self-center">
			<div class="h-8 w-28 bg-white/10 rounded-lg animate-pulse"></div>
			<div class="h-8 w-20 bg-white/5 rounded-lg animate-pulse"></div>
			<div class="size-8 rounded-lg bg-white/5 animate-pulse"></div>
		</div>
	</div>

	<!-- 2. Time Range Selector Control Bar Skeleton -->
	<div
		class="flex items-center justify-between bg-card border border-border rounded-lg p-1.5 flex-wrap gap-2"
	>
		<div class="flex items-center gap-1.5">
			{#each ['24H', '7D', '30D', '90D', '1Y'] as _}
				<div class="h-7 w-12 rounded-md bg-white/6 animate-pulse"></div>
			{/each}
		</div>
		<div class="h-4 w-48 bg-white/5 rounded animate-pulse hidden sm:block"></div>
	</div>

	<!-- 3. Key Metrics Summary Grid Cards Skeleton (MonitorStatsGrid) -->
	<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
		{#each Array(4) as _}
			<div
				class="bg-card border border-border rounded-2xl p-5 space-y-3 flex flex-col justify-between shadow-lg h-[130px]"
			>
				<div class="flex justify-between items-center">
					<div class="h-3 w-24 bg-white/10 rounded animate-pulse"></div>
				</div>
				<div class="space-y-1.5">
					<div class="h-8 w-28 bg-white/10 rounded font-mono animate-pulse"></div>
					<div class="h-4 w-20 bg-white/5 rounded animate-pulse"></div>
				</div>
			</div>
		{/each}
	</div>

	<!-- 4. Interactive Response Time Latency Line Chart Skeleton -->
	<div class="bg-card border border-border rounded-2xl p-5 md:p-6 space-y-4 shadow-sm">
		<div class="flex justify-between items-center">
			<div class="h-5 w-56 bg-white/10 rounded animate-pulse"></div>
			<div class="h-4 w-32 bg-white/5 rounded animate-pulse"></div>
		</div>
		<div class="h-[260px] w-full rounded-xl bg-white/4 border border-border/80 animate-pulse"></div>
	</div>

	<!-- 5. Alert Rules & Incident Policies Skeleton -->
	<div class="bg-card border border-border rounded-2xl p-5 md:p-6 space-y-4">
		<div class="flex items-center justify-between">
			<div class="space-y-1">
				<div class="h-4.5 w-44 bg-white/10 rounded animate-pulse"></div>
				<div class="h-3 w-64 bg-white/5 rounded animate-pulse"></div>
			</div>
			<div class="h-8 w-28 bg-white/10 rounded-lg animate-pulse"></div>
		</div>

		<div class="space-y-2.5">
			{#each Array(2) as _}
				<div
					class="flex items-center justify-between p-4 rounded-xl bg-white/3 border border-border/80 animate-pulse"
				>
					<div class="space-y-1.5 min-w-0 flex-1">
						<div class="h-4 w-44 bg-white/10 rounded"></div>
						<div class="h-3 w-60 bg-white/5 rounded font-mono"></div>
					</div>
					<div class="flex items-center gap-3">
						<div class="h-6 w-11 rounded-full bg-white/10"></div>
						<div class="size-7 rounded-md bg-white/5"></div>
					</div>
				</div>
			{/each}
		</div>
	</div>

	<!-- 6. Recent Outages & Incidents Skeleton -->
	<div class="bg-card border border-border rounded-lg p-5 md:p-6 space-y-4">
		<div class="flex items-center justify-between">
			<div class="h-4 w-44 bg-white/10 rounded animate-pulse"></div>
			<div class="h-3 w-28 bg-white/5 rounded font-mono animate-pulse"></div>
		</div>
		<div class="p-4 rounded-lg bg-white/3 border border-border flex items-center justify-between">
			<div class="flex items-center gap-3">
				<div class="size-3 rounded-full bg-white/10 animate-pulse"></div>
				<div class="h-3.5 w-48 bg-white/10 rounded animate-pulse"></div>
			</div>
			<div class="h-3.5 w-24 bg-white/5 rounded animate-pulse font-mono"></div>
		</div>
	</div>
</div>
