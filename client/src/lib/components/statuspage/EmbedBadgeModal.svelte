<script lang="ts">
	import type { StatusPage } from '$lib/api/statuspage';
	import * as Dialog from '$lib/components/ui/dialog';
	import { toast } from 'svelte-sonner';
	import { Copy, Check, ExternalLink, Code2, FileJson, Sparkles } from 'lucide-svelte';
	import { env } from '$env/dynamic/public';

	let {
		isOpen = $bindable(false),
		statusPage
	}: {
		isOpen: boolean;
		statusPage: StatusPage | null;
	} = $props();

	let copiedType = $state<string | null>(null);

	// Get base API URL or origin
	let origin = $derived.by(() => {
		if (typeof window !== 'undefined') {
			return window.location.origin;
		}
		return env.PUBLIC_API_URL || 'http://localhost:27346';
	});

	let badgeUrl = $derived(
		statusPage ? `${origin}/api/v1/public/statuspage/${statusPage.Subdomain}/badge.svg` : ''
	);

	let pageUrl = $derived(statusPage ? `${origin}/status/${statusPage.Subdomain}` : '');

	let statusJsonUrl = $derived(
		statusPage ? `${origin}/api/v1/public/statuspage/${statusPage.Subdomain}/status.json` : ''
	);

	let summaryJsonUrl = $derived(
		statusPage ? `${origin}/api/v1/public/statuspage/${statusPage.Subdomain}/summary.json` : ''
	);

	let markdownCode = $derived(
		statusPage ? `[![${statusPage.Title} Status](${badgeUrl})](${pageUrl})` : ''
	);

	let htmlCode = $derived(
		statusPage
			? `<a href="${pageUrl}" target="_blank" rel="noopener noreferrer">\n  <img src="${badgeUrl}" alt="${statusPage.Title} Status" />\n</a>`
			: ''
	);

	async function copyToClipboard(text: string, type: string) {
		try {
			await navigator.clipboard.writeText(text);
			copiedType = type;
			toast.success('Copied to clipboard!');
			setTimeout(() => {
				if (copiedType === type) copiedType = null;
			}, 2000);
		} catch {
			toast.error('Failed to copy to clipboard.');
		}
	}
</script>

<Dialog.Root bind:open={isOpen}>
	<Dialog.Content
		class="sm:max-w-[600px] max-h-[90vh] flex flex-col bg-card border border-border rounded-2xl p-0 overflow-hidden shadow-2xl"
	>
		<!-- Header -->
		<div class="p-6 border-b border-border/60 flex items-center justify-between">
			<div class="flex items-center gap-3">
				<div>
					<Dialog.Title class="text-lg font-bold text-foreground flex items-center gap-2">
						Embed Status Badge
					</Dialog.Title>
					<Dialog.Description class="text-xs text-muted-foreground mt-0.5">
						Display realtime operational status of <span class="text-foreground font-medium"
							>{statusPage?.Title}</span
						> on GitHub, websites, or documentation.
					</Dialog.Description>
				</div>
			</div>
		</div>

		<!-- Body -->
		<div class="flex-1 overflow-y-auto p-6 space-y-6">
			<!-- Live Badge Preview -->
			<div
				class="p-5 rounded-xl border border-border/70 bg-white/3 flex flex-col items-center justify-center gap-3 text-center"
			>
				{#if statusPage}
					<div
						class="p-3 bg-black/40 rounded-lg border border-border/40 inline-flex items-center justify-center"
					>
						<img src={badgeUrl} alt="Live Status Badge" class="h-5" />
					</div>
					<p class="text-[11px] text-muted-foreground">
						Badge updates automatically in real-time as monitor statuses change.
					</p>
				{/if}
			</div>

			<!-- Embed Snippets -->
			<div class="space-y-4">
				<!-- Markdown -->
				<div class="space-y-1.5">
					<div class="flex items-center justify-between">
						<label class="text-xs font-semibold text-foreground flex items-center gap-1.5">
							<Code2 class="size-3.5 text-primary" />
							Markdown (for GitHub README)
						</label>
						<button
							type="button"
							class="text-[11px] text-primary hover:underline flex items-center gap-1 cursor-pointer"
							onclick={() => copyToClipboard(markdownCode, 'markdown')}
						>
							{#if copiedType === 'markdown'}
								<Check class="size-3 text-emerald-400" />
								<span>Copied</span>
							{:else}
								<Copy class="size-3" />
								<span>Copy Markdown</span>
							{/if}
						</button>
					</div>
					<pre
						class="p-3 bg-black/50 border border-border/80 rounded-lg text-xs font-mono text-foreground/90 overflow-x-auto select-all">{markdownCode}</pre>
				</div>

				<!-- HTML -->
				<div class="space-y-1.5">
					<div class="flex items-center justify-between">
						<label class="text-xs font-semibold text-foreground flex items-center gap-1.5">
							<Code2 class="size-3.5 text-primary" />
							HTML (for Websites & Portals)
						</label>
						<button
							type="button"
							class="text-[11px] text-primary hover:underline flex items-center gap-1 cursor-pointer"
							onclick={() => copyToClipboard(htmlCode, 'html')}
						>
							{#if copiedType === 'html'}
								<Check class="size-3 text-emerald-400" />
								<span>Copied</span>
							{:else}
								<Copy class="size-3" />
								<span>Copy HTML</span>
							{/if}
						</button>
					</div>
					<pre
						class="p-3 bg-black/50 border border-border/80 rounded-lg text-xs font-mono text-foreground/90 overflow-x-auto select-all whitespace-pre-wrap">{htmlCode}</pre>
				</div>

				<!-- Direct Badge Image URL -->
				<div class="space-y-1.5">
					<div class="flex items-center justify-between">
						<label class="text-xs font-semibold text-foreground flex items-center gap-1.5">
							<ExternalLink class="size-3.5 text-primary" />
							Direct SVG Badge URL
						</label>
						<button
							type="button"
							class="text-[11px] text-primary hover:underline flex items-center gap-1 cursor-pointer"
							onclick={() => copyToClipboard(badgeUrl, 'badgeUrl')}
						>
							{#if copiedType === 'badgeUrl'}
								<Check class="size-3 text-emerald-400" />
								<span>Copied</span>
							{:else}
								<Copy class="size-3" />
								<span>Copy URL</span>
							{/if}
						</button>
					</div>
					<input
						readonly
						value={badgeUrl}
						class="w-full px-3 py-2 bg-black/50 border border-border/80 rounded-lg text-xs font-mono text-foreground/90 select-all outline-none"
					/>
				</div>
			</div>

			<!-- Public JSON API Links -->
			<div class="p-4 rounded-xl border border-border/70 bg-white/2 space-y-2">
				<div class="flex items-center gap-1.5 text-xs font-semibold text-foreground">
					<FileJson class="size-4 text-primary" />
					<span>Public Status JSON Endpoints</span>
				</div>
				<p class="text-[11px] text-muted-foreground leading-relaxed">
					Use these unauthenticated endpoints for bots, CLI tools, or health check integrations:
				</p>
				<div class="space-y-1.5 pt-1">
					<div
						class="flex items-center justify-between text-xs font-mono bg-black/40 p-2 rounded border border-border/50"
					>
						<span class="text-foreground/80 truncate pr-2">GET .../status.json</span>
						<a
							href={statusJsonUrl}
							target="_blank"
							rel="noopener noreferrer"
							class="text-primary hover:text-primary/80 flex items-center gap-1 shrink-0 text-[11px]"
						>
							Open <ExternalLink class="size-3" />
						</a>
					</div>
					<div
						class="flex items-center justify-between text-xs font-mono bg-black/40 p-2 rounded border border-border/50"
					>
						<span class="text-foreground/80 truncate pr-2">GET .../summary.json</span>
						<a
							href={summaryJsonUrl}
							target="_blank"
							rel="noopener noreferrer"
							class="text-primary hover:text-primary/80 flex items-center gap-1 shrink-0 text-[11px]"
						>
							Open <ExternalLink class="size-3" />
						</a>
					</div>
				</div>
			</div>
		</div>
	</Dialog.Content>
</Dialog.Root>
