<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Mail, Webhook, TriangleAlert } from 'lucide-svelte';
	import { NotificationAPI, type ChannelType } from '$lib/api/notification';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { toast } from 'svelte-sonner';

	interface Props {
		open: boolean;
		orgId: string;
		onSuccess: () => void;
	}

	let { open = $bindable(), orgId, onSuccess }: Props = $props();

	let newName = $state('');
	let newType = $state<ChannelType>('email');
	let newEmail = $state('');
	let newSlackUrl = $state('');
	let newWebhookUrl = $state('');
	let newDiscordUrl = $state('');
	let isCreating = $state(false);
	let createError = $state('');

	export function resetForm() {
		newName = '';
		newType = 'email';
		newEmail = '';
		newSlackUrl = '';
		newWebhookUrl = '';
		newDiscordUrl = '';
		createError = '';
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
			open = false;
			resetForm();
			onSuccess();
		} catch (err: unknown) {
			createError = parseApiError(err, 'Failed to create channel');
		} finally {
			isCreating = false;
		}
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-130 bg-card border border-border text-foreground p-6 rounded-lg">
		<Dialog.Header>
			<Dialog.Title class="text-xl font-semibold tracking-tight">
				Add Notification Channel
			</Dialog.Title>
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
					<span class="text-xs text-muted-foreground/60 block">
						Must be an HTTPS Slack Webhook URL.
					</span>
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
					<span class="text-xs text-muted-foreground/60 block">
						Must be an HTTPS Discord Webhook URL.
					</span>
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
						placeholder="https://api.yourdomain.com/webhooks/netra-uptime"
						class="bg-background border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-primary focus-visible:border-primary rounded-lg h-10 text-sm font-mono"
						required
					/>
					<span class="text-xs text-muted-foreground/60 block">
						Must be a public HTTPS URL. Internal IPs are blocked.
					</span>
				</div>
			{/if}

			<Dialog.Footer
				class="bg-transparent! m-0! p-0! pt-6 mt-2 flex justify-end gap-3 sm:space-x-0 border-t border-border"
			>
				<button
					type="button"
					onclick={() => (open = false)}
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
