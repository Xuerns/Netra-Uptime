<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { headerStore } from '$lib/stores/headerStore.svelte';
	import {
		Bell,
		Plus,
		Trash2,
		Mail,
		Webhook,
		Link2,
		ToggleLeft,
		ToggleRight,
		AlertCircle,
		CheckCircle,
		Clock,
		X,
		TriangleAlert
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
	import { userTimezone } from '$lib/stores/timezone';
	import { formatDateTime } from '$lib/utils/datetime';
	import { parseApiError, truncateErrorDetail } from '$lib/utils/parseApiError';
	import { DataSkeleton } from '$lib/components/common';

	const orgId = $derived($page.params.id as string);

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

	// Error detail expansion state
	let expandedErrors = $state<Record<string, boolean>>({});

	function toggleErrorDetail(id: string) {
		expandedErrors[id] = !expandedErrors[id];
	}

	const currentUserId = $derived($page.data?.user?.id || $page.data?.user?.ID);
	const currentMember = $derived(
		($page.data?.members ?? []).find(
			(m: any) => (m.user_id ?? m.UserID ?? m.id ?? m.ID) === currentUserId
		)
	);
	const userRole = $derived(currentMember?.Role ?? currentMember?.role ?? 'admin');
	const canManage = $derived(userRole === 'owner' || userRole === 'admin');

	onMount(() => {
		if (orgId) {
			loadChannels();
			loadLogs();
		}
	});

	async function loadChannels() {
		if (!orgId) return;
		isLoadingChannels = true;
		channelsError = '';
		try {
			channels = await NotificationAPI.getChannels(orgId);
		} catch (err: any) {
			channelsError = parseApiError(err, 'Failed to load notification channels');
		} finally {
			isLoadingChannels = false;
		}
	}

	async function loadLogs(reset = true) {
		if (!orgId) return;
		if (reset) logsOffset = 0;
		isLoadingLogs = true;
		logsError = '';
		try {
			const res = await NotificationAPI.getLogs(orgId, logsLimit, logsOffset);
			logs = reset ? (res.logs ?? []) : [...logs, ...(res.logs ?? [])];
			logsTotal = res.total;
		} catch (err: any) {
			logsError = parseApiError(err, 'Failed to load notification logs');
		} finally {
			isLoadingLogs = false;
		}
	}

	async function handleCreateChannel(e: Event) {
		e.preventDefault();
		if (!newName.trim() || !orgId) return;

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
			createError = parseApiError(err, 'Failed to create channel');
		} finally {
			isCreating = false;
		}
	}

	async function handleToggleChannel(ch: NotificationChannel) {
		if (!orgId) return;
		togglingId = ch.id;
		try {
			await NotificationAPI.updateChannel(orgId, ch.id, { is_active: !ch.is_active });
			toast.success(ch.is_active ? 'Channel disabled' : 'Channel enabled');
			await loadChannels();
		} catch (err: any) {
			toast.error(parseApiError(err, 'Failed to update channel'));
		} finally {
			togglingId = null;
		}
	}

	function openDeleteConfirm(ch: NotificationChannel) {
		channelToDelete = ch;
		isDeleteConfirmOpen = true;
	}

	async function confirmDeleteChannel() {
		if (!channelToDelete || !orgId) return;
		isDeleting = true;
		try {
			await NotificationAPI.deleteChannel(orgId, channelToDelete.id);
			toast.success(`Channel "${channelToDelete.name}" deleted`);
			await loadChannels();
		} catch (err: any) {
			toast.error(parseApiError(err, 'Failed to delete channel'));
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
			}[type] ?? 'bg-white/5 text-foreground/60'
		);
	}

	let tz = $derived($userTimezone);
	const formatDate = (iso: string) => formatDateTime(iso, tz);

	// Sync with layout top bar
	$effect(() => {
		headerStore.set({
			actionType: activeSubTab === 'channels' && canManage ? 'button' : null,
			actionLabel: 'Add Channel',
			onAction: () => {
				resetForm();
				isAddModalOpen = true;
			},
			badge: {
				text: `${channels.length} Channel${channels.length !== 1 ? 's' : ''}`,
				variant: 'slate'
			}
		});
	});


</script>

<svelte:head>
	<title>Notification Channels | PulseWatch</title>
</svelte:head>

<div class="text-foreground antialiased">
	<!-- Page Header -->
	<header class="flex flex-wrap items-start justify-between gap-4 mb-6">
		<div>
			<h1 class="text-[32px] font-semibold leading-tight tracking-tight text-foreground mb-1.5">
				Notifications
			</h1>
			<p class="text-base leading-relaxed text-muted-foreground max-w-[50ch]">
				Configure channels to receive real-time alerts when your monitors experience downtime or
				recover.
			</p>
		</div>
	</header>

	<!-- Sub-nav Tabs -->
	<div class="flex items-center justify-between border-b border-border pb-3 mb-6">
		<div class="flex items-center gap-2">
			<button
				class="px-3.5 py-1.5 rounded-lg text-xs font-medium transition-colors {activeSubTab ===
				'channels'
					? 'bg-border text-foreground font-semibold'
					: 'text-muted-foreground hover:text-foreground hover:bg-white/5'}"
				onclick={() => (activeSubTab = 'channels')}
			>
				Channels ({channels.length})
			</button>
			<button
				class="px-3.5 py-1.5 rounded-lg text-xs font-medium transition-colors {activeSubTab ===
				'history'
					? 'bg-border text-foreground font-semibold'
					: 'text-muted-foreground hover:text-foreground hover:bg-white/5'}"
				onclick={() => (activeSubTab = 'history')}
			>
				History Logs
			</button>
		</div>
	</div>

	<!-- TAB 1: CHANNELS -->
	{#if activeSubTab === 'channels'}
		{#if isLoadingChannels}
			<DataSkeleton variant="channel" count={4} />
		{:else if channelsError}
			<div
				class="flex flex-col items-center justify-center gap-3 py-16 border border-border rounded-lg bg-white/3"
			>
				<TriangleAlert class="size-8 text-destructive" />
				<p class="text-sm text-muted-foreground">{channelsError}</p>
				<Button
					class="bg-secondary border border-border text-muted-foreground hover:text-foreground hover:bg-accent h-9 px-4 text-xs font-medium"
					onclick={loadChannels}
				>
					Try again
				</Button>
			</div>
		{:else if channels.length === 0}
			<div
				class="flex flex-col items-center justify-center text-center py-20 px-6 gap-3.5 bg-white/3 border border-border rounded-lg backdrop-blur-md"
			>
				<div
					class="flex items-center justify-center size-[56px] rounded-lg bg-primary/8 text-primary"
				>
					<Bell class="size-6" />
				</div>
				<h2 class="text-[18px] font-semibold text-foreground m-0">No notification channels yet</h2>
				<p class="text-sm leading-relaxed text-muted-foreground max-w-[380px] m-0">
					{#if canManage}
						Add Email, Slack Incoming Webhooks, Discord, or Custom Webhooks to send monitor downtime
						alerts to your team.
					{:else}
						No notification channels configured for this organization.
					{/if}
				</p>
				{#if canManage}
					<Button
						class="mt-2 bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground h-10 px-5 text-xs font-medium"
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
						class="bg-white/3 border border-border rounded-lg p-5 flex flex-col justify-between gap-4 hover:border-primary/30 transition-colors"
					>
						<div class="flex items-start justify-between gap-3">
							<div class="flex items-center gap-3 min-w-0">
								<div
									class="flex items-center justify-center size-9 rounded-md bg-white/5 text-muted-foreground shrink-0"
								>
									{#if ch.type === 'email'}
										<Mail class="size-4 text-blue-400" />
									{:else if ch.type === 'slack'}
										<svg
											class="size-4"
											viewBox="0 0 122.8 122.8"
											xmlns="http://www.w3.org/2000/svg"
										>
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
									{:else if ch.type === 'discord'}
										<svg
											class="size-4"
											viewBox="0 0 127.14 96.36"
											xmlns="http://www.w3.org/2000/svg"
										>
											<path
												d="M107.7 8.07A105.15 105.15 0 0 0 81.47 0a72.06 72.06 0 0 0-3.36 6.83 97.68 97.68 0 0 0-29.17 0A72.37 72.37 0 0 0 45.57 0a105.89 105.89 0 0 0-26.25 8.09C2.79 32.65-1.71 56.6.54 80.21a105.73 105.73 0 0 0 32.17 16.15 77.7 77.7 0 0 0 6.89-11.11 68.42 68.42 0 0 1-10.85-5.18c.91-.66 1.8-1.34 2.66-2a75.57 75.57 0 0 0 64.32 0c.87.71 1.76 1.39 2.66 2a68.68 68.68 0 0 1-10.87 5.19 77 77 0 0 0 6.89 11.1 105.25 105.25 0 0 0 32.19-16.14c2.64-27.38-4.51-51.11-18.9-72.15ZM42.45 65.69C36.18 65.69 31 60 31 53s5-12.74 11.43-12.74S54 46 53.89 53s-5.05 12.69-11.44 12.69Zm42.24 0C78.41 65.69 73.25 60 73.25 53s5-12.74 11.44-12.74S96.23 46 96.12 53s-5.04 12.69-11.43 12.69Z"
												fill="#5865f2"
											/>
										</svg>
									{:else}
										<Webhook class="size-4 text-orange-400" />
									{/if}
								</div>

								<div class="flex flex-col min-w-0">
									<h3 class="font-medium text-foreground text-sm truncate">{ch.name}</h3>
									<span
										class="inline-block mt-0.5 text-[11px] font-medium px-2 py-0.5 rounded-full border border-border w-max {channelTypeBadge(
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
									class="p-1.5 rounded-lg text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors"
									title="Delete channel"
								>
									<Trash2 class="size-4" />
								</button>
							{/if}
						</div>

						<div class="space-y-1">
							{#each Object.entries(ch.config) as [k, v]}
								<p class="text-xs font-mono text-muted-foreground/80 truncate">
									<span class="text-muted-foreground/50">{k}:</span>
									{v}
								</p>
							{/each}
						</div>

						<div class="flex items-center justify-between pt-3 border-t border-border text-xs">
							<span class="text-muted-foreground/60">
								{ch.is_active ? 'Active' : 'Disabled'}
							</span>

							{#if canManage}
								<button
									onclick={() => handleToggleChannel(ch)}
									disabled={togglingId === ch.id}
									class="flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium transition-colors {ch.is_active
										? 'bg-primary/10 text-primary hover:bg-primary/20'
										: 'bg-white/5 text-muted-foreground hover:bg-accent'}"
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

	<!-- TAB 2: HISTORY LOGS -->
	{#if activeSubTab === 'history'}
		{#if isLoadingLogs && logs.length === 0}
			<DataSkeleton variant="list" count={4} />
		{:else if logsError}
			<div
				class="flex flex-col items-center justify-center gap-3 py-16 border border-border rounded-lg bg-white/3"
			>
				<TriangleAlert class="size-8 text-destructive" />
				<p class="text-sm text-muted-foreground">{logsError}</p>
				<Button
					class="bg-secondary border border-border text-muted-foreground hover:text-foreground h-9 px-4 text-xs"
					onclick={() => loadLogs()}
				>
					Try again
				</Button>
			</div>
		{:else if logs.length === 0}
			<div
				class="flex flex-col items-center justify-center text-center py-20 px-6 gap-3.5 bg-white/3 border border-border rounded-lg backdrop-blur-md"
			>
				<div
					class="flex items-center justify-center size-[56px] rounded-lg bg-white/5 text-muted-foreground/60"
				>
					<Clock class="size-6" />
				</div>
				<h2 class="text-[18px] font-semibold text-foreground m-0">No notification logs</h2>
				<p class="text-sm leading-relaxed text-muted-foreground max-w-[380px] m-0">
					History will appear here when monitor alerts are triggered and dispatched to your
					channels.
				</p>
			</div>
		{:else}
			<div class="flex flex-col gap-2.5">
				{#each logs as log (log.id)}
					<div
						class="flex items-start gap-3 p-4 rounded-lg bg-white/3 border border-border hover:border-border transition-colors"
					>
						<div class="mt-0.5 shrink-0">
							{#if log.status === 'sent'}
								<CheckCircle class="size-4 text-primary" />
							{:else}
								<AlertCircle class="size-4 text-destructive" />
							{/if}
						</div>

						<div class="flex flex-col min-w-0 flex-1">
							<div class="flex items-center gap-2 flex-wrap">
								<span class="text-sm font-medium text-foreground truncate">{log.message}</span>
								<span
									class="text-[11px] font-medium px-2 py-0.5 rounded-full border {channelTypeBadge(
										log.channel
									)}"
								>
									{channelTypeLabel(log.channel)}
								</span>
							</div>

							{#if log.error_detail}
								<div
									class="mt-2 rounded-lg border border-destructive/20 bg-destructive/5 px-3 py-2 text-xs"
								>
									<div class="flex items-center justify-between gap-2">
										<span class="font-medium text-destructive flex items-center gap-1.5">
											<AlertCircle class="size-3.5 shrink-0" />
											Error Details
										</span>
										{#if log.error_detail.length > 120}
											<button
												type="button"
												onclick={() => toggleErrorDetail(log.id)}
												class="text-[11px] text-muted-foreground hover:text-foreground underline transition-colors cursor-pointer"
											>
												{expandedErrors[log.id] ? 'Show less' : 'Show details'}
											</button>
										{/if}
									</div>
									<p
										class="mt-1 font-mono text-[11px] leading-relaxed text-foreground/80 break-words whitespace-pre-wrap select-text"
									>
										{expandedErrors[log.id] || log.error_detail.length <= 120
											? log.error_detail
											: truncateErrorDetail(log.error_detail, 120)}
									</p>
								</div>
							{/if}

							<span class="text-xs text-muted-foreground/60 mt-1">{formatDate(log.sent_at)}</span>
						</div>

						<span
							class="shrink-0 text-[10px] font-semibold px-2.5 py-0.5 rounded-full uppercase tracking-wider {log.status ===
							'sent'
								? 'bg-primary/10 text-primary'
								: 'bg-destructive/10 text-destructive'}"
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
					class="self-center mt-3 bg-secondary border border-border text-muted-foreground hover:text-foreground h-9 px-6 text-xs"
				>
					{isLoadingLogs ? 'Loading…' : 'Load more'}
				</Button>
			{/if}
		{/if}
	{/if}
</div>

<!-- ADD CHANNEL DIALOG -->
<Dialog.Root bind:open={isAddModalOpen}>
	<Dialog.Content
		class="sm:max-w-[520px] bg-card border border-border text-foreground p-6 rounded-lg"
	>
		<Dialog.Header>
			<Dialog.Title class="text-xl font-semibold tracking-tight"
				>Add Notification Channel</Dialog.Title
			>
			<Dialog.Description class="text-muted-foreground text-sm mt-1">
				Configure a new destination for monitor status change alerts.
			</Dialog.Description>
		</Dialog.Header>

		<div class="h-px w-full bg-border my-4"></div>

		{#if createError}
			<div
				class="p-3 text-[13px] text-destructive bg-destructive/10 border border-destructive/20 rounded-lg font-medium flex items-center gap-2 mb-4"
			>
				<TriangleAlert class="size-4 shrink-0" />
				<p class="m-0">{createError}</p>
			</div>
		{/if}

		<form onsubmit={handleCreateChannel} class="flex flex-col gap-5">
			<!-- Name -->
			<div class="space-y-2">
				<Label for="ch-name" class="text-sm font-medium text-foreground">
					Channel Name <span class="text-destructive font-bold">*</span>
				</Label>
				<Input
					id="ch-name"
					bind:value={newName}
					placeholder="e.g. Team Slack Alerts, Ops Webhook"
					class="bg-background border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-primary focus-visible:border-primary rounded-lg h-10 text-sm"
					required
				/>
			</div>

			<!-- Type Selection -->
			<div class="space-y-2">
				<Label class="text-sm font-medium text-foreground">
					Channel Type <span class="text-destructive font-bold">*</span>
				</Label>
				<div class="grid grid-cols-4 gap-2">
					<!-- Email -->
					<button
						type="button"
						onclick={() => (newType = 'email')}
						class="h-10 rounded-lg border text-xs font-medium transition-all flex items-center justify-center gap-1.5 {newType ===
						'email'
							? 'bg-primary/10 border-primary/40 text-primary font-semibold'
							: 'border-border bg-background text-foreground/70 hover:bg-white/5 hover:text-foreground'}"
					>
						<Mail class="size-3.5 text-blue-400" />
						Email
					</button>
					<!-- Slack -->
					<button
						type="button"
						onclick={() => (newType = 'slack')}
						class="h-10 rounded-lg border text-xs font-medium transition-all flex items-center justify-center gap-1.5 {newType ===
						'slack'
							? 'bg-primary/10 border-primary/40 text-primary font-semibold'
							: 'border-border bg-background text-foreground/70 hover:bg-white/5 hover:text-foreground'}"
					>
						<svg class="size-3.5" viewBox="0 0 122.8 122.8" xmlns="http://www.w3.org/2000/svg">
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
						Slack
					</button>
					<!-- Discord -->
					<button
						type="button"
						onclick={() => (newType = 'discord')}
						class="h-10 rounded-lg border text-xs font-medium transition-all flex items-center justify-center gap-1.5 {newType ===
						'discord'
							? 'bg-primary/10 border-primary/40 text-primary font-semibold'
							: 'border-border bg-background text-foreground/70 hover:bg-white/5 hover:text-foreground'}"
					>
						<svg class="size-3.5" viewBox="0 0 127.14 96.36" xmlns="http://www.w3.org/2000/svg">
							<path
								d="M107.7 8.07A105.15 105.15 0 0 0 81.47 0a72.06 72.06 0 0 0-3.36 6.83 97.68 97.68 0 0 0-29.17 0A72.37 72.37 0 0 0 45.57 0a105.89 105.89 0 0 0-26.25 8.09C2.79 32.65-1.71 56.6.54 80.21a105.73 105.73 0 0 0 32.17 16.15 77.7 77.7 0 0 0 6.89-11.11 68.42 68.42 0 0 1-10.85-5.18c.91-.66 1.8-1.34 2.66-2a75.57 75.57 0 0 0 64.32 0c.87.71 1.76 1.39 2.66 2a68.68 68.68 0 0 1-10.87 5.19 77 77 0 0 0 6.89 11.1 105.25 105.25 0 0 0 32.19-16.14c2.64-27.38-4.51-51.11-18.9-72.15ZM42.45 65.69C36.18 65.69 31 60 31 53s5-12.74 11.43-12.74S54 46 53.89 53s-5.05 12.69-11.44 12.69Zm42.24 0C78.41 65.69 73.25 60 73.25 53s5-12.74 11.44-12.74S96.23 46 96.12 53s-5.04 12.69-11.43 12.69Z"
								fill="#5865f2"
							/>
						</svg>
						Discord
					</button>
					<!-- Webhook -->
					<button
						type="button"
						onclick={() => (newType = 'webhook')}
						class="h-10 rounded-lg border text-xs font-medium transition-all flex items-center justify-center gap-1.5 {newType ===
						'webhook'
							? 'bg-primary/10 border-primary/40 text-primary font-semibold'
							: 'border-border bg-background text-foreground/70 hover:bg-white/5 hover:text-foreground'}"
					>
						<Webhook class="size-3.5 text-orange-400" />
						Webhook
					</button>
				</div>
			</div>

			<!-- Dynamic Input -->
			{#if newType === 'email'}
				<div class="space-y-2">
					<Label for="ch-email" class="text-sm font-medium text-foreground">
						Recipient Email <span class="text-destructive font-bold">*</span>
					</Label>
					<Input
						id="ch-email"
						type="email"
						bind:value={newEmail}
						placeholder="alerts@company.com"
						class="bg-background border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-primary focus-visible:border-primary rounded-lg h-10 text-sm"
						required
					/>
				</div>
			{:else if newType === 'slack'}
				<div class="space-y-2">
					<Label for="ch-slack" class="text-sm font-medium text-foreground">
						Slack Incoming Webhook URL <span class="text-destructive font-bold">*</span>
					</Label>
					<Input
						id="ch-slack"
						type="url"
						bind:value={newSlackUrl}
						placeholder="https://hooks.slack.com/services/..."
						class="bg-background border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-primary focus-visible:border-primary rounded-lg h-10 text-sm font-mono"
						required
					/>
					<span class="text-xs text-muted-foreground/60 block"
						>Must be an HTTPS Slack Webhook URL.</span
					>
				</div>
			{:else if newType === 'discord'}
				<div class="space-y-2">
					<Label for="ch-discord" class="text-sm font-medium text-foreground">
						Discord Webhook URL <span class="text-destructive font-bold">*</span>
					</Label>
					<Input
						id="ch-discord"
						type="url"
						bind:value={newDiscordUrl}
						placeholder="https://discord.com/api/webhooks/..."
						class="bg-background border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-primary focus-visible:border-primary rounded-lg h-10 text-sm font-mono"
						required
					/>
					<span class="text-xs text-muted-foreground/60 block"
						>Must be an HTTPS Discord Webhook URL.</span
					>
				</div>
			{:else}
				<div class="space-y-2">
					<Label for="ch-webhook" class="text-sm font-medium text-foreground">
						Webhook Destination URL <span class="text-destructive font-bold">*</span>
					</Label>
					<Input
						id="ch-webhook"
						type="url"
						bind:value={newWebhookUrl}
						placeholder="https://api.yourdomain.com/webhooks/pulsewatch"
						class="bg-background border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-primary focus-visible:border-primary rounded-lg h-10 text-sm font-mono"
						required
					/>
					<span class="text-xs text-muted-foreground/60 block"
						>Must be a public HTTPS URL. Internal IPs are blocked.</span
					>
				</div>
			{/if}

			<Dialog.Footer
				class="bg-transparent! m-0! p-0! pt-6 mt-2 flex justify-end gap-3 sm:space-x-0 border-t border-border"
			>
				<button
					type="button"
					onclick={() => (isAddModalOpen = false)}
					class="px-5 py-2.5 rounded-lg text-sm font-medium text-foreground hover:bg-accent transition-colors"
				>
					Cancel
				</button>
				<button
					type="submit"
					disabled={isCreating || !newName.trim()}
					class="h-10 px-5 rounded-lg text-sm font-medium bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground active:scale-[0.98] transition-all duration-200 disabled:opacity-50"
				>
					{isCreating ? 'Creating…' : 'Create Channel'}
				</button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>

<!-- DELETE CONFIRMATION DIALOG -->
<ConfirmDialog
	bind:open={isDeleteConfirmOpen}
	title="Delete notification channel?"
	description="Are you sure you want to delete channel '{channelToDelete?.name}'? Alerts will no longer be sent to this destination."
	confirmText="Delete"
	cancelText="Cancel"
	destructive={true}
	onConfirm={confirmDeleteChannel}
/>
