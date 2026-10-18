<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';
	import { toast } from 'svelte-sonner';

	let isDeleteDialogOpen = $state(false);

	async function deleteAccount() {
		toast.info('Account deletion is restricted. Please contact your administrator.');
		isDeleteDialogOpen = false;
	}
</script>

<Card.Root class="bg-card border border-destructive/30">
	<Card.Header>
		<Card.Title class="text-destructive flex items-center gap-2">Danger Zone</Card.Title>
		<Card.Description class="text-muted-foreground">
			Irreversible and destructive actions for your account.
		</Card.Description>
	</Card.Header>
	<Card.Content>
		<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pt-4">
			<div class="space-y-1">
				<h4 class="font-medium text-foreground">Delete Account</h4>
				<p class="text-sm text-muted-foreground">
					Permanently delete your account and all associated data. This action cannot be undone.
				</p>
			</div>
			<Button
				class="shrink-0 h-11 px-6 bg-secondary border border-border text-destructive hover:bg-destructive hover:border-destructive hover:text-foreground transition-all duration-200 active:scale-[0.98] font-medium shadow-sm cursor-pointer"
				onclick={() => (isDeleteDialogOpen = true)}
			>
				Delete Account
			</Button>
		</div>
	</Card.Content>
</Card.Root>

<ConfirmDialog
	bind:open={isDeleteDialogOpen}
	title="Delete Account"
	description="Are you absolutely sure you want to delete your account? This action cannot be undone and will permanently delete your data."
	confirmText="Yes, delete my account"
	destructive={true}
	onConfirm={deleteAccount}
/>
