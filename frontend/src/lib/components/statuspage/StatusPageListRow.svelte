<script lang="ts">
	import { Globe, ExternalLink, Building2 } from 'lucide-svelte';
	import type { StatusPage } from '$lib/api/statuspage';

	let { sp }: { sp: StatusPage } = $props();
</script>

<div class="list-row">
	<div class="flex flex-col min-w-0">
		<div class="flex items-center gap-3 min-w-0">
			<Globe class="size-4 text-on-surface-variant shrink-0" />
			<span class="text-sm font-medium text-on-surface truncate">{sp.Title}</span>
		</div>
		{#if sp.Organization}
			<div class="flex items-center gap-1 text-[11px] text-on-surface-variant/60 mt-1 pl-7">
				<Building2 class="size-3" />
				<span class="truncate max-w-37.5">{sp.Organization.Name}</span>
			</div>
		{/if}
	</div>
	<span class="text-[13px] text-on-surface-variant truncate">{sp.Subdomain}</span>
	<div>
		{#if sp.IsPublic}
			<span class="text-[10px] font-mono px-2 py-0.5 bg-primary/10 text-primary border border-primary/20 rounded-full uppercase tracking-wider">Public</span>
		{:else}
			<span class="text-[10px] font-mono px-2 py-0.5 bg-white/10 text-on-surface-variant border border-white/20 rounded-full uppercase tracking-wider">Private</span>
		{/if}
	</div>
	<span class="text-[13px] text-on-surface-variant">{sp.StatusPageMonitors?.length || 0} monitors</span>
	<div>
		<a 
			href={`/status/${sp.Subdomain}`} 
			target="_blank" 
			rel="noopener noreferrer"
			class="flex items-center justify-center size-7 rounded bg-white/5 border border-white/10 hover:bg-white/10 hover:text-primary transition-colors text-on-surface-variant"
			aria-label="View status page"
		>
			<ExternalLink class="size-3.5" />
		</a>
	</div>
</div>

<style>
	/* Shared from page */
	.list-row {
		display: grid;
		grid-template-columns: 1fr 1fr 100px 100px 60px;
		align-items: center;
		gap: 16px;
		padding: 12px 20px;
		transition: background 140ms ease;
	}

	.list-row:hover {
		background: rgba(255, 255, 255, 0.025);
	}

	@media (prefers-reduced-motion: reduce) {
		.list-row {
			transition: none;
		}
	}
</style>
