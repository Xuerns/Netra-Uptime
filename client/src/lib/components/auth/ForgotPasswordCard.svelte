<script lang="ts">
	import Logo from '$lib/assets/Logo.svg';
	import { Mail, Loader2 } from 'lucide-svelte';
	import { AuthAPI } from '$lib/api/auth';
	import { validateEmail } from '$lib/utils/validation';
	import { debounce } from '$lib/utils';
	import { parseApiError } from '$lib/utils/parseApiError';

	interface Props {
		onSuccess?: (email: string) => void;
	}

	let { onSuccess }: Props = $props();

	let email = $state('');
	let requestLoading = $state(false);
	let requestError = $state('');
	let requestSuccess = $state(false);

	let fieldError = $state('');
	let shake = $state(false);

	const handleEmailInput = debounce((val: string) => {
		const err = validateEmail(val);
		if (err) {
			fieldError = err;
			shake = false;
			setTimeout(() => (shake = true), 10);
		} else {
			fieldError = '';
			shake = false;
		}
	}, 500);

	async function handleRequestReset(event: Event) {
		event.preventDefault();

		const err = validateEmail(email);
		fieldError = err;
		if (err) return;

		requestLoading = true;
		requestError = '';

		try {
			await AuthAPI.forgotPassword(email);
			requestSuccess = true;
			onSuccess?.(email);
		} catch (err: unknown) {
			requestError = parseApiError(err, 'Failed to send reset link.');
		} finally {
			requestLoading = false;
		}
	}
</script>

{#if requestSuccess}
	<!-- Success state -->
	<div class="flex flex-col items-center text-center py-md space-y-md animate-fade-in">
		<div class="w-14 h-14 rounded-full bg-primary/10 flex items-center justify-center">
			<Mail class="w-7 h-7 text-primary" />
		</div>
		<div>
			<h1
				class="font-headline-lg text-[24px] font-semibold text-foreground tracking-tight"
				style="text-wrap: balance;"
			>
				Check your email
			</h1>
			<p class="font-body-md text-muted-foreground mt-xs">
				We sent a password reset link to
				<span class="text-foreground font-medium">{email}</span>. Check your inbox and follow the
				link to reset your password.
			</p>
		</div>
		<div class="w-full space-y-3 pt-xs">
			<button
				type="button"
				class="w-full bg-transparent hover:bg-accent text-foreground font-label-md py-2.5 px-4 rounded-lg border border-border transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 focus:ring-offset-background"
				onclick={() => {
					requestSuccess = false;
					email = '';
				}}
			>
				Try a different email
			</button>
			<a
				href="/auth/login"
				class="block w-full text-center text-primary hover:underline underline-offset-4 font-label-md py-1"
			>
				Back to login
			</a>
		</div>
	</div>
{:else}
	<!-- Request form -->
	<form class="space-y-md" onsubmit={handleRequestReset}>
		<div class="flex flex-col items-center mb-lg">
			<div class="mb-sm">
				<img alt="Netra Uptime Logo" class="h-10 w-auto" src={Logo} />
			</div>
			<h1
				class="font-headline-lg text-[24px] font-semibold text-foreground tracking-tight"
				style="text-wrap: balance;"
			>
				Forgot your password?
			</h1>
			<p class="font-body-md text-muted-foreground text-center mt-xs">
				Enter your email and we'll send you a reset link.
			</p>
		</div>

		<!-- Error -->
		{#if requestError}
			<div
				class="flex items-start gap-2 p-3 rounded-lg bg-destructive/10 border border-destructive/20 animate-fade-in"
			>
				<p class="font-body-md text-sm text-destructive">{requestError}</p>
			</div>
		{/if}

		<!-- Fields -->
		<div class="space-y-sm">
			<div class="flex flex-col gap-1">
				<div class="flex items-center gap-3">
					<label class="font-label-md text-muted-foreground" for="email"> Work Email </label>
					{#if fieldError}
						<span class="text-xs text-destructive font-medium">{fieldError}</span>
					{/if}
				</div>
				<input
					bind:value={email}
					oninput={() => handleEmailInput(email)}
					class="w-full rounded-lg px-3 py-2.5 bg-background border border-border text-foreground font-body-md placeholder:text-muted-foreground/50 focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary transition-colors {fieldError
						? 'border-destructive/70 focus:border-destructive focus:ring-destructive'
						: ''} {shake ? 'animate-shake' : ''}"
					id="email"
					name="email"
					placeholder="john@company.com"
					required
					type="email"
				/>
			</div>
		</div>

		<!-- Submit -->
		<div class="pt-2">
			<button
				class="w-full bg-primary hover:bg-primary/90 text-primary-foreground font-label-md font-semibold py-2.5 px-4 rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 focus:ring-offset-background flex items-center justify-center gap-2"
				type="submit"
				disabled={requestLoading}
			>
				{#if requestLoading}
					<Loader2 class="w-4 h-4 animate-spin" />
					Sending...
				{:else}
					Send reset link
				{/if}
			</button>
		</div>
	</form>

	<p class="mt-lg text-center font-body-md text-sm text-muted-foreground">
		Remember your password? <a
			class="text-primary hover:underline underline-offset-4"
			href="/auth/login">Log in</a
		>
	</p>
{/if}
