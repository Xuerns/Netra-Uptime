<script lang="ts">
	import { Send, Loader2 } from 'lucide-svelte';

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

<div class="bg-[#161618] border border-white/10 rounded-lg p-5 space-y-3">
	<h2 class="text-sm font-semibold text-white tracking-tight">Post Incident Status Update</h2>

	<div class="space-y-3">
		<textarea
			bind:value={message}
			placeholder="Describe current status, root cause findings, or mitigation steps..."
			class="w-full bg-black/40 border border-white/10 rounded-lg p-3 text-xs text-white placeholder:text-white/30 focus:outline-none focus:border-primary transition-colors resize-none"
			rows="3"
		></textarea>

		<div class="flex justify-end">
			<button
				class="bg-primary text-on-primary px-4 py-2 rounded-lg text-xs font-semibold hover:bg-primary-container transition-colors flex items-center gap-2 cursor-pointer disabled:opacity-50"
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
