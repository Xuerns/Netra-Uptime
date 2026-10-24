<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import {
		Key,
		Trash2,
		Copy,
		Check,
		Shield,
		ShieldCheck,
		AlertCircle,
		Clock
	} from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';
	import { APIKeyAPI, type APIKey, type CreateAPIKeyRequest } from '$lib/api/apikey';
	import { headerStore } from '$lib/stores/headerStore.svelte';
	import { toast } from 'svelte-sonner';
	import { userTimezone } from '$lib/stores/timezone';
	import { formatDateOnly } from '$lib/utils/datetime';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { EmptyState, DataSkeleton } from '$lib/components/common';
	import {
		getCurrentUserId,
		findMemberByUserId,
		getMemberRole,
		canManageOrg
	} from '$lib/utils/role';

	const orgId = $derived($page.params.id as string);

	let apiKeys = $state<APIKey[]>([]);
	let isLoading = $state(true);
	let fetchError = $state('');

	// Create Modal state
	let isCreateOpen = $state(false);
	let name = $state('');
	let permissions = $state<'read' | 'write'>('read');
	let expiresInDays = $state<number | null>(null);
	let isSubmitting = $state(false);
	let createError = $state('');

	// Show Plaintext Key Modal state
	let createdPlaintextKey = $state('');
	let isShowKeyOpen = $state(false);
	let isCopied = $state(false);

	// Revoke Confirm Modal state
	let keyToRevoke = $state<APIKey | null>(null);
	let isRevokeOpen = $state(false);
	let isRevoking = $state(false);

	const currentUserId = $derived(getCurrentUserId($page.data?.user));
	const currentMember = $derived(findMemberByUserId($page.data?.members, currentUserId));
	const userRole = $derived(getMemberRole(currentMember, 'admin'));
	const canManage = $derived(canManageOrg(userRole));

	async function loadAPIKeys() {
		if (!orgId) return;
		isLoading = true;
		fetchError = '';
		try {
			const res = await APIKeyAPI.getAllByOrg(orgId);
			apiKeys = res.data ?? [];
		} catch (err: any) {
			fetchError = parseApiError(err, 'Failed to load API keys');
		} finally {
			isLoading = false;
		}
	}

	async function handleCreateKey(e: Event) {
		e.preventDefault();
		if (!name.trim() || !orgId) return;

		isSubmitting = true;
		createError = '';

		try {
			const req: CreateAPIKeyRequest = {
				name: name.trim(),
				permissions,
				expires_in_days: expiresInDays
			};
			const res = await APIKeyAPI.create(orgId, req);
			createdPlaintextKey = res.data.plaintext_key;

			isCreateOpen = false;
			isShowKeyOpen = true;
			name = '';
			permissions = 'read';
			expiresInDays = null;

			await loadAPIKeys();
		} catch (err: any) {
			createError = parseApiError(err, 'Failed to create API key');
		} finally {
			isSubmitting = false;
		}
	}

	async function handleCopyKey() {
		if (!createdPlaintextKey) return;
		await navigator.clipboard.writeText(createdPlaintextKey);
		isCopied = true;
		toast.success('API key copied to clipboard');
		setTimeout(() => (isCopied = false), 2000);
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
		} catch (err: any) {
			toast.error(parseApiError(err, 'Failed to revoke API key'));
		} finally {
			isRevoking = false;
		}
	}

	let tz = $derived($userTimezone);
	const formatDate = (d?: string) => formatDateOnly(d, tz);

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
				<div
					class="p-4 rounded-lg bg-card border border-border shadow-xs flex items-center justify-between gap-4 flex-wrap transition-colors hover:border-primary/30"
				>
					<div class="flex flex-col gap-1 min-w-0">
						<div class="flex items-center gap-2.5 flex-wrap">
							<span class="text-sm font-semibold text-foreground truncate">{key.name}</span>

							<!-- Scope Badge -->
							<Badge
								variant="outline"
								class="text-[11px] font-mono px-2 py-0.5 font-medium {key.permissions === 'write'
									? 'bg-amber-500/10 border-amber-500/20 text-amber-400'
									: 'bg-blue-500/10 border-blue-500/20 text-blue-400'}"
							>
								{key.permissions === 'write' ? 'Full Access (Read & Write)' : 'Read Only'}
							</Badge>
						</div>

						<div
							class="flex items-center gap-4 text-[11.5px] font-mono text-muted-foreground flex-wrap"
						>
							<span
								class="bg-white/5 px-2 py-0.5 rounded border border-border text-foreground font-semibold"
							>
								{key.key_prefix}...****
							</span>
							{#if key.created_by}
								<span class="flex items-center gap-1">
									Created by <strong class="text-muted-foreground">{key.created_by}</strong>
								</span>
							{/if}
							<span class="flex items-center gap-1">
								<Clock class="size-3 shrink-0 opacity-60" /> Expires: {formatDate(key.expires_at)}
							</span>
						</div>
					</div>

					<div class="flex items-center gap-3 shrink-0">
						<div class="text-right text-[11px] text-muted-foreground/60 hidden sm:block">
							{#if key.last_used_at}
								Last used {new Date(key.last_used_at).toLocaleDateString()}
								{#if key.last_used_ip}
									({key.last_used_ip}){/if}
							{:else}
								Never used
							{/if}
						</div>

						{#if canManage}
							<button
								class="p-2 rounded-lg text-muted-foreground hover:text-rose-400 hover:bg-rose-500/10 transition-colors"
								title="Revoke API key"
								onclick={() => {
									keyToRevoke = key;
									isRevokeOpen = true;
								}}
							>
								<Trash2 class="size-4" />
							</button>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Modal 1: Create API Key -->
<Dialog.Root bind:open={isCreateOpen}>
	<Dialog.Content
		class="sm:max-w-[480px] bg-card border border-border text-foreground p-6 rounded-lg"
	>
		<Dialog.Header>
			<Dialog.Title class="text-lg font-semibold flex items-center gap-2">
				<Key class="size-5 text-primary" /> Create New API Key
			</Dialog.Title>
			<Dialog.Description class="text-xs text-muted-foreground mt-1">
				Generate an API key for automated access. You will only be shown the key once.
			</Dialog.Description>
		</Dialog.Header>

		<form onsubmit={handleCreateKey} class="flex flex-col gap-4 mt-4">
			{#if createError}
				<div
					class="p-3 text-xs text-rose-400 bg-rose-500/10 border border-rose-500/20 rounded-lg flex items-center gap-2"
				>
					<AlertCircle class="size-4 shrink-0" />
					{createError}
				</div>
			{/if}

			<div class="space-y-1.5">
				<Label for="keyName" class="text-xs font-semibold text-foreground">
					Key Name <span class="text-destructive">*</span>
				</Label>
				<Input
					id="keyName"
					placeholder="e.g., GitHub Actions CI/CD"
					bind:value={name}
					required
					class="bg-white/5 border-border text-xs focus:border-primary rounded-lg h-10 text-sm"
				/>
			</div>

			<div class="space-y-1.5">
				<Label class="text-xs font-semibold text-foreground">Scope / Permissions</Label>
				<div class="grid grid-cols-2 gap-3">
					<button
						type="button"
						class="p-3 rounded-lg border text-left transition-all {permissions === 'read'
							? 'bg-primary/10 border-primary text-primary'
							: 'bg-white/3 border-border text-muted-foreground hover:bg-white/5'}"
						onclick={() => (permissions = 'read')}
					>
						<div class="flex items-center gap-1.5 font-semibold text-xs mb-1">
							<Shield class="size-3.5" /> Read Only
						</div>
						<p class="text-[11px] opacity-75">
							Read-only access to monitors, checks & status pages.
						</p>
					</button>

					<button
						type="button"
						class="p-3 rounded-lg border text-left transition-all {permissions === 'write'
							? 'bg-amber-500/10 border-amber-500/20 text-amber-400'
							: 'bg-white/3 border-border text-muted-foreground hover:bg-white/5'}"
						onclick={() => (permissions = 'write')}
					>
						<div class="flex items-center gap-1.5 font-semibold text-xs mb-1">
							<ShieldCheck class="size-3.5" /> Full Access
						</div>
						<p class="text-[11px] opacity-75">
							Full read & write access to create/update monitors & incidents.
						</p>
					</button>
				</div>
			</div>

			<div class="space-y-1.5">
				<Label for="expiration" class="text-xs font-semibold text-foreground">Expiration</Label>
				<select
					id="expiration"
					bind:value={expiresInDays}
					class="w-full h-10 px-3 text-xs bg-secondary border border-border rounded-lg text-foreground focus:outline-none focus:border-primary"
				>
					<option value={null}>Never (Recommended for CI/CD)</option>
					<option value={30}>30 Days</option>
					<option value={90}>90 Days</option>
					<option value={365}>1 Year</option>
				</select>
			</div>

			<div class="flex justify-end gap-2 pt-3 border-t border-border">
				<Button
					type="button"
					variant="ghost"
					class="text-foreground h-10 px-5 hover:text-foreground hover:bg-accent rounded-lg text-xs font-medium"
					onclick={() => (isCreateOpen = false)}
				>
					Cancel
				</Button>
				<Button
					type="submit"
					disabled={isSubmitting || !name.trim()}
					class="h-10 px-5 rounded-lg text-xs font-medium bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground active:scale-[0.98] transition-all duration-200 disabled:opacity-50"
				>
					{isSubmitting ? 'Generating...' : 'Generate API Key'}
				</Button>
			</div>
		</form>
	</Dialog.Content>
</Dialog.Root>

<!-- Modal 2: Show Plaintext API Key Once -->
<Dialog.Root bind:open={isShowKeyOpen}>
	<Dialog.Content
		class="sm:max-w-[500px] bg-card border border-border text-foreground p-6 rounded-lg"
	>
		<Dialog.Header>
			<Dialog.Title class="text-lg font-semibold text-emerald-400 flex items-center gap-2">
				<Check class="size-5" /> API Key Created Successfully
			</Dialog.Title>
			<Dialog.Description class="text-xs text-muted-foreground mt-1">
				Please copy your API key now. For security purposes, <strong class="text-rose-400"
					>it will never be displayed again!</strong
				>
			</Dialog.Description>
		</Dialog.Header>

		<div class="flex flex-col gap-4 mt-4">
			<div
				class="p-3 bg-foreground/10 border border-border rounded-lg font-mono text-xs text-primary break-all flex items-center justify-between gap-3"
			>
				<span>{createdPlaintextKey}</span>
				<Button
					size="sm"
					class="shrink-0 gap-1.5 bg-secondary border border-border text-muted-foreground hover:text-foreground hover:bg-accent active:scale-[0.98] transition-all duration-200 rounded-lg text-xs font-medium"
					onclick={handleCopyKey}
				>
					{#if isCopied}
						<Check class="size-3.5 text-emerald-400" /> Copied
					{:else}
						<Copy class="size-3.5" /> Copy
					{/if}
				</Button>
			</div>

			<div
				class="p-3 rounded-lg bg-amber-500/10 border border-amber-500/20 text-amber-300 text-xs flex items-start gap-2"
			>
				<AlertCircle class="size-4 shrink-0 mt-0.5" />
				<div>
					<strong>Keep this key secure:</strong> Store it in your encrypted secrets manager or CI/CD variables.
				</div>
			</div>

			<div class="flex justify-end pt-2">
				<Button
					class="bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground active:scale-[0.98] transition-all duration-200 h-9 px-5 text-xs font-medium rounded-lg"
					onclick={() => {
						isShowKeyOpen = false;
						createdPlaintextKey = '';
					}}
				>
					Done
				</Button>
			</div>
		</div>
	</Dialog.Content>
</Dialog.Root>

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
