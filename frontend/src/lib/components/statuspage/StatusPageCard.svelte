<script lang="ts">
	import { Globe, ExternalLink, Building2 } from 'lucide-svelte';
	import type { StatusPage } from '$lib/api/statuspage';

	let { sp }: { sp: StatusPage } = $props();
</script>

<div
	class="flex flex-col bg-background/80 border border-border rounded-xl p-5 hover:border-border transition-colors shadow-sm h-full w-full"
>
	<div class="flex items-start justify-between gap-4 mb-3">
		<div class="flex items-center gap-3 min-w-0">
			<div
				class="flex items-center justify-center size-10 bg-white/5 rounded-lg shrink-0 border border-border"
			>
				<Globe class="size-5 text-muted-foreground" />
			</div>
			<div class="min-w-0 flex flex-col">
				<h3 class="text-base font-semibold text-foreground truncate m-0">{sp.Title}</h3>
				<p class="text-sm text-muted-foreground truncate m-0">{sp.Subdomain}</p>
			</div>
		</div>
		{#if sp.IsPublic}
			<span
				class="text-[10px] font-mono px-2 py-0.5 bg-primary/10 text-primary border border-primary/20 rounded-full uppercase tracking-wider shrink-0"
				>Public</span
			>
		{:else}
			<span
				class="text-[10px] font-mono px-2 py-0.5 bg-white/10 text-muted-foreground border border-border rounded-full uppercase tracking-wider shrink-0"
				>Private</span
			>
		{/if}
	</div>

	<!-- Organization Name Display -->
	{#if sp.Organization}
		<div
			class="flex items-center gap-1.5 text-[12px] text-muted-foreground/80 mb-3 bg-white/5 w-fit px-2 py-1 rounded-md"
		>
			<Building2 class="size-3" />
			<span class="font-medium truncate max-w-50">{sp.Organization.Name}</span>
		</div>
	{/if}

	<p class="text-[13px] text-muted-foreground/80 line-clamp-2 flex-1 mb-4">
		{sp.Description || 'No description provided.'}
	</p>

	<div class="flex items-center justify-between mt-auto pt-4 border-t border-border">
		<div class="text-[12px] text-muted-foreground/60 flex items-center gap-1.5 font-medium">
			<span>{sp.StatusPageMonitors?.length || 0} monitors</span>
		</div>
		<a
			href={`/status/${sp.Subdomain}`}
			target="_blank"
			rel="noopener noreferrer"
			class="flex items-center gap-1.5 text-[12px] font-medium text-primary hover:text-primary/80 transition-colors"
		>
			View Page <ExternalLink class="size-3" />
		</a>
	</div>
</div>
