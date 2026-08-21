<script lang="ts">
	import { onMount } from 'svelte';
	import {
		Key,
		Plus,
		Trash2,
		Copy,
		Check,
		Shield,
		ShieldCheck,
		AlertCircle,
		Clock,
		Eye
	} from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';
	import { APIKeyAPI, type APIKey, type CreateAPIKeyRequest } from '$lib/api/apikey';
	import { toast } from 'svelte-sonner';
	import { userTimezone } from '$lib/stores/timezone';
	import { formatDateOnly } from '$lib/utils/datetime';

	let { orgId, userRole } = $props<{ orgId: string; userRole: string }>();

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

	let canManage = $derived(userRole === 'owner' || userRole === 'admin');

	async function loadAPIKeys() {
		isLoading = true;
		fetchError = '';
		try {
			const res = await APIKeyAPI.getAllByOrg(orgId);
			apiKeys = res.data ?? [];
		} catch (err: any) {
			fetchError = err.message || 'Failed to load API keys';
		} finally {
			isLoading = false;
		}
	}

	async function handleCreateKey(e: Event) {
		e.preventDefault();
		if (!name.trim()) return;

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
			createError = err.message || 'Failed to create API key';
		} finally {
			isSubmitting = false;
		}
	}

	async function handleCopyKey() {
		if (!createdPlaintextKey) return;
		await navigator.clipboard.writeText(createdPlaintextKey);
		isCopied = true;
		setTimeout(() => (isCopied = false), 2000);
	}

	async function confirmRevokeKey() {
		if (!keyToRevoke) return;
		isRevoking = true;
		try {
			await APIKeyAPI.revoke(orgId, keyToRevoke.id);
			isRevokeOpen = false;
			keyToRevoke = null;
			await loadAPIKeys();
		} catch (err: any) {
			alert(err.message || 'Failed to revoke API key');
		} finally {
			isRevoking = false;
		}
	}

	// Format date menggunakan timezone dari store
	let tz = $derived($userTimezone);
	const formatDate = (d?: string) => formatDateOnly(d, tz);

	onMount(() => {
		loadAPIKeys();
	});
</script>

<div class="flex flex-col gap-6 w-full">
	<!-- Header & Action -->
	<div class="flex items-center justify-between gap-4 flex-wrap pb-4 border-b border-white/8">
		<div>
			<h2 class="text-lg font-semibold text-on-surface flex items-center gap-2">API Keys</h2>
			<p class="text-xs text-on-surface-variant/70 mt-0.5">
				Manage API keys for programmatically accessing the PulseWatch REST API.
			</p>
		</div>

		{#if canManage}
			<Button
				class="bg-[#1A1A1B] border border-[#2A2A2B] text-primary hover:bg-primary hover:border-primary hover:text-white active:scale-[0.98] transition-all duration-200 h-9 px-4 text-xs font-medium shadow-sm rounded-lg flex items-center gap-1.5"
				onclick={() => (isCreateOpen = true)}
			>
				<Plus class="size-4" /> Create API Key
			</Button>
		{/if}
	</div>

	<!-- Content Area -->
	{#if isLoading}
		<div
			class="p-8 text-center text-xs text-on-surface-variant/60 bg-white/2 border border-white/6 rounded-lg"
		>
			Loading API keys...
		</div>
	{:else if fetchError}
		<div
			class="p-4 text-xs text-rose-400 bg-rose-500/10 border border-rose-500/20 rounded-lg flex items-center gap-2"
		>
			<AlertCircle class="size-4 shrink-0" />
			{fetchError}
		</div>
	{:else if apiKeys.length === 0}
		<div
			class="p-10 text-center flex flex-col items-center justify-center gap-3 bg-white/2 border border-white/6 rounded-lg"
		>
			<div
				class="size-10 rounded-full bg-white/5 border border-white/10 flex items-center justify-center text-on-surface-variant"
			>
				<Key class="size-5" />
			</div>
			<div>
				<h3 class="text-sm font-semibold text-on-surface">No API keys created yet</h3>
				<p class="text-xs text-on-surface-variant/70 mt-1">
					Create an API key to allow CI/CD scripts, bots, or custom services to access your monitors
					and incidents.
				</p>
			</div>
			{#if canManage}
				<Button
					class="mt-2 bg-[#1A1A1B] border border-[#2A2A2B] text-primary hover:bg-primary hover:border-primary hover:text-white active:scale-[0.98] transition-all duration-200 h-9 px-4 text-xs font-medium shadow-sm rounded-lg flex items-center gap-1.5"
					onclick={() => (isCreateOpen = true)}
				>
					<Plus class="size-3.5" /> Create First Key
				</Button>
			{/if}
		</div>
	{:else}
		<!-- API Keys Table / Card List -->
		<div class="flex flex-col gap-3">
			{#each apiKeys as key (key.id)}
				<div
					class="p-4 rounded-lg bg-white/3 border border-white/8 flex items-center justify-between gap-4 flex-wrap transition-colors hover:border-white/14"
				>
					<div class="flex flex-col gap-1 min-w-0">
						<div class="flex items-center gap-2.5 flex-wrap">
							<span class="text-sm font-semibold text-on-surface truncate">{key.name}</span>

							<!-- Scope Badge -->
							<span
								class="text-[10.5px] font-mono px-2 py-0.5 rounded-md border font-medium {key.permissions ===
								'write'
									? 'bg-amber-500/10 border-amber-500/20 text-amber-400'
									: 'bg-blue-500/10 border-blue-500/20 text-blue-400'}"
							>
								{key.permissions === 'write' ? 'Full Access (Read & Write)' : 'Read Only'}
							</span>
						</div>

						<div
							class="flex items-center gap-4 text-[11.5px] font-mono text-on-surface-variant/70 flex-wrap"
						>
							<span
								class="bg-white/5 px-2 py-0.5 rounded border border-white/8 text-on-surface font-semibold"
							>
								{key.key_prefix}...****
							</span>
							<span class="flex items-center gap-1">
								Created by <strong class="text-on-surface-variant">{key.created_by}</strong>
							</span>
							<span class="flex items-center gap-1">
								<Clock class="size-3 shrink-0 opacity-60" /> Expires: {formatDate(key.expires_at)}
							</span>
						</div>
					</div>

					<div class="flex items-center gap-3 shrink-0">
						<div class="text-right text-[11px] text-on-surface-variant/60 hidden sm:block">
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
								class="p-2 rounded-lg text-on-surface-variant/70 hover:text-rose-400 hover:bg-rose-500/10 transition-colors"
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
		class="sm:max-w-[480px] bg-[#161618] border border-white/10 text-white p-6 rounded-lg shadow-xl"
	>
		<Dialog.Header>
			<Dialog.Title class="text-lg font-semibold flex items-center gap-2">
				<Key class="size-5 text-primary" /> Create New API Key
			</Dialog.Title>
			<Dialog.Description class="text-xs text-on-surface-variant/70 mt-1">
				Generate an API key for automated access. You will only be shown the key once.
			</Dialog.Description>
		</Dialog.Header>

		<form onsubmit={handleCreateKey} class="flex flex-col gap-4 mt-4">
			{#if createError}
				<div class="p-3 text-xs text-rose-400 bg-rose-500/10 border border-rose-500/20 rounded-lg">
					{createError}
				</div>
			{/if}

			<div class="space-y-1.5">
				<Label for="keyName" class="text-xs font-semibold text-on-surface"
					>Key Name <span class="text-rose-400">*</span></Label
				>
				<Input
					id="keyName"
					placeholder="e.g., GitHub Actions CI/CD"
					bind:value={name}
					required
					class="bg-white/5 border-white/10 text-xs focus:border-primary rounded-lg"
				/>
			</div>

			<div class="space-y-1.5">
				<Label class="text-xs font-semibold text-on-surface">Scope / Permissions</Label>
				<div class="grid grid-cols-2 gap-3">
					<!-- svelte-ignore a11y_click_events_have_key_events -->
					<!-- svelte-ignore a11y_no_static_element_interactions -->
					<div
						class="p-3 rounded-lg border text-left cursor-pointer transition-all {permissions ===
						'read'
							? 'bg-primary/10 border-primary text-primary'
							: 'bg-white/3 border-white/8 text-on-surface-variant hover:bg-white/5'}"
						onclick={() => (permissions = 'read')}
					>
						<div class="flex items-center gap-1.5 font-semibold text-xs mb-1">
							<Shield class="size-3.5" /> Read Only
						</div>
						<p class="text-[11px] opacity-75">
							Read-only access to monitors, checks & status pages.
						</p>
					</div>

					<!-- svelte-ignore a11y_click_events_have_key_events -->
					<!-- svelte-ignore a11y_no_static_element_interactions -->
					<div
						class="p-3 rounded-lg border text-left cursor-pointer transition-all {permissions ===
						'write'
							? 'bg-amber-500/10 border-amber-500/20 text-amber-400'
							: 'bg-white/3 border-white/8 text-on-surface-variant hover:bg-white/5'}"
						onclick={() => (permissions = 'write')}
					>
						<div class="flex items-center gap-1.5 font-semibold text-xs mb-1">
							<ShieldCheck class="size-3.5" /> Full Access
						</div>
						<p class="text-[11px] opacity-75">
							Full read & write access to create/update monitors & incidents.
						</p>
					</div>
				</div>
			</div>

			<div class="space-y-1.5">
				<Label for="expiration" class="text-xs font-semibold text-on-surface">Expiration</Label>
				<select
					id="expiration"
					bind:value={expiresInDays}
					class="w-full p-2.5 text-xs bg-[#1a1a1c] border border-white/10 rounded-lg text-on-surface focus:outline-none focus:border-primary"
				>
					<option value={null}>Never (Recommended for CI/CD)</option>
					<option value={30}>30 Days</option>
					<option value={90}>90 Days</option>
					<option value={365}>1 Year</option>
				</select>
			</div>

			<div class="flex justify-end gap-2 pt-3 border-t border-white/8">
				<Button
					type="button"
					variant="ghost"
					class="text-white h-11 px-6 hover:text-white hover:bg-white/10 rounded-lg text-xs font-medium"
					onclick={() => (isCreateOpen = false)}
				>
					Cancel
				</Button>
				<Button
					type="submit"
					disabled={isSubmitting || !name.trim()}
					class="h-11 px-6 rounded-lg text-xs font-medium bg-[#1A1A1B] border border-[#2A2A2B] text-primary hover:bg-primary hover:border-primary hover:text-white active:scale-[0.98] transition-all duration-200 shadow-sm disabled:opacity-50"
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
		class="sm:max-w-[500px] bg-[#161618] border border-white/10 text-white p-6 rounded-lg shadow-xl"
	>
		<Dialog.Header>
			<Dialog.Title class="text-lg font-semibold text-emerald-400 flex items-center gap-2">
				<Check class="size-5" /> API Key Created Successfully
			</Dialog.Title>
			<Dialog.Description class="text-xs text-on-surface-variant/70 mt-1">
				Please copy your API key now. For your security, <strong class="text-rose-400"
					>it will not be shown again!</strong
				>
			</Dialog.Description>
		</Dialog.Header>

		<div class="flex flex-col gap-4 mt-4">
			<div
				class="p-3 bg-black/40 border border-white/10 rounded-lg font-mono text-xs text-primary break-all flex items-center justify-between gap-3"
			>
				<span>{createdPlaintextKey}</span>
				<Button
					size="sm"
					class="shrink-0 gap-1.5 bg-[#1A1A1B] border border-[#2A2A2B] text-slate-300 hover:text-white hover:bg-[#2A2A2B] active:scale-[0.98] transition-all duration-200 rounded-lg text-xs font-medium"
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
					<strong>Keep this key secret!</strong> Store it securely in your CI/CD environment variables
					or password manager.
				</div>
			</div>

			<div class="flex justify-end pt-2">
				<Button
					class="bg-[#1A1A1B] border border-[#2A2A2B] text-primary hover:bg-primary hover:border-primary hover:text-white active:scale-[0.98] transition-all duration-200 h-9 px-5 text-xs font-medium rounded-lg shadow-sm"
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
	description="Are you sure you want to revoke '{keyToRevoke?.name}'? Any scripts or integrations using this key will immediately lose access to the PulseWatch API."
	confirmText={isRevoking ? 'Revoking...' : 'Revoke Key'}
	cancelText="Cancel"
	destructive={true}
	onConfirm={confirmRevokeKey}
/>
