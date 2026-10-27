<script lang="ts">
	import { Check, Copy, AlertCircle } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { toast } from 'svelte-sonner';

	interface Props {
		open?: boolean;
		plaintextKey: string;
		onClose: () => void;
	}

	let { open = $bindable(false), plaintextKey, onClose }: Props = $props();

	let isCopied = $state(false);

	async function handleCopyKey() {
		if (!plaintextKey) return;
		await navigator.clipboard.writeText(plaintextKey);
		isCopied = true;
		toast.success('API key copied to clipboard');
		setTimeout(() => (isCopied = false), 2000);
	}

	function handleDone() {
		open = false;
		onClose();
	}
</script>

<Dialog.Root bind:open>
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
				<span>{plaintextKey}</span>
				<Button
					size="sm"
					class="shrink-0 gap-1.5 bg-secondary border border-border text-muted-foreground hover:text-foreground hover:bg-accent active:scale-[0.98] transition-all duration-200 rounded-lg text-xs font-medium cursor-pointer"
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
					class="bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground active:scale-[0.98] transition-all duration-200 h-9 px-5 text-xs font-medium rounded-lg cursor-pointer"
					onclick={handleDone}
				>
					Done
				</Button>
			</div>
		</div>
	</Dialog.Content>
</Dialog.Root>
