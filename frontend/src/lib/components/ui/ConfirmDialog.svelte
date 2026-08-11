<script lang="ts">
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import { buttonVariants } from '$lib/components/ui/button/index.js';

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

	let isConfirming = $state(false);

	async function handleConfirm() {
		isConfirming = true;
		try {
			await onConfirm();
			open = false;
		} finally {
			isConfirming = false;
		}
	}

	function handleCancel() {
		if (onCancel) onCancel();
		open = false;
	}
</script>

<AlertDialog.Root bind:open>
	<AlertDialog.Content class={destructive ? "bg-[#131314]" : ""}>
		<AlertDialog.Header>
			<AlertDialog.Title class={destructive ? "text-red-500 flex items-center gap-2" : ""}>{title}</AlertDialog.Title>
			{#if children}
				{@render children()}
			{:else}
				<AlertDialog.Description>{description}</AlertDialog.Description>
			{/if}
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel onclick={handleCancel} disabled={isConfirming} class="text-white hover:text-white h-9 hover:bg-white/10 transition-colors">
				{cancelText}
			</AlertDialog.Cancel>
			<AlertDialog.Action
				onclick={handleConfirm}
				class={destructive ? 'bg-[#1A1A1B] border h-9 border-[#2A2A2B] text-red-500 hover:bg-red-600 hover:border-red-600 hover:text-white transition-all duration-200 active:scale-[0.98] font-medium shadow-sm' : ''}
				disabled={isConfirming}
			>
				{#if isConfirming}
					<div class="size-4 border-2 border-current border-t-transparent rounded-full animate-spin mr-2"></div>
				{/if}
				{confirmText}
			</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
