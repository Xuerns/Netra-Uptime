<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLInputAttributes } from 'svelte/elements';
	import { Eye, EyeOff } from 'lucide-svelte';

	interface Props {
		id?: string;
		name?: string;
		label?: string;
		value: string;
		placeholder?: string;
		error?: string;
		shake?: boolean;
		required?: boolean;
		minlength?: number;
		autocomplete?: HTMLInputAttributes['autocomplete'];
		oninput?: (event: Event & { currentTarget: HTMLInputElement }) => void;
		headerRight?: Snippet;
	}

	let {
		id = 'password',
		name = 'password',
		label = 'Password',
		value = $bindable(''),
		placeholder = '••••••••',
		error = '',
		shake = false,
		required = true,
		minlength,
		autocomplete,
		oninput,
		headerRight
	}: Props = $props();

	let showPassword = $state(false);
</script>

<div class="flex flex-col gap-1">
	<div class="flex justify-between items-center">
		<div class="flex items-center gap-3">
			{#if label}
				<label class="font-label-md text-muted-foreground" for={id}>
					{label}
				</label>
			{/if}
			{#if error}
				<span class="text-xs text-destructive font-medium">{error}</span>
			{/if}
		</div>
		{#if headerRight}
			{@render headerRight()}
		{/if}
	</div>
	<div class="relative">
		<input
			bind:value
			{id}
			{name}
			{placeholder}
			{required}
			{minlength}
			{autocomplete}
			type={showPassword ? 'text' : 'password'}
			oninput={(e) => oninput?.(e)}
			class="w-full rounded-lg px-3 py-2.5 bg-background border border-border text-foreground font-body-md placeholder:text-muted-foreground/50 focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary transition-colors pr-10 {error
				? 'border-destructive/70 focus:border-destructive focus:ring-destructive'
				: ''} {shake ? 'animate-shake' : ''}"
		/>
		<button
			type="button"
			class="absolute inset-y-0 right-0 pr-3 flex items-center text-muted-foreground hover:text-foreground transition-colors focus:outline-none"
			onclick={() => (showPassword = !showPassword)}
			tabindex="-1"
			aria-label={showPassword ? 'Hide password' : 'Show password'}
		>
			{#if showPassword}
				<EyeOff class="w-5 h-5" />
			{:else}
				<Eye class="w-5 h-5" />
			{/if}
		</button>
	</div>
</div>
