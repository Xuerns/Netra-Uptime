<script lang="ts">
	import { Shield, ShieldCheck, AlertCircle } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { APIKeyAPI, type CreateAPIKeyRequest } from '$lib/api/apikey';
	import { parseApiError } from '$lib/utils/parseApiError';

	interface Props {
		open?: boolean;
		orgId: string;
		onSuccess: (plaintextKey: string) => void;
	}

	let { open = $bindable(false), orgId, onSuccess }: Props = $props();

	let name = $state('');
	let permissions = $state<'read' | 'write'>('read');
	let expiresInDays = $state<number | null>(null);
	let isSubmitting = $state(false);
	let createError = $state('');

	async function handleSubmit(e: Event) {
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
			const plaintextKey = res.data.plaintext_key;

			name = '';
			permissions = 'read';
			expiresInDays = null;
			createError = '';
			open = false;

			onSuccess(plaintextKey);
		} catch (err: unknown) {
			createError = parseApiError(err, 'Failed to create API key');
		} finally {
			isSubmitting = false;
		}
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content
		class="sm:max-w-[480px] bg-card border border-border text-foreground p-6 rounded-lg"
	>
		<Dialog.Header>
			<Dialog.Title class="text-lg font-semibold flex items-center gap-2">
				Create New API Key
			</Dialog.Title>
			<Dialog.Description class="text-xs text-muted-foreground mt-1">
				Generate an API key for automated access. You will only be shown the key once.
			</Dialog.Description>
		</Dialog.Header>

		<form onsubmit={handleSubmit} class="flex flex-col gap-4 mt-4">
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
					onclick={() => (open = false)}
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
