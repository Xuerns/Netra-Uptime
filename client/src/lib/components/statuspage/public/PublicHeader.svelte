<script lang="ts">
	import { RefreshCw } from 'lucide-svelte';
	import logo from '$lib/assets/Logo.svg';
	import { formatTimeAgo } from '$lib/utils/datetime';
	import type { PublicStatusPage } from '$lib/api/statuspage';

	let {
		statusPage,
		themeColor,
		statusConfig,
		lastUpdated,
		secondsUntilRefresh,
		isRefreshing,
		onRefresh
	} = $props<{
		statusPage: PublicStatusPage;
		themeColor: string;
		statusConfig: any;
		lastUpdated: Date;
		secondsUntilRefresh: number;
		isRefreshing: boolean;
		onRefresh: () => void;
	}>();
</script>

<header
	class="sticky top-0 w-full z-50 bg-background/80 backdrop-blur-md border-b border-border/60 transition-colors"
>
	<div class="max-w-5xl mx-auto flex justify-between items-center px-4 sm:px-6 py-4 w-full">
		<div class="flex items-center gap-3">
			{#if statusPage.logo_url}
				<img
					class="h-8 max-w-36 object-contain rounded"
					src={statusPage.logo_url}
					alt={`${statusPage.title} Logo`}
				/>
			{:else}
				<img class="h-8 w-8 object-contain" src={logo} alt="Netra Uptime Logo" />
			{/if}
			<div class="flex flex-col">
				<div class="flex items-center gap-2">
					<span class="font-headline-md text-headline-md font-bold text-foreground">
						{statusPage.title}
					</span>
					<span
						style="border-color: color-mix(in srgb, {themeColor} 35%, transparent); color: {themeColor};"
						class="text-[11px] font-mono px-2 py-0.5 rounded-full bg-surface-dim border"
					>
						STATUS
					</span>
				</div>
				{#if statusPage.description}
					<span class="text-xs sm:text-sm text-muted-foreground font-body-md mt-0.5 line-clamp-1">
						{statusPage.description}
					</span>
				{/if}
			</div>
		</div>

		<!-- Live & Refresh Controls -->
		<div class="flex items-center gap-2.5">
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
				<span
					style={statusConfig.isThemed ? statusConfig.dotStyle : ''}
					class="size-2 rounded-full {statusConfig.dotClass}"
				></span>
				<span class="font-mono">{secondsUntilRefresh}s</span>
			</span>
		</div>
	</div>
</header>
