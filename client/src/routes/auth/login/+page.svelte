<script lang="ts">
	import Logo from '$lib/assets/Logo.svg';
	import { goto } from '$app/navigation';
	import { AuthAPI } from '$lib/api/auth';
	import { validateEmail, validatePassword } from '$lib/utils/validation';
	import { debounce } from '$lib/utils';
	import { isSelfHosted } from '$lib/utils/appMode';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { OAuthButton, PasswordInput } from '$lib/components/auth';
	import type { PageData } from './$types';

	// eslint-disable-next-line @typescript-eslint/no-unused-vars
	let { data }: { data: PageData } = $props();

	let email = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);

	let fieldErrors = $state({ email: '', password: '' });
	let shake = $state({ email: false, password: false });

	const handleEmailInput = debounce((val: string) => {
		const err = validateEmail(val);
		if (err) {
			fieldErrors.email = err;
			shake.email = false;
			setTimeout(() => (shake.email = true), 10);
		} else {
			fieldErrors.email = '';
			shake.email = false;
		}
	}, 500);

	const handlePasswordInput = debounce((val: string) => {
		const err = validatePassword(val);
		if (err) {
			fieldErrors.password = err;
			shake.password = false;
			setTimeout(() => (shake.password = true), 10);
		} else {
			fieldErrors.password = '';
			shake.password = false;
		}
	}, 500);

	async function handleLogin(event: Event) {
		event.preventDefault();

		const emailErr = validateEmail(email);
		const passErr = validatePassword(password);
		fieldErrors.email = emailErr;
		fieldErrors.password = passErr;
		if (emailErr || passErr) return;

		loading = true;
		error = '';

		try {
			await AuthAPI.login({ email, password });
			goto('/dashboard/org');
		} catch (err: unknown) {
			error = parseApiError(err, 'Login failed');
		} finally {
			loading = false;
		}
	}

	async function loginWithGoogle() {
		try {
			await AuthAPI.startGoogleOAuth();
		} catch (err: unknown) {
			error = parseApiError(err, 'Failed to start Google login');
		}
	}
</script>

<svelte:head>
	<title>Login | Netra Uptime</title>
</svelte:head>

{#snippet forgotPasswordLink()}
	<a
		href="/auth/forgot"
		class="text-sm font-label-md text-primary hover:underline underline-offset-4"
	>
		Forgot password?
	</a>
{/snippet}

<main
	class="min-h-screen flex items-center justify-center p-sm md:p-md text-foreground antialiased w-full relative auth-bg"
>
	<div
		class="w-full max-w-112 bg-card rounded-xl border border-border p-lg shadow-sm animate-fade-in"
	>
		<form class="space-y-md" onsubmit={handleLogin}>
			<div class="flex flex-col items-center mb-lg">
				<div class="mb-sm">
					<img alt="Netra Uptime Logo" class="h-10 w-auto" src={Logo} />
				</div>
				<h1
					class="font-headline-lg text-[24px] font-semibold text-foreground tracking-tight"
					style="text-wrap: balance;"
				>
					Welcome back
				</h1>
				<p class="font-body-md text-muted-foreground text-center mt-xs">
					Log in to monitor your systems.
				</p>
			</div>

			<!-- Error Banner -->
			{#if error}
				<div
					class="flex items-start gap-2 p-3 rounded-lg bg-destructive/10 border border-destructive/20 animate-fade-in"
				>
					<p class="font-body-md text-sm text-destructive">{error}</p>
				</div>
			{/if}

			<!-- Fields -->
			<div class="space-y-sm">
				<div class="flex flex-col gap-1">
					<div class="flex items-center gap-3">
						<label class="font-label-md text-muted-foreground" for="email"> Work Email </label>
						{#if fieldErrors.email}
							<span class="text-xs text-destructive font-medium">{fieldErrors.email}</span>
						{/if}
					</div>
					<input
						bind:value={email}
						oninput={() => handleEmailInput(email)}
						class="w-full rounded-lg px-3 py-2.5 bg-background border border-border text-foreground font-body-md placeholder:text-muted-foreground/50 focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary transition-colors {fieldErrors.email
							? 'border-destructive/70 focus:border-destructive focus:ring-destructive'
							: ''} {shake.email ? 'animate-shake' : ''}"
						id="email"
						name="email"
						placeholder="john@company.com"
						required
						type="email"
					/>
				</div>

				<PasswordInput
					id="password"
					name="password"
					label="Password"
					bind:value={password}
					error={fieldErrors.password}
					shake={shake.password}
					oninput={() => handlePasswordInput(password)}
					headerRight={forgotPasswordLink}
				/>
			</div>

			<!-- Submit -->
			<div class="pt-2">
				<button
					class="w-full bg-primary hover:bg-primary/90 text-primary-foreground font-label-md font-semibold py-2.5 px-4 rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 focus:ring-offset-background"
					type="submit"
					disabled={loading}
				>
					{#if loading}Loading...{:else}Login{/if}
				</button>
			</div>

			{#if !isSelfHosted()}
				<OAuthButton onclick={loginWithGoogle} />
			{/if}
		</form>

		<p class="mt-lg text-center font-body-md text-sm text-muted-foreground">
			Don't have an account? <a
				class="text-primary hover:underline underline-offset-4"
				href="/auth/register">Sign up</a
			>
		</p>
	</div>
</main>

<style>
	.auth-bg {
		background-color: var(--background);
		background-image: radial-gradient(var(--border) 1.5px, transparent 1.5px);
		background-size: 24px 24px;
	}
	:global(.dark) .auth-bg {
		background-image: radial-gradient(rgba(1, 101, 53, 0.15) 1.5px, transparent 1.5px);
	}
</style>
