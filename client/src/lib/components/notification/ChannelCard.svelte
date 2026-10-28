<script lang="ts">
	import { Mail, Webhook, Trash2, ToggleRight, ToggleLeft } from 'lucide-svelte';
	import type { NotificationChannel } from '$lib/api/notification';

	interface Props {
		channel: NotificationChannel;
		canManage: boolean;
		isToggling: boolean;
		onToggle: (ch: NotificationChannel) => void;
		onDelete: (ch: NotificationChannel) => void;
	}

	let { channel, canManage, isToggling, onToggle, onDelete }: Props = $props();

	function channelTypeLabel(type: string): string {
		return { email: 'Email', slack: 'Slack', webhook: 'Webhook', discord: 'Discord' }[type] ?? type;
	}

	function channelTypeBadge(type: string): string {
		return (
			{
				email: 'bg-blue-500/10 text-blue-400 border-blue-500/20',
				slack: 'bg-purple-500/10 text-purple-400 border-purple-500/20',
				webhook: 'bg-orange-500/10 text-orange-400 border-orange-500/20',
				discord: 'bg-indigo-500/10 text-indigo-400 border-indigo-500/20'
			}[type] ?? 'bg-white/5 text-foreground/60 border-border'
		);
	}

	const targetConfig = $derived.by(() => {
		if (!channel.config || typeof channel.config !== 'object') return '-';
		if (channel.config.webhook_url) return channel.config.webhook_url;
		if (channel.config.url) return channel.config.url;
		if (channel.config.email) return channel.config.email;
		if (channel.config.channel) return channel.config.channel;
		const entries = Object.entries(channel.config);
		if (entries.length === 0) return '-';
		return entries.map(([k, v]) => (entries.length === 1 ? v : `${k}: ${v}`)).join(', ');
	});
</script>

<div
	class="grid grid-cols-1 md:grid-cols-[minmax(180px,1.5fr)_110px_minmax(200px,2fr)_120px_80px] gap-4 px-5 py-2.5 items-center hover:bg-white/3 transition-colors"
>
	<!-- Column 1: Channel -->
	<div class="flex items-center gap-3 min-w-0">
		<div
			class="flex items-center justify-center size-8 rounded-md bg-white/5 text-muted-foreground shrink-0 border border-border/50"
		>
			{#if channel.type === 'email'}
				<Mail class="size-4 text-blue-400" />
			{:else if channel.type === 'slack'}
				<svg class="size-4" viewBox="0 0 122.8 122.8" xmlns="http://www.w3.org/2000/svg">
					<path
						d="M25.8 77.6c0 7.1-5.8 12.9-12.9 12.9S0 84.7 0 77.6s5.8-12.9 12.9-12.9h12.9v12.9zm6.5 0c0-7.1 5.8-12.9 12.9-12.9s12.9 5.8 12.9 12.9v32.3c0 7.1-5.8 12.9-12.9 12.9s-12.9-5.8-12.9-12.9V77.6z"
						fill="#e01e5a"
					/>
					<path
						d="M45.2 25.8c-7.1 0-12.9-5.8-12.9-12.9S38.1 0 45.2 0s12.9 5.8 12.9 12.9v12.9H45.2zm0 6.5c7.1 0 12.9 5.8 12.9 12.9s-5.8 12.9-12.9 12.9H12.9C5.8 58.1 0 52.3 0 45.2s5.8-12.9 12.9-12.9h32.3z"
						fill="#36c5f0"
					/>
					<path
						d="M97 45.2c0-7.1 5.8-12.9 12.9-12.9s12.9 5.8 12.9 12.9-5.8 12.9-12.9 12.9H97V45.2zm-6.5 0c0 7.1-5.8 12.9-12.9 12.9s-12.9-5.8-12.9-12.9V12.9C64.7 5.8 70.5 0 77.6 0s12.9 5.8 12.9 12.9v32.3z"
						fill="#2eb67d"
					/>
					<path
						d="M77.6 97c7.1 0 12.9 5.8 12.9 12.9s-5.8 12.9-12.9 12.9-12.9-5.8-12.9-12.9V97h12.9zm0-6.5c-7.1 0-12.9-5.8-12.9-12.9s5.8-12.9 12.9-12.9h32.3c7.1 0 12.9 5.8 12.9 12.9s-5.8 12.9-12.9 12.9H77.6z"
						fill="#ecb22e"
					/>
				</svg>
			{:else if channel.type === 'discord'}
				<svg class="size-4" viewBox="0 0 127.14 96.36" xmlns="http://www.w3.org/2000/svg">
					<path
						d="M107.7 8.07A105.15 105.15 0 0 0 81.47 0a72.06 72.06 0 0 0-3.36 6.83 97.68 97.68 0 0 0-29.17 0A72.37 72.37 0 0 0 45.57 0a105.89 105.89 0 0 0-26.25 8.09C2.79 32.65-1.71 56.6.54 80.21a105.73 105.73 0 0 0 32.17 16.15 77.7 77.7 0 0 0 6.89-11.11 68.42 68.42 0 0 1-10.85-5.18c.91-.66 1.8-1.34 2.66-2a75.57 75.57 0 0 0 64.32 0c.87.71 1.76 1.39 2.66 2a68.68 68.68 0 0 1-10.87 5.19 77 77 0 0 0 6.89 11.1 105.25 105.25 0 0 0 32.19-16.14c2.64-27.38-4.51-51.11-18.9-72.15ZM42.45 65.69C36.18 65.69 31 60 31 53s5-12.74 11.43-12.74S54 46 53.89 53s-5.05 12.69-11.44 12.69Zm42.24 0C78.41 65.69 73.25 60 73.25 53s5-12.74 11.44-12.74S96.23 46 96.12 53s-5.04 12.69-11.43 12.69Z"
						fill="#5865f2"
					/>
				</svg>
			{:else}
				<Webhook class="size-4 text-orange-400" />
			{/if}
		</div>
		<span class="font-medium text-foreground text-sm truncate" title={channel.name}>
			{channel.name}
		</span>
	</div>

	<!-- Column 2: Type -->
	<div class="flex items-center">
		<span
			class="inline-flex items-center text-[11px] font-medium px-2 py-0.5 rounded-full border w-max {channelTypeBadge(
				channel.type
			)}"
		>
			{channelTypeLabel(channel.type)}
		</span>
	</div>

	<!-- Column 3: Target / Config -->
	<div class="min-w-0">
		<span class="text-xs font-mono text-muted-foreground/80 truncate block" title={targetConfig}>
			{targetConfig}
		</span>
	</div>

	<!-- Column 4: Status -->
	<div class="flex items-center">
		{#if canManage}
			<button
				type="button"
				onclick={() => onToggle(channel)}
				disabled={isToggling}
				class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium transition-colors cursor-pointer disabled:opacity-50 {channel.is_active
					? 'bg-primary/10 text-primary hover:bg-primary/20'
					: 'bg-white/5 text-muted-foreground hover:bg-white/10'}"
			>
				{#if channel.is_active}
					<ToggleRight class="size-4" /> Enabled
				{:else}
					<ToggleLeft class="size-4" /> Disabled
				{/if}
			</button>
		{:else}
			<span class="text-xs text-muted-foreground/60 flex items-center gap-1.5">
				{#if channel.is_active}
					<span class="size-1.5 rounded-full bg-emerald-500"></span> Active
				{:else}
					<span class="size-1.5 rounded-full bg-muted-foreground/40"></span> Disabled
				{/if}
			</span>
		{/if}
	</div>

	<!-- Column 5: Actions -->
	<div class="flex items-center md:justify-end">
		{#if canManage}
			<button
				type="button"
				onclick={() => onDelete(channel)}
				class="p-1.5 rounded-lg text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors cursor-pointer"
				title="Delete channel"
				aria-label="Delete channel"
			>
				<Trash2 class="size-4" />
			</button>
		{/if}
	</div>
</div>
