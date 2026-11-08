<script lang="ts">
	import { Send, Loader2 } from 'lucide-svelte';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import { debounce } from '$lib/utils';
	import { validateIncidentMessage } from '$lib/utils/validation';

	interface Props {
		isPosting: boolean;
		onPostUpdate: (message: string) => Promise<void>;
	}

	let { isPosting, onPostUpdate }: Props = $props();

	let message = $state('');
	let fieldError = $state('');
	let shake = $state(false);

	const handleValidation = debounce((val: string) => {
		if (!val.trim()) {
			fieldError = '';
			shake = false;
			return;
		}
		const err = validateIncidentMessage(val);
		if (err) {
			fieldError = err;
			shake = false;
			setTimeout(() => (shake = true), 10);
		} else {
			fieldError = '';
			shake = false;
		}
	}, 500);

	function handleInput() {
		if (!validateIncidentMessage(message)) {
			fieldError = '';
			shake = false;
		}
		handleValidation(message);
	}

	async function handleSubmit() {
		const trimmed = message.trim();
		if (!trimmed || isPosting) return;

		const err = validateIncidentMessage(trimmed);
		if (err) {
			fieldError = err;
			shake = false;
			setTimeout(() => (shake = true), 10);
			return;
		}

		fieldError = '';
		const msg = trimmed;
		message = '';
		await onPostUpdate(msg);
	}
</script>

<div class="pt-6 border-t border-border/60 space-y-3">
	<div class="flex items-start justify-between gap-2">
		<div class="space-y-0.5">
			<div class="flex items-center gap-3">
				<h3 class="text-sm font-semibold text-foreground tracking-tight">Post Status Update</h3>
				{#if fieldError}
					<span class="text-xs text-destructive font-medium">{fieldError}</span>
				{/if}
			</div>
			<p class="text-xs text-muted-foreground">
				Provide status findings, root cause analysis, or resolution steps.
			</p>
		</div>
		<span
			class="text-[11px] font-mono shrink-0 {message.trim().length > 200
				? 'text-destructive font-semibold'
				: 'text-muted-foreground'}"
		>
			{message.trim().length}/200 characters
		</span>
	</div>

	<div class="space-y-3">
		<Textarea
			bind:value={message}
			oninput={handleInput}
			placeholder="Describe current status, root cause findings, or mitigation steps..."
			class="min-h-24 resize-none text-xs {fieldError
				? 'border-destructive/70 focus-visible:border-destructive focus-visible:ring-error'
				: ''} {shake ? 'animate-shake' : ''}"
			rows={3}
		/>

		<div class="flex justify-end">
			<button
				class="bg-primary text-primary-foreground px-4 py-2 rounded-lg text-xs font-semibold hover:bg-primary/90 transition-colors flex items-center gap-2 cursor-pointer disabled:opacity-50"
				disabled={!message.trim() || isPosting || !!fieldError}
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
