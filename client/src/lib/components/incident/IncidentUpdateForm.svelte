<script lang="ts">
	import { Send, Loader2 } from 'lucide-svelte';
	import { Textarea } from '$lib/components/ui/textarea/index.js';

	interface Props {
		isPosting: boolean;
		onPostUpdate: (message: string) => Promise<void>;
	}

	let { isPosting, onPostUpdate }: Props = $props();

	let message = $state('');

	async function handleSubmit() {
		if (!message.trim() || isPosting) return;
		const msg = message;
		message = '';
		await onPostUpdate(msg);
	}
</script>

<div class="pt-6 border-t border-border/60 space-y-3">
	<div class="space-y-0.5">
		<h3 class="text-sm font-semibold text-foreground tracking-tight">Post Status Update</h3>
		<p class="text-xs text-muted-foreground">
			Provide status findings, root cause analysis, or resolution steps.
		</p>
	</div>

	<div class="space-y-3">
		<Textarea
			bind:value={message}
			placeholder="Describe current status, root cause findings, or mitigation steps..."
			class="min-h-24 resize-none text-xs"
			rows={3}
		/>

		<div class="flex justify-end">
			<button
				class="bg-primary text-primary-foreground px-4 py-2 rounded-lg text-xs font-semibold hover:bg-primary/90 transition-colors flex items-center gap-2 cursor-pointer disabled:opacity-50"
				disabled={!message.trim() || isPosting}
				onclick={handleSubmit}
			>
				{#if isPosting}
					<Loader2 class="size-3.5 animate-spin" />
					<span>Posting Update...</span>
				{:else}
					<Send class="size-3.5" />
					<span>Post Status Update</span>
				{/if}
			</button>
		</div>
	</div>
</div>
