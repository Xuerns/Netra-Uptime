<script lang="ts">
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';

	let {
		open = $bindable(false),
		title = 'Are you sure?',
		description = 'This action cannot be undone.',
		confirmText = 'Confirm',
		cancelText = 'Cancel',
		destructive = false,
		onConfirm,
		onCancel,
		children
	}: {
		open?: boolean;
		title?: string;
		description?: string;
		confirmText?: string;
		cancelText?: string;
		destructive?: boolean;
		onConfirm: () => void | Promise<void>;
		onCancel?: () => void;
		children?: import('svelte').Snippet;
	} = $props();
</script>

<ConfirmDialog
	bind:open
	{title}
	{description}
	{confirmText}
	{cancelText}
	{destructive}
	{onConfirm}
	{onCancel}
>
	{#if children}
		{@render children()}
	{/if}
</ConfirmDialog>
