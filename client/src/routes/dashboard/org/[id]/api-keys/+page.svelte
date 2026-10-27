<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { Key, AlertCircle } from 'lucide-svelte';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';
	import { APIKeyAPI, type APIKey } from '$lib/api/apikey';
	import { headerStore } from '$lib/stores/headerStore.svelte';
	import { toast } from 'svelte-sonner';
	import { userTimezone } from '$lib/stores/timezone';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { EmptyState, DataSkeleton } from '$lib/components/common';
	import CreateApiKeyModal from '$lib/components/apikey/CreateApiKeyModal.svelte';
	import ShowApiKeyModal from '$lib/components/apikey/ShowApiKeyModal.svelte';
	import ApiKeyItem from '$lib/components/apikey/ApiKeyItem.svelte';
	import {
		getCurrentUserId,
		findMemberByUserId,
		getMemberRole,
		canManageOrg
	} from '$lib/utils/role';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const orgId = $derived($page.params.id as string);

	let apiKeys = $state<APIKey[]>([]);
	let isLoading = $state(true);
	let fetchError = $state('');

	// Create Modal state
	let isCreateOpen = $state(false);

	// Show Plaintext Key Modal state
	let createdPlaintextKey = $state('');
	let isShowKeyOpen = $state(false);

	// Revoke Confirm Modal state
	let keyToRevoke = $state<APIKey | null>(null);
	let isRevokeOpen = $state(false);
	let isRevoking = $state(false);

	const currentUserId = $derived(getCurrentUserId(data?.user ?? $page.data?.user));
	const currentMember = $derived(
		findMemberByUserId(data?.members ?? $page.data?.members, currentUserId)
	);
	const userRole = $derived(getMemberRole(currentMember, 'admin'));
	const canManage = $derived(canManageOrg(userRole));

	async function loadAPIKeys() {
		if (!orgId) return;
		isLoading = true;
		fetchError = '';
		try {
			const res = await APIKeyAPI.getAllByOrg(orgId);
			apiKeys = res.data ?? [];
		} catch (err: unknown) {
			fetchError = parseApiError(err, 'Failed to load API keys');
		} finally {
			isLoading = false;
		}
	}

	async function handleKeyCreated(plaintextKey: string) {
		createdPlaintextKey = plaintextKey;
		isShowKeyOpen = true;
		await loadAPIKeys();
	}

	function handlePromptRevoke(key: APIKey) {
		keyToRevoke = key;
		isRevokeOpen = true;
	}

	async function confirmRevokeKey() {
		if (!keyToRevoke || !orgId) return;
		isRevoking = true;
		try {
			await APIKeyAPI.revoke(orgId, keyToRevoke.id);
			toast.success(`Key "${keyToRevoke.name}" revoked`);
			isRevokeOpen = false;
			keyToRevoke = null;
			await loadAPIKeys();
		} catch (err: unknown) {
			toast.error(parseApiError(err, 'Failed to revoke API key'));
		} finally {
			isRevoking = false;
		}
	}

	// Sync with layout top bar
	$effect(() => {
		headerStore.set({
			actionType: canManage ? 'button' : null,
			actionLabel: 'Create API Key',
			onAction: () => {
				isCreateOpen = true;
			},
			badge: {
				text: `${apiKeys.length} API Key${apiKeys.length !== 1 ? 's' : ''}`,
				variant: 'slate'
			}
		});
	});

	onMount(() => {
		if (orgId) {
			loadAPIKeys();
		}
	});
</script>

<svelte:head>
	<title>API Keys | Netra Uptime</title>
</svelte:head>

<div class="text-foreground antialiased">
	<!-- Page Header -->
	<header class="flex flex-wrap items-start justify-between gap-4 mb-6">
		<div>
			<h1 class="text-[32px] font-semibold leading-tight tracking-tight text-foreground mb-1.5">
				API Keys
			</h1>
			<p class="text-base leading-relaxed text-muted-foreground max-w-[50ch]">
				Manage API tokens for programmatically integrating and accessing the Netra Uptime REST API.
			</p>
		</div>
	</header>

	<!-- Content Area -->
	{#if isLoading}
		<DataSkeleton variant="list" count={3} />
	{:else if fetchError}
		<div
			class="p-4 text-xs text-rose-400 bg-rose-500/10 border border-rose-500/20 rounded-lg flex items-center gap-2"
		>
			<AlertCircle class="size-4 shrink-0" />
			{fetchError}
		</div>
	{:else if apiKeys.length === 0}
		<EmptyState
			icon={Key}
			title="No API keys created yet"
			description="Create an API key to allow CI/CD pipelines, automated scripts, or custom microservices to access monitors and incidents."
			actionText={canManage ? 'Create First Key' : undefined}
			onAction={() => (isCreateOpen = true)}
		/>
	{:else}
		<!-- API Keys List -->
		<div class="flex flex-col gap-3">
			{#each apiKeys as key (key.id)}
				<ApiKeyItem
					apiKey={key}
					{canManage}
					timezone={$userTimezone}
					onRevoke={handlePromptRevoke}
				/>
			{/each}
		</div>
	{/if}
</div>

<!-- Modal 1: Create API Key -->
<CreateApiKeyModal bind:open={isCreateOpen} {orgId} onSuccess={handleKeyCreated} />

<!-- Modal 2: Show Plaintext API Key Once -->
<ShowApiKeyModal
	bind:open={isShowKeyOpen}
	plaintextKey={createdPlaintextKey}
	onClose={() => {
		createdPlaintextKey = '';
	}}
/>

<!-- Modal 3: Revoke Confirm -->
<ConfirmDialog
	bind:open={isRevokeOpen}
	title="Revoke API Key?"
	description="Are you sure you want to revoke '{keyToRevoke?.name}'? Any scripts or integrations using this key will immediately lose access to the Netra Uptime API."
	confirmText={isRevoking ? 'Revoking...' : 'Revoke Key'}
	cancelText="Cancel"
	destructive={true}
	onConfirm={confirmRevokeKey}
/>

<!-- Modal 3: Revoke Confirm -->
<ConfirmDialog
	bind:open={isRevokeOpen}
	title="Revoke API Key?"
	description="Are you sure you want to revoke '{keyToRevoke?.name}'? Any scripts or integrations using this key will immediately lose access to the Netra Uptime API."
	confirmText={isRevoking ? 'Revoking...' : 'Revoke Key'}
	cancelText="Cancel"
	destructive={true}
	onConfirm={confirmRevokeKey}
/>
