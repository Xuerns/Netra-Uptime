<script lang="ts">
	import { onMount } from 'svelte';
	import {
		Bell,
		Plus,
		Trash2,
		Mail,
		MessageSquare,
		Link2,
		ToggleLeft,
		ToggleRight,
		AlertCircle,
		CheckCircle,
		Clock,
		X,
		TriangleAlert,
		RefreshCw
	} from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';
	import {
		NotificationAPI,
		type NotificationChannel,
		type NotificationLog,
		type ChannelType
	} from '$lib/api/notification';
	import { toast } from 'svelte-sonner';

	let { orgId, userRole } = $props<{ orgId: string; userRole: string }>();

	type SubTab = 'channels' | 'history';
	let activeSubTab = $state<SubTab>('channels');

	// Channels state
	let channels = $state<NotificationChannel[]>([]);
	let isLoadingChannels = $state(true);
	let channelsError = $state('');

	// Add Modal state
	let isAddModalOpen = $state(false);
	let newName = $state('');
	let newType = $state<ChannelType>('email');
	let newEmail = $state('');
	let newSlackUrl = $state('');
	let newWebhookUrl = $state('');
	let newDiscordUrl = $state('');
	let isCreating = $state(false);
	let createError = $state('');

	// Toggle & Delete state
	let togglingId = $state<string | null>(null);
	let isDeleteConfirmOpen = $state(false);
	let channelToDelete = $state<NotificationChannel | null>(null);
	let isDeleting = $state(false);

	// History state
	let logs = $state<NotificationLog[]>([]);
	let logsTotal = $state(0);
	let isLoadingLogs = $state(false);
	let logsError = $state('');
	let logsOffset = $state(0);
	const logsLimit = 20;

	let canManage = $derived(userRole === 'owner' || userRole === 'admin');

	onMount(() => {
		loadChannels();
		loadLogs();
	});

	async function loadChannels() {
		isLoadingChannels = true;
		channelsError = '';
		try {
			channels = await NotificationAPI.getChannels(orgId);
		} catch (err: any) {
			channelsError = err.message || 'Failed to load notification channels';
		} finally {
			isLoadingChannels = false;
		}
	}

	async function loadLogs(reset = true) {
		if (reset) logsOffset = 0;
		isLoadingLogs = true;
		logsError = '';
		try {
			const res = await NotificationAPI.getLogs(orgId, logsLimit, logsOffset);
			logs = reset ? (res.logs ?? []) : [...logs, ...(res.logs ?? [])];
			logsTotal = res.total;
		} catch (err: any) {
			logsError = err.message || 'Failed to load notification logs';
		} finally {
			isLoadingLogs = false;
		}
	}

	async function handleCreateChannel(e: Event) {
		e.preventDefault();
		if (!newName.trim()) return;

		isCreating = true;
		createError = '';

		const config: Record<string, string> = {};
		if (newType === 'email') config.email = newEmail.trim();
		if (newType === 'slack') config.webhook_url = newSlackUrl.trim();
		if (newType === 'webhook') config.url = newWebhookUrl.trim();
		if (newType === 'discord') config.webhook_url = newDiscordUrl.trim();

		try {
			await NotificationAPI.createChannel(orgId, {
				name: newName.trim(),
				type: newType,
				config
			});
			toast.success('Notification channel created');
			isAddModalOpen = false;
			resetForm();
			await loadChannels();
		} catch (err: any) {
			createError = err.message || 'Failed to create channel';
		} finally {
			isCreating = false;
		}
	}

	async function handleToggleChannel(ch: NotificationChannel) {
		togglingId = ch.id;
		try {
			await NotificationAPI.updateChannel(orgId, ch.id, { is_active: !ch.is_active });
			toast.success(ch.is_active ? 'Channel disabled' : 'Channel enabled');
			await loadChannels();
		} catch (err: any) {
			toast.error(err.message || 'Failed to update channel');
		} finally {
			togglingId = null;
		}
	}

	function openDeleteConfirm(ch: NotificationChannel) {
		channelToDelete = ch;
		isDeleteConfirmOpen = true;
	}

	async function confirmDeleteChannel() {
		if (!channelToDelete) return;
		isDeleting = true;
		try {
			await NotificationAPI.deleteChannel(orgId, channelToDelete.id);
			toast.success(`Channel "${channelToDelete.name}" deleted`);
			await loadChannels();
		} catch (err: any) {
			toast.error(err.message || 'Failed to delete channel');
		} finally {
			isDeleting = false;
			channelToDelete = null;
		}
	}

	function resetForm() {
		newName = '';
		newType = 'email';
		newEmail = '';
		newSlackUrl = '';
		newWebhookUrl = '';
		newDiscordUrl = '';
		createError = '';
	}

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
			}[type] ?? 'bg-white/5 text-white/60'
		);
	}

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleString('en-GB', {
			day: '2-digit',
			month: 'short',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}
</script>

{#snippet SlackIcon(className = 'size-4')}
	<svg viewBox="0 0 24 24" class="{className} fill-current" xmlns="http://www.w3.org/2000/svg">
		<path
			d="M5.042 15.165a2.528 2.528 0 0 1-2.52 2.523A2.528 2.528 0 0 1 0 15.165a2.527 2.527 0 0 1 2.522-2.52h2.52v2.52zM6.313 15.165a2.527 2.527 0 0 1 2.521-2.52 2.527 2.527 0 0 1 2.521 2.52v6.313A2.528 2.528 0 0 1 8.834 24a2.528 2.528 0 0 1-2.521-2.522v-6.313zM8.834 5.042a2.528 2.528 0 0 1-2.521-2.52A2.528 2.528 0 0 1 8.834 0a2.528 2.528 0 0 1 2.521 2.522v2.52H8.834zM8.834 6.313a2.528 2.528 0 0 1 2.521 2.521 2.528 2.528 0 0 1-2.521 2.521H2.522A2.528 2.528 0 0 1 0 8.834a2.528 2.528 0 0 1 2.522-2.521h6.312zM18.956 8.834a2.528 2.528 0 0 1 2.522-2.521A2.528 2.528 0 0 1 24 8.834a2.528 2.528 0 0 1-2.522 2.521h-2.522V8.834zM17.688 8.834a2.528 2.528 0 0 1-2.523 2.521 2.527 2.527 0 0 1-2.52-2.521V2.522A2.527 2.527 0 0 1 15.165 0a2.528 2.528 0 0 1 2.523 2.522v6.312zM15.165 18.956a2.528 2.528 0 0 1 2.523 2.522A2.528 2.528 0 0 1 15.165 24a2.527 2.527 0 0 1-2.52-2.522v-2.522h2.52zM15.165 17.688a2.527 2.527 0 0 1-2.52-2.523 2.526 2.526 0 0 1 2.52-2.52h6.313A2.527 2.527 0 0 1 24 15.165a2.528 2.528 0 0 1-2.522 2.523h-6.313z"
		/>
	</svg>
{/snippet}

{#snippet DiscordIcon(className = 'size-4')}
	<svg viewBox="0 0 24 24" class="{className} fill-current" xmlns="http://www.w3.org/2000/svg">
		<path
			d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0 12.64 12.64 0 0 0-.617-1.25.077.077 0 0 0-.079-.037A19.736 19.736 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 0 0 .031.057 19.9 19.9 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028c.462-.63.874-1.295 1.226-1.994.021-.041.001-.09-.041-.106a13.107 13.107 0 0 1-1.872-.892.077.077 0 0 1-.008-.128 10.2 10.2 0 0 0 .372-.292.074.074 0 0 1 .077-.01c3.928 1.793 8.18 1.793 12.061 0a.074.074 0 0 1 .078.01c.12.098.246.198.373.292a.077.077 0 0 1-.006.127 12.299 12.299 0 0 1-1.873.892.077.077 0 0 0-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 0 0 .084.028 19.839 19.839 0 0 0 6.002-3.03.077.077 0 0 0 .032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.028zM8.02 15.33c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.956-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.956 2.418-2.157 2.418zm7.975 0c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.955-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.946 2.418-2.157 2.418z"
		/>
	</svg>
{/snippet}

<div class="flex flex-col gap-6 h-full pt-2">
	<!-- ── Sub-nav headers ────────────────────────────────────────────────── -->
	<div class="flex items-center justify-between border-b border-white/8 pb-3">
		<div class="flex items-center gap-2">
			<button
				class="px-3.5 py-1.5 rounded-lg text-xs font-medium transition-colors
					{activeSubTab === 'channels'
					? 'bg-white/10 text-white font-semibold'
					: 'text-on-surface-variant hover:text-white hover:bg-white/5'}"
				onclick={() => (activeSubTab = 'channels')}
			>
				Channels ({channels.length})
			</button>
			<button
				class="px-3.5 py-1.5 rounded-lg text-xs font-medium transition-colors
					{activeSubTab === 'history'
					? 'bg-white/10 text-white font-semibold'
					: 'text-on-surface-variant hover:text-white hover:bg-white/5'}"
				onclick={() => (activeSubTab = 'history')}
			>
				History Logs
			</button>
		</div>

		{#if activeSubTab === 'channels' && canManage}
			<Button
				class="bg-[#1A1A1B] border border-[#2A2A2B] text-primary hover:bg-primary hover:border-primary hover:text-white active:scale-[0.98] transition-all duration-200 h-9 px-4 text-xs font-medium shadow-sm"
				onclick={() => {
					resetForm();
					isAddModalOpen = true;
				}}
			>
				<Plus class="size-3.5 mr-1.5" />
				Add Channel
			</Button>
		{/if}
	</div>

	<!-- ╔══════════════════════════════════════════════════════════════╗
	     ║  SUB-TAB 1: CHANNELS                                        ║
	     ╚══════════════════════════════════════════════════════════════╝ -->
	{#if activeSubTab === 'channels'}
		{#if isLoadingChannels}
			<div class="flex items-center justify-center h-32">
				<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
			</div>
		{:else if channelsError}
			<div
				class="flex flex-col items-center justify-center gap-3 py-16 border border-white/8 rounded-lg bg-white/3"
			>
				<TriangleAlert class="size-8 text-error" />
				<p class="text-sm text-on-surface-variant">{channelsError}</p>
				<Button
					class="bg-[#1A1A1B] border border-[#2A2A2B] text-slate-300 hover:text-white hover:bg-[#2A2A2B] h-9 px-4 text-xs font-medium"
					onclick={loadChannels}
				>
					Try again
				</Button>
			</div>
		{:else if channels.length === 0}
			<div
				class="flex flex-col items-center justify-center text-center py-20 px-6 gap-3.5 bg-white/3 border border-white/8 rounded-lg backdrop-blur-md"
			>
				<div
					class="flex items-center justify-center size-[56px] rounded-lg bg-primary/8 text-primary"
				>
					<Bell class="size-6" />
				</div>
				<h2 class="text-[18px] font-semibold text-on-surface m-0">No notification channels yet</h2>
				<p class="text-sm leading-relaxed text-on-surface-variant max-w-[380px] m-0">
					{#if canManage}
						Add Email, Slack Incoming Webhooks, or Custom Webhooks to send monitor downtime alerts to
						your team.
					{:else}
						No notification channels configured for this organization.
					{/if}
				</p>
				{#if canManage}
					<Button
						class="mt-2 bg-[#1A1A1B] border border-[#2A2A2B] text-primary hover:bg-primary hover:border-primary hover:text-white h-10 px-5 text-xs font-medium"
						onclick={() => {
							resetForm();
							isAddModalOpen = true;
						}}
					>
						<Plus class="size-4 mr-1.5" />
						Add Channel
					</Button>
				{/if}
			</div>
		{:else}
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
				{#each channels as ch (ch.id)}
					<div
						class="bg-white/3 border border-white/8 rounded-xl p-5 flex flex-col justify-between gap-4 hover:border-white/14 transition-colors"
					>
						<div class="flex items-start justify-between gap-3">
							<div class="flex items-center gap-3 min-w-0">
								<div
									class="flex items-center justify-center size-9 rounded-lg bg-white/5 text-on-surface-variant shrink-0"
								>
									{#if ch.type === 'email'}
										<Mail class="size-4 text-blue-400" />
									{:else if ch.type === 'slack'}
										{@render SlackIcon('size-4 text-purple-400')}
									{:else if ch.type === 'discord'}
										{@render DiscordIcon('size-4 text-indigo-400')}
									{:else}
										<Link2 class="size-4 text-orange-400" />
									{/if}
								</div>

								<div class="flex flex-col min-w-0">
									<h3 class="font-medium text-white text-sm truncate">{ch.name}</h3>
									<span
										class="inline-block mt-0.5 text-[11px] font-medium px-2 py-0.5 rounded-full border border-white/10 w-max {channelTypeBadge(
											ch.type
										)}"
									>
										{channelTypeLabel(ch.type)}
									</span>
								</div>
							</div>

							{#if canManage}
								<button
									onclick={() => openDeleteConfirm(ch)}
									class="p-1.5 rounded-lg text-on-surface-variant hover:text-error hover:bg-error/10 transition-colors"
									title="Delete channel"
								>
									<Trash2 class="size-4" />
								</button>
							{/if}
						</div>

						<div class="space-y-1">
							{#each Object.entries(ch.config) as [k, v]}
								<p class="text-xs font-mono text-on-surface-variant/80 truncate">
									<span class="text-on-surface-variant/50">{k}:</span>
									{v}
								</p>
							{/each}
						</div>

						<div class="flex items-center justify-between pt-3 border-t border-white/8 text-xs">
							<span class="text-on-surface-variant/60">
								{ch.is_active ? 'Active' : 'Disabled'}
							</span>

							{#if canManage}
								<button
									onclick={() => handleToggleChannel(ch)}
									disabled={togglingId === ch.id}
									class="flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium transition-colors
										{ch.is_active
										? 'bg-primary/10 text-primary hover:bg-primary/20'
										: 'bg-white/5 text-on-surface-variant hover:bg-white/10'}"
								>
									{#if ch.is_active}
										<ToggleRight class="size-4" /> Enabled
									{:else}
										<ToggleLeft class="size-4" /> Disabled
									{/if}
								</button>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		{/if}
	{/if}

	<!-- ╔══════════════════════════════════════════════════════════════╗
	     ║  SUB-TAB 2: HISTORY LOGS                                    ║
	     ╚══════════════════════════════════════════════════════════════╝ -->
	{#if activeSubTab === 'history'}
		{#if isLoadingLogs && logs.length === 0}
			<div class="flex items-center justify-center h-32">
				<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
			</div>
		{:else if logsError}
			<div
				class="flex flex-col items-center justify-center gap-3 py-16 border border-white/8 rounded-lg bg-white/3"
			>
				<TriangleAlert class="size-8 text-error" />
				<p class="text-sm text-on-surface-variant">{logsError}</p>
				<Button
					class="bg-[#1A1A1B] border border-[#2A2A2B] text-slate-300 hover:text-white h-9 px-4 text-xs"
					onclick={() => loadLogs()}
				>
					Try again
				</Button>
			</div>
		{:else if logs.length === 0}
			<div
				class="flex flex-col items-center justify-center text-center py-20 px-6 gap-3.5 bg-white/3 border border-white/8 rounded-lg backdrop-blur-md"
			>
				<div
					class="flex items-center justify-center size-[56px] rounded-lg bg-white/5 text-on-surface-variant/60"
				>
					<Clock class="size-6" />
				</div>
				<h2 class="text-[18px] font-semibold text-on-surface m-0">No notification logs</h2>
				<p class="text-sm leading-relaxed text-on-surface-variant max-w-[380px] m-0">
					History will appear here when monitor alerts are triggered and dispatched to your
					channels.
				</p>
			</div>
		{:else}
			<div class="flex flex-col gap-2.5">
				{#each logs as log (log.id)}
					<div
						class="flex items-start gap-3 p-4 rounded-xl bg-white/3 border border-white/8 hover:border-white/14 transition-colors"
					>
						<div class="mt-0.5 shrink-0">
							{#if log.status === 'sent'}
								<CheckCircle class="size-4 text-primary" />
							{:else}
								<AlertCircle class="size-4 text-error" />
							{/if}
						</div>

						<div class="flex flex-col min-w-0 flex-1">
							<div class="flex items-center gap-2 flex-wrap">
								<span class="text-sm font-medium text-white truncate">{log.message}</span>
								<span
									class="text-[11px] font-medium px-2 py-0.5 rounded-full border {channelTypeBadge(
										log.channel
									)}"
								>
									{channelTypeLabel(log.channel)}
								</span>
							</div>

							{#if log.error_detail}
								<p class="text-xs text-error/80 mt-1 font-mono break-all bg-error/5 p-2 rounded">
									{log.error_detail}
								</p>
							{/if}

							<span class="text-xs text-on-surface-variant/60 mt-1">{formatDate(log.sent_at)}</span>
						</div>

						<span
							class="shrink-0 text-[10px] font-semibold px-2.5 py-0.5 rounded-full uppercase tracking-wider
							{log.status === 'sent' ? 'bg-primary/10 text-primary' : 'bg-error/10 text-error'}"
						>
							{log.status}
						</span>
					</div>
				{/each}
			</div>

			{#if logs.length < logsTotal}
				<Button
					onclick={async () => {
						logsOffset += logsLimit;
						await loadLogs(false);
					}}
					disabled={isLoadingLogs}
					class="self-center mt-3 bg-[#1A1A1B] border border-[#2A2A2B] text-slate-300 hover:text-white h-9 px-6 text-xs"
				>
					{isLoadingLogs ? 'Loading…' : 'Load more'}
				</Button>
			{/if}
		{/if}
	{/if}
</div>

<!-- ╔══════════════════════════════════════════════════════════════════════╗
     ║  ADD CHANNEL DIALOG                                                  ║
     ╚══════════════════════════════════════════════════════════════════════╝ -->
<Dialog.Root bind:open={isAddModalOpen}>
	<Dialog.Content
		class="sm:max-w-[520px] bg-[rgba(255,255,255,0.06)]! backdrop-blur-[20px] bg-gradient-to-b from-transparent to-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.08)] shadow-[0_0_20px_rgba(78,222,163,0.05)] text-white p-6 rounded-2xl"
	>
		<Dialog.Header>
			<Dialog.Title class="text-xl font-semibold tracking-tight">Add Notification Channel</Dialog.Title>
			<Dialog.Description class="text-[#bbcabf] text-sm mt-1">
				Configure a new destination for monitor status change alerts.
			</Dialog.Description>
		</Dialog.Header>

		<div class="h-px w-full bg-white/10 my-4"></div>

		{#if createError}
			<div
				class="p-3 text-[13px] text-[#690005] bg-[#ffb4ab] border border-[#93000a] rounded-lg font-medium flex items-center gap-2 mb-4"
			>
				<TriangleAlert class="size-4 shrink-0" />
				<p class="m-0">{createError}</p>
			</div>
		{/if}

		<form onsubmit={handleCreateChannel} class="flex flex-col gap-5">
			<!-- Name -->
			<div class="space-y-2">
				<Label for="ch-name" class="text-sm font-medium text-[#e5e2e3]"
					>Channel Name <span class="text-error font-bold">*</span></Label
				>
				<Input
					id="ch-name"
					bind:value={newName}
					placeholder="e.g. Team Slack Alerts, Ops Webhook"
					class="bg-[rgba(0,0,0,0.2)] border-white/10 text-white placeholder:text-white/30 focus-visible:ring-1 focus-visible:ring-[#3B82F6] focus-visible:border-[#3B82F6] rounded-lg h-10 text-sm"
					required
				/>
			</div>

			<!-- Type Selection -->
			<div class="space-y-2">
				<Label class="text-sm font-medium text-[#e5e2e3]">Channel Type <span class="text-error font-bold">*</span></Label>
				<div class="grid grid-cols-4 gap-2">
					{#each [
						{ id: 'email', label: 'Email' },
						{ id: 'slack', label: 'Slack' },
						{ id: 'discord', label: 'Discord' },
						{ id: 'webhook', label: 'Webhook' }
					] as t}
						<button
							type="button"
							onclick={() => (newType = t.id as ChannelType)}
							class="h-10 rounded-lg border text-xs font-medium transition-all flex items-center justify-center gap-1.5
								{newType === t.id
									? 'bg-primary/10 border-primary/40 text-primary font-semibold shadow-sm'
									: 'border-white/10 bg-[rgba(0,0,0,0.2)] text-white/70 hover:bg-white/5 hover:text-white'}"
						>
							{#if t.id === 'email'}
								<Mail class="size-3.5 text-blue-400" />
							{:else if t.id === 'slack'}
								{@render SlackIcon('size-3.5 text-purple-400')}
							{:else if t.id === 'discord'}
								{@render DiscordIcon('size-3.5 text-indigo-400')}
							{:else}
								<Link2 class="size-3.5 text-orange-400" />
							{/if}
							{t.label}
						</button>
					{/each}
				</div>
			</div>

			<!-- Dynamic Input -->
			{#if newType === 'email'}
				<div class="space-y-2">
					<Label for="ch-email" class="text-sm font-medium text-[#e5e2e3]"
						>Recipient Email <span class="text-error font-bold">*</span></Label
					>
					<Input
						id="ch-email"
						type="email"
						bind:value={newEmail}
						placeholder="alerts@company.com"
						class="bg-[rgba(0,0,0,0.2)] border-white/10 text-white placeholder:text-white/30 focus-visible:ring-1 focus-visible:ring-[#3B82F6] focus-visible:border-[#3B82F6] rounded-lg h-10 text-sm"
						required
					/>
				</div>
			{:else if newType === 'slack'}
				<div class="space-y-2">
					<Label for="ch-slack" class="text-sm font-medium text-[#e5e2e3]"
						>Slack Incoming Webhook URL <span class="text-error font-bold">*</span></Label
					>
					<Input
						id="ch-slack"
						type="url"
						bind:value={newSlackUrl}
						placeholder="https://hooks.slack.com/services/..."
						class="bg-[rgba(0,0,0,0.2)] border-white/10 text-white placeholder:text-white/30 focus-visible:ring-1 focus-visible:ring-[#3B82F6] focus-visible:border-[#3B82F6] rounded-lg h-10 text-sm font-mono"
						required
					/>
					<span class="text-xs text-white/40 block"
						>Must be an HTTPS Slack Webhook URL.</span
					>
				</div>
			{:else if newType === 'discord'}
				<div class="space-y-2">
					<Label for="ch-discord" class="text-sm font-medium text-[#e5e2e3]"
						>Discord Webhook URL <span class="text-error font-bold">*</span></Label
					>
					<Input
						id="ch-discord"
						type="url"
						bind:value={newDiscordUrl}
						placeholder="https://discord.com/api/webhooks/..."
						class="bg-[rgba(0,0,0,0.2)] border-white/10 text-white placeholder:text-white/30 focus-visible:ring-1 focus-visible:ring-[#3B82F6] focus-visible:border-[#3B82F6] rounded-lg h-10 text-sm font-mono"
						required
					/>
					<span class="text-xs text-white/40 block"
						>Must be an HTTPS Discord Webhook URL.</span
					>
				</div>
			{:else}
				<div class="space-y-2">
					<Label for="ch-webhook" class="text-sm font-medium text-[#e5e2e3]"
						>Webhook Destination URL <span class="text-error font-bold">*</span></Label
					>
					<Input
						id="ch-webhook"
						type="url"
						bind:value={newWebhookUrl}
						placeholder="https://api.yourdomain.com/webhooks/pulsewatch"
						class="bg-[rgba(0,0,0,0.2)] border-white/10 text-white placeholder:text-white/30 focus-visible:ring-1 focus-visible:ring-[#3B82F6] focus-visible:border-[#3B82F6] rounded-lg h-10 text-sm font-mono"
						required
					/>
					<span class="text-xs text-white/40 block"
						>Must be a public HTTPS URL. Internal IPs are blocked.</span
					>
				</div>
			{/if}

			<Dialog.Footer
				class="bg-transparent! m-0! p-0! pt-6 mt-2 flex justify-end gap-3 sm:space-x-0 border-t border-white/10"
			>
				<button
					type="button"
					onclick={() => (isAddModalOpen = false)}
					class="px-5 py-2.5 rounded-lg text-sm font-medium text-white hover:bg-white/10 transition-colors"
				>
					Cancel
				</button>
				<button
					type="submit"
					disabled={isCreating || !newName.trim()}
					class="h-11 px-6 rounded-lg text-sm font-medium bg-[#1A1A1B] border border-[#2A2A2B] text-primary hover:bg-primary hover:border-primary hover:text-white active:scale-[0.98] transition-all duration-200 shadow-sm disabled:opacity-50"
				>
					{isCreating ? 'Creating…' : 'Create Channel'}
				</button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>

<!-- ╔══════════════════════════════════════════════════════════════════════╗
     ║  DELETE CONFIRMATION DIALOG                                         ║
     ╚══════════════════════════════════════════════════════════════════════╝ -->
<ConfirmDialog
	bind:open={isDeleteConfirmOpen}
	title="Delete notification channel?"
	description="Are you sure you want to delete channel '{channelToDelete?.name}'? Alerts will no longer be sent to this destination."
	confirmText="Delete"
	cancelText="Cancel"
	destructive={true}
	onConfirm={confirmDeleteChannel}
/>
