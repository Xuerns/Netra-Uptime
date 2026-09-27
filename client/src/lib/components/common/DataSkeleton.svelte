<script lang="ts">
	import { cn } from '$lib/helper/utils';

	type SkeletonVariant =
		| 'cards'
		| 'list'
		| 'table'
		| 'kpi'
		| 'detail'
		| 'form'
		| 'channel'
		| 'monitor';

	let {
		variant = 'list',
		viewMode = 'grid',
		count = 4,
		class: className = ''
	} = $props<{
		variant?: SkeletonVariant;
		viewMode?: 'grid' | 'list';
		count?: number;
		class?: string;
	}>();

	// Resolve variant if viewMode is specified on a list/grid component
	const effectiveVariant = $derived(
		variant === 'monitor'
			? 'monitor'
			: viewMode === 'list' && variant === 'cards'
				? 'list'
				: variant
	);
</script>

<div class={cn('w-full', className)} aria-busy="true" aria-live="polite">
	{#if effectiveVariant === 'monitor'}
		<!-- Monitor Page Skeleton: 4 KPI Cards + Grid/List Content -->
		<div class="space-y-6">
			<!-- 4 KPI Stats -->
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
				{#each Array(4) as _}
					<div class="p-4 rounded-xl bg-card border border-border flex flex-col justify-between h-[130px] animate-pulse">
						<div class="flex items-center justify-between">
							<div class="h-3.5 w-24 bg-white/10 rounded"></div>
							<div class="h-3.5 w-12 bg-white/5 rounded"></div>
						</div>
						<div class="space-y-1.5">
							<div class="h-7 w-28 bg-white/10 rounded font-mono"></div>
							<div class="h-3 w-36 bg-white/5 rounded"></div>
						</div>
						<div class="w-full bg-white/5 rounded-full h-1 overflow-hidden">
							<div class="h-full bg-white/10 w-4/5 rounded-full"></div>
						</div>
					</div>
				{/each}
			</div>

			<!-- Filter/Action Bar Placeholder -->
			<div class="flex items-center justify-between gap-4 py-1">
				<div class="flex items-center gap-2">
					<div class="h-8 w-16 bg-white/5 rounded-lg animate-pulse"></div>
					<div class="h-8 w-16 bg-white/5 rounded-lg animate-pulse"></div>
					<div class="h-8 w-20 bg-white/5 rounded-lg animate-pulse"></div>
				</div>
				<div class="h-8 w-20 bg-white/5 rounded-lg animate-pulse"></div>
			</div>

			<!-- Items Content -->
			{#if viewMode === 'grid'}
				<div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 min-[1400px]:grid-cols-4 gap-4">
					{#each Array(count || 6) as _}
						<div class="p-4 rounded-xl bg-card border border-border flex flex-col justify-between h-[210px] animate-pulse gap-3">
							<div class="flex items-start gap-3">
								<div class="size-8 rounded-lg bg-white/10 shrink-0"></div>
								<div class="space-y-1.5 flex-1 min-w-0">
									<div class="h-4 w-32 bg-white/10 rounded"></div>
									<div class="h-3 w-44 bg-white/5 rounded font-mono"></div>
								</div>
								<div class="h-5 w-14 rounded bg-white/10 shrink-0"></div>
							</div>
							<div class="h-8 bg-white/5 rounded-md my-1"></div>
							<div class="grid grid-cols-3 gap-2 pt-2 border-t border-border">
								<div class="h-6 bg-white/5 rounded"></div>
								<div class="h-6 bg-white/5 rounded"></div>
								<div class="h-6 bg-white/5 rounded"></div>
							</div>
						</div>
					{/each}
				</div>
			{:else}
				<div class="border border-border rounded-lg bg-card divide-y divide-border overflow-hidden">
					{#each Array(count || 6) as _}
						<div class="flex items-center justify-between p-3.5 gap-4 animate-pulse">
							<div class="flex items-center gap-3 flex-1 min-w-0">
								<div class="size-7 rounded-lg bg-white/10 shrink-0"></div>
								<div class="space-y-1 flex-1 min-w-0">
									<div class="h-3.5 w-36 bg-white/10 rounded"></div>
									<div class="h-2.5 w-48 bg-white/5 rounded font-mono"></div>
								</div>
							</div>
							<div class="h-5 w-14 rounded bg-white/10 shrink-0"></div>
							<div class="h-4 w-16 bg-white/5 rounded shrink-0 hidden sm:block"></div>
							<div class="h-4 w-16 bg-white/5 rounded shrink-0 hidden md:block"></div>
							<div class="size-7 rounded bg-white/5 shrink-0"></div>
						</div>
					{/each}
				</div>
			{/if}
		</div>

	{:else if effectiveVariant === 'cards'}
		<!-- Cards Grid Skeleton (used in Org, StatusPage, Monitors) -->
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
			{#each Array(count) as _}
				<div class="bg-card border border-border rounded-xl p-5 flex flex-col justify-between shadow-sm min-h-[190px] animate-pulse gap-3">
					<div class="flex items-start justify-between gap-3">
						<div class="flex items-center gap-3 min-w-0">
							<div class="size-9 rounded-lg bg-white/10 shrink-0"></div>
							<div class="space-y-1.5 flex-1 min-w-0">
								<div class="h-4 w-36 bg-white/10 rounded"></div>
								<div class="h-3 w-24 bg-white/5 rounded font-mono"></div>
							</div>
						</div>
						<div class="h-5 w-14 rounded-full bg-white/10 shrink-0"></div>
					</div>

					<div class="space-y-2 py-2">
						<div class="h-3 w-3/4 bg-white/5 rounded"></div>
						<div class="h-3 w-1/2 bg-white/5 rounded"></div>
					</div>

					<div class="flex items-center justify-between pt-3 border-t border-border">
						<div class="h-3.5 w-20 bg-white/5 rounded"></div>
						<div class="h-3.5 w-16 bg-white/10 rounded"></div>
					</div>
				</div>
			{/each}
		</div>

	{:else if effectiveVariant === 'list'}
		<!-- Standard List Rows Skeleton (used in Members, API Keys, Notifications) -->
		<div class="border border-border rounded-xl bg-card divide-y divide-border overflow-hidden">
			{#each Array(count) as _}
				<div class="flex items-center justify-between p-4 animate-pulse gap-4">
					<div class="flex items-center gap-3.5 min-w-0 flex-1">
						<div class="size-9 rounded-xl bg-white/10 shrink-0"></div>
						<div class="space-y-1.5 flex-1 min-w-0">
							<div class="h-4 w-40 bg-white/10 rounded"></div>
							<div class="h-3 w-28 bg-white/5 rounded font-mono"></div>
						</div>
					</div>
					<div class="flex items-center gap-3 shrink-0">
						<div class="h-6 w-16 rounded-md bg-white/10"></div>
						<div class="size-7 rounded-md bg-white/5"></div>
					</div>
				</div>
			{/each}
		</div>

	{:else if effectiveVariant === 'table'}
		<!-- Table Skeleton (used in Audit Logs) -->
		<div class="border border-border rounded-xl bg-card overflow-hidden">
			<div class="p-3.5 bg-muted/40 border-b border-border flex items-center justify-between gap-4">
				<div class="h-4 w-28 bg-white/10 rounded"></div>
				<div class="h-4 w-24 bg-white/10 rounded"></div>
				<div class="h-4 w-32 bg-white/10 rounded hidden sm:block"></div>
				<div class="h-4 w-20 bg-white/10 rounded"></div>
			</div>
			<div class="divide-y divide-border">
				{#each Array(count || 8) as _}
					<div class="p-4 flex items-center justify-between gap-4 animate-pulse">
						<div class="flex items-center gap-3 flex-1 min-w-0">
							<div class="size-7 rounded bg-white/10 shrink-0"></div>
							<div class="h-3.5 w-44 bg-white/10 rounded"></div>
						</div>
						<div class="h-3.5 w-24 bg-white/5 rounded shrink-0 font-mono"></div>
						<div class="h-3.5 w-32 bg-white/5 rounded shrink-0 hidden sm:block"></div>
						<div class="h-6 w-16 bg-white/10 rounded shrink-0"></div>
					</div>
				{/each}
			</div>
		</div>

	{:else if effectiveVariant === 'kpi'}
		<!-- Standalone KPI Grid -->
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
			{#each Array(count || 4) as _}
				<div class="p-4 rounded-xl bg-card border border-border flex flex-col justify-between h-[120px] animate-pulse">
					<div class="h-3.5 w-24 bg-white/10 rounded"></div>
					<div class="h-7 w-28 bg-white/10 rounded font-mono"></div>
					<div class="h-3 w-32 bg-white/5 rounded"></div>
				</div>
			{/each}
		</div>

	{:else if effectiveVariant === 'detail'}
		<!-- Detail Page Skeleton (Incident Detail, Monitor Detail) -->
		<div class="space-y-6 animate-pulse">
			<div class="bg-card border border-border rounded-2xl p-6 flex flex-col md:flex-row md:items-center justify-between gap-4">
				<div class="flex items-center gap-4 flex-1">
					<div class="size-12 rounded-xl bg-white/10 shrink-0"></div>
					<div class="space-y-2 flex-1">
						<div class="h-6 w-48 bg-white/10 rounded"></div>
						<div class="h-3.5 w-64 bg-white/5 rounded"></div>
					</div>
				</div>
				<div class="flex items-center gap-2">
					<div class="h-9 w-24 bg-white/10 rounded-lg"></div>
					<div class="size-9 bg-white/5 rounded-lg"></div>
				</div>
			</div>

			<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
				{#each Array(3) as _}
					<div class="p-4 rounded-xl bg-card border border-border h-24 flex flex-col justify-between">
						<div class="h-3.5 w-20 bg-white/10 rounded"></div>
						<div class="h-6 w-24 bg-white/10 rounded"></div>
					</div>
				{/each}
			</div>

			<div class="bg-card border border-border rounded-2xl p-6 h-[280px]">
				<div class="h-4 w-36 bg-white/10 rounded mb-4"></div>
				<div class="h-full bg-white/5 rounded-xl"></div>
			</div>
		</div>

	{:else if effectiveVariant === 'form'}
		<!-- Form Page Skeleton (Settings) -->
		<div class="space-y-6 animate-pulse">
			<div class="bg-card border border-border rounded-xl p-6 space-y-5">
				<div class="space-y-1.5 pb-4 border-b border-border">
					<div class="h-5 w-44 bg-white/10 rounded"></div>
					<div class="h-3.5 w-72 bg-white/5 rounded"></div>
				</div>
				<div class="space-y-4">
					<div class="space-y-2">
						<div class="h-3.5 w-28 bg-white/10 rounded"></div>
						<div class="h-10 w-full bg-white/5 rounded-lg border border-border"></div>
					</div>
					<div class="space-y-2">
						<div class="h-3.5 w-28 bg-white/10 rounded"></div>
						<div class="h-10 w-full bg-white/5 rounded-lg border border-border"></div>
					</div>
				</div>
			</div>
		</div>

	{:else if effectiveVariant === 'channel'}
		<!-- Notification Channel Skeleton -->
		<div class="flex flex-col gap-3">
			{#each Array(count) as _}
				<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-4.5 rounded-xl bg-card border border-border animate-pulse">
					<div class="flex items-center gap-3.5 flex-1 min-w-0">
						<div class="size-10 rounded-xl bg-white/10 shrink-0"></div>
						<div class="space-y-1.5 flex-1 min-w-0">
							<div class="h-4 w-36 bg-white/10 rounded"></div>
							<div class="h-3 w-48 bg-white/5 rounded font-mono"></div>
						</div>
					</div>
					<div class="flex items-center gap-3 shrink-0">
						<div class="h-6 w-11 rounded-full bg-white/10"></div>
						<div class="size-8 rounded-lg bg-white/5"></div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
