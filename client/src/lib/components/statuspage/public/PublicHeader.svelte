<script lang="ts">
	import { RefreshCw } from 'lucide-svelte';
	import logo from '$lib/assets/Logo.svg';
	import { formatTimeAgo } from '$lib/utils/datetime';
	import type { PublicStatusPage } from '$lib/api/statuspage';

	let {
		statusPage,
		lastUpdated,
		secondsUntilRefresh,
		isRefreshing,
		onRefresh
	}: {
		statusPage: PublicStatusPage;
		lastUpdated: Date;
		secondsUntilRefresh: number;
		isRefreshing: boolean;
		onRefresh: () => void;
	} = $props();
</script>

<header
	class="sticky top-0 w-full z-50 bg-background/80 backdrop-blur-md border-b border-border/60 transition-colors"
>
	<div class="max-w-5xl mx-auto flex justify-between items-center px-4 sm:px-6 py-4 w-full">
		<div class="flex items-center gap-3 min-w-0 flex-1 mr-4">
			{#if statusPage.logo_url}
				<img
					class="h-8 max-w-36 object-contain rounded shrink-0"
					src={statusPage.logo_url}
					alt={`${statusPage.title} Logo`}
				/>
			{:else}
				<img class="h-8 w-8 object-contain shrink-0" src={logo} alt="Netra Uptime Logo" />
			{/if}
			<div class="flex flex-col min-w-0">
				<div class="flex items-center gap-2 min-w-0">
					<span
						class="font-headline-md text-headline-md font-bold text-foreground truncate"
						title={statusPage.title}
					>
						{statusPage.title}
					</span>
				</div>
				{#if statusPage.description}
					<span
						class="text-xs sm:text-sm text-muted-foreground font-body-md mt-0.5 break-words break-all [overflow-wrap:anywhere] line-clamp-2"
						title={statusPage.description}
					>
						{statusPage.description}
					</span>
				{/if}
			</div>
		</div>

		<!-- Live & Refresh Controls -->
		<div class="flex items-center gap-2.5 shrink-0">
			<button
				type="button"
				class="hidden sm:inline-flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground bg-white/5 hover:bg-white/10 border border-border/80 px-2.5 py-1 rounded-full transition-colors cursor-pointer"
				title="Click to refresh status immediately"
				onclick={onRefresh}
			>
				<RefreshCw
					class="size-3 text-muted-foreground {isRefreshing ? 'animate-spin text-primary' : ''}"
				/>
				<span
					>{isRefreshing
						? 'Updating...'
						: `Updated ${formatTimeAgo(lastUpdated.toISOString())}`}</span
				>
			</button>

			<span
				class="inline-flex items-center gap-1.5 text-xs text-muted-foreground bg-white/5 border border-border/80 px-2.5 py-1 rounded-full"
				title={`Auto-refresh in ${secondsUntilRefresh}s`}
			>
				<span class="font-mono">{secondsUntilRefresh}s</span>
			</span>
		</div>
	</div>
</header>
