<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { AlertTriangle } from 'lucide-svelte';

	interface Props {
		isOwner: boolean;
		onOpenTransfer: () => void;
		onOpenDisband: () => void;
		onOpenLeave: () => void;
	}

	let { isOwner, onOpenTransfer, onOpenDisband, onOpenLeave }: Props = $props();
</script>

<Card.Root class="bg-card border border-red-900/30">
	<Card.Header>
		<Card.Title class="text-red-500 flex items-center gap-2">
			<AlertTriangle class="size-5" />
			Danger Zone
		</Card.Title>
		<Card.Description class="text-muted-foreground">
			Irreversible and destructive actions for your organization.
		</Card.Description>
	</Card.Header>
	<Card.Content class="space-y-6">
		{#if isOwner}
			<!-- Transfer Ownership -->
			<div
				class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 py-4 border-b border-border/50"
			>
				<div class="space-y-1">
					<h4 class="font-medium text-foreground">Transfer Ownership</h4>
					<p class="text-sm text-muted-foreground">Transfer this organization to another member.</p>
				</div>
				<Button
					variant="outline"
					onclick={onOpenTransfer}
					class="bg-secondary border-border text-muted-foreground hover:bg-accent h-11 hover:text-foreground transition-colors cursor-pointer"
				>
					Transfer
				</Button>
			</div>

			<!-- Disband Organization -->
			<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 py-2">
				<div class="space-y-1">
					<h4 class="font-medium text-foreground">Disband Organization</h4>
					<p class="text-sm text-muted-foreground">
						Permanently delete this organization and all of its data.
					</p>
				</div>
				<Button
					variant="outline"
					onclick={onOpenDisband}
					class="bg-secondary border-border text-red-500 hover:bg-red-950/50 hover:border-red-900 hover:text-red-400 active:scale-[0.98] h-11 transition-all duration-200 font-medium cursor-pointer"
				>
					Disband Organization
				</Button>
			</div>
		{:else}
			<!-- Leave Organization -->
			<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 py-2">
				<div class="space-y-1">
					<h4 class="font-medium text-foreground">Leave Organization</h4>
					<p class="text-sm text-muted-foreground">
						Remove yourself from this organization. You will lose access to all resources.
					</p>
				</div>
				<Button
					variant="outline"
					onclick={onOpenLeave}
					class="bg-secondary border-border text-red-500 hover:bg-red-950/50 hover:border-red-900 hover:text-red-400 active:scale-[0.98] h-11 transition-all duration-200 font-medium cursor-pointer"
				>
					Leave Organization
				</Button>
			</div>
		{/if}
	</Card.Content>
</Card.Root>
