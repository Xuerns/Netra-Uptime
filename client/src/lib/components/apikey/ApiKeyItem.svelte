<script lang="ts">
	import { Trash2, Clock } from 'lucide-svelte';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import type { APIKey } from '$lib/api/apikey';
	import { formatDateOnly } from '$lib/utils/datetime';

	interface Props {
		apiKey: APIKey;
		canManage: boolean;
		timezone: string;
		onRevoke: (key: APIKey) => void;
	}

	let { apiKey, canManage, timezone, onRevoke }: Props = $props();

	const formattedExpiresAt = $derived(formatDateOnly(apiKey.expires_at, timezone));
</script>

<div
	class="p-4 rounded-lg bg-card border border-border flex items-center justify-between gap-4 flex-wrap transition-colors hover:border-primary/30"
>
	<div class="flex flex-col gap-1 min-w-0">
		<div class="flex items-center gap-2.5 flex-wrap">
			<span class="text-sm font-semibold text-foreground truncate">{apiKey.name}</span>

			<!-- Scope Badge -->
			<Badge
				variant="outline"
				class="text-[11px] font-mono px-2 py-0.5 font-medium {apiKey.permissions === 'write'
					? 'bg-amber-500/10 border-amber-500/20 text-amber-400'
					: 'bg-blue-500/10 border-blue-500/20 text-blue-400'}"
			>
				{apiKey.permissions === 'write' ? 'Full Access (Read & Write)' : 'Read Only'}
			</Badge>
		</div>

		<div class="flex items-center gap-4 text-[11.5px] font-mono text-muted-foreground flex-wrap">
			<span
				class="bg-white/5 px-2 py-0.5 rounded border border-border text-foreground font-semibold"
			>
				{apiKey.key_prefix}...****
			</span>
			{#if apiKey.created_by}
				<span class="flex items-center gap-1">
					Created by <strong class="text-muted-foreground">{apiKey.created_by}</strong>
				</span>
			{/if}
			<span class="flex items-center gap-1">
				<Clock class="size-3 shrink-0 opacity-60" /> Expires: {formattedExpiresAt}
			</span>
		</div>
	</div>

	<div class="flex items-center gap-3 shrink-0">
		<div class="text-right text-[11px] text-muted-foreground/60 hidden sm:block">
			{#if apiKey.last_used_at}
				Last used {new Date(apiKey.last_used_at).toLocaleDateString()}
				{#if apiKey.last_used_ip}
					({apiKey.last_used_ip}){/if}
			{:else}
				Never used
			{/if}
		</div>

		{#if canManage}
			<button
				class="p-2 rounded-lg text-muted-foreground hover:text-rose-400 hover:bg-rose-500/10 transition-colors cursor-pointer"
				title="Revoke API key"
				onclick={() => onRevoke(apiKey)}
			>
				<Trash2 class="size-4" />
			</button>
		{/if}
	</div>
</div>
