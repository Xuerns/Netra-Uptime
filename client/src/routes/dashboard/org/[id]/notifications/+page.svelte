<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { headerStore } from '$lib/stores/headerStore.svelte';
	import { Bell, Plus, TriangleAlert } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';
	import {
		NotificationAPI,
		type NotificationChannel,
		type NotificationLog
	} from '$lib/api/notification';
	import { toast } from 'svelte-sonner';
	import { userTimezone } from '$lib/stores/timezone';
	import { formatDateTime } from '$lib/utils/datetime';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { DataSkeleton } from '$lib/components/common';
	import AddChannelModal from '$lib/components/notification/AddChannelModal.svelte';
	import ChannelCard from '$lib/components/notification/ChannelCard.svelte';
	import NotificationHistoryList from '$lib/components/notification/NotificationHistoryList.svelte';
	import {
		getCurrentUserId,
		findMemberByUserId,
		getMemberRole,
		canManageOrg
	} from '$lib/utils/role';

	const orgId = $derived($page.params.id as string);

	type SubTab = 'channels' | 'history';
	let activeSubTab = $state<string>('channels');

	// Channels state
	let channels = $state<NotificationChannel[]>([]);
	let isLoadingChannels = $state(true);
	let channelsError = $state('');

	// Add Modal state
	let isAddModalOpen = $state(false);

	// Toggle & Delete state
	let togglingId = $state<string | null>(null);
	let isDeleteConfirmOpen = $state(false);
	let channelToDelete = $state<NotificationChannel | null>(null);

	// History state
	let logs = $state<NotificationLog[]>([]);
	let logsTotal = $state(0);
	let isLoadingLogs = $state(false);
	let logsError = $state('');
	let logsOffset = $state(0);
	const logsLimit = 20;

	const currentUserId = $derived(getCurrentUserId($page.data?.user));
	const currentMember = $derived(findMemberByUserId($page.data?.members, currentUserId));
	const userRole = $derived(getMemberRole(currentMember, 'admin'));
	const canManage = $derived(canManageOrg(userRole));

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
		} catch (err: unknown) {
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
		} catch (err: unknown) {
			logsError = parseApiError(err, 'Failed to load notification logs');
		} finally {
			isLoadingLogs = false;
		}
	}

	async function handleToggleChannel(ch: NotificationChannel) {
		if (!orgId) return;
		togglingId = ch.id;
		try {
			await NotificationAPI.updateChannel(orgId, ch.id, { is_active: !ch.is_active });
			toast.success(ch.is_active ? 'Channel disabled' : 'Channel enabled');
			await loadChannels();
		} catch (err: unknown) {
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
		try {
			await NotificationAPI.deleteChannel(orgId, channelToDelete.id);
			toast.success(`Channel "${channelToDelete.name}" deleted`);
			await loadChannels();
		} catch (err: unknown) {
			toast.error(parseApiError(err, 'Failed to delete channel'));
		} finally {
			channelToDelete = null;
		}
	}

	let tz = $derived($userTimezone);
	const formatDate = (iso: string) => formatDateTime(iso, tz);

	// Sync with layout top bar
	$effect(() => {
		headerStore.set({
			actionType: activeSubTab === 'channels' && canManage ? 'button' : null,
			actionLabel: 'Add Channel',
			onAction: () => {
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
	<title>Notification Channels | Netra Uptime</title>
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
	<Tabs.Root bind:value={activeSubTab} class="w-full">
		<div class="flex items-center justify-between border-b border-border pb-3 mb-6">
			<Tabs.List class="h-9">
				<Tabs.Trigger value="channels" class="px-3.5 py-1.5 text-xs font-medium cursor-pointer">
					Channels ({channels.length})
				</Tabs.Trigger>
				<Tabs.Trigger value="history" class="px-3.5 py-1.5 text-xs font-medium cursor-pointer">
					History Logs
				</Tabs.Trigger>
			</Tabs.List>
		</div>

		<!-- TAB 1: CHANNELS -->
		<Tabs.Content value="channels">
			{#if isLoadingChannels}
				<DataSkeleton variant="channel" count={4} />
			{:else if channelsError}
				<div
					class="flex flex-col items-center justify-center gap-3 py-16 border border-border rounded-lg bg-card"
				>
					<TriangleAlert class="size-8 text-destructive" />
					<p class="text-sm text-muted-foreground">{channelsError}</p>
					<Button
						class="bg-secondary border border-border text-muted-foreground hover:text-foreground hover:bg-accent h-9 px-4 text-xs font-medium cursor-pointer"
						onclick={loadChannels}
					>
						Try again
					</Button>
				</div>
			{:else if channels.length === 0}
				<div
					class="flex flex-col items-center justify-center text-center py-20 px-6 gap-3.5 bg-card border border-border rounded-lg"
				>
					<div
						class="flex items-center justify-center size-14 rounded-lg bg-primary/8 text-primary"
					>
						<Bell class="size-6" />
					</div>
					<h2 class="text-[18px] font-semibold text-foreground m-0">
						No notification channels yet
					</h2>
					<p class="text-sm leading-relaxed text-muted-foreground max-w-95 m-0">
						{#if canManage}
							Add Email, Slack Incoming Webhooks, Discord, or Custom Webhooks to send monitor
							downtime alerts to your team.
						{:else}
							No notification channels configured for this organization.
						{/if}
					</p>
					{#if canManage}
						<Button
							class="mt-2 bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground h-10 px-5 text-xs font-medium cursor-pointer"
							onclick={() => {
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
						<ChannelCard
							channel={ch}
							{canManage}
							isToggling={togglingId === ch.id}
							onToggle={handleToggleChannel}
							onDelete={openDeleteConfirm}
						/>
					{/each}
				</div>
			{/if}
		</Tabs.Content>

		<!-- TAB 2: HISTORY LOGS -->
		<Tabs.Content value="history">
			{#if isLoadingLogs && logs.length === 0}
				<DataSkeleton variant="list" count={4} />
			{:else if logsError}
				<div
					class="flex flex-col items-center justify-center gap-3 py-16 border border-border rounded-lg bg-card"
				>
					<TriangleAlert class="size-8 text-destructive" />
					<p class="text-sm text-muted-foreground">{logsError}</p>
					<Button
						class="bg-secondary border border-border text-muted-foreground hover:text-foreground h-9 px-4 text-xs cursor-pointer"
						onclick={() => loadLogs()}
					>
						Try again
					</Button>
				</div>
			{:else}
				<NotificationHistoryList
					{logs}
					{logsTotal}
					{isLoadingLogs}
					{formatDate}
					onLoadMore={async () => {
						logsOffset += logsLimit;
						await loadLogs(false);
					}}
				/>
			{/if}
		</Tabs.Content>
	</Tabs.Root>
</div>

<!-- ADD CHANNEL DIALOG -->
<AddChannelModal bind:open={isAddModalOpen} {orgId} onSuccess={loadChannels} />

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
