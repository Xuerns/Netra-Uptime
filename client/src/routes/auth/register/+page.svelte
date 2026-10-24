<script lang="ts">
	import Logo from '$lib/assets/Logo.svg';
	import { Eye, EyeOff } from 'lucide-svelte';
	import { goto } from '$app/navigation';
	import { AuthAPI } from '$lib/api/auth';
	import {
		validateEmail,
		validatePassword,
		validateName,
		validateConfirmPassword
	} from '$lib/utils/validation';
	import { debounce } from '$lib/helper/utils';
	import { isSelfHosted } from '$lib/utils/appMode';
	import { parseApiError } from '$lib/utils/parseApiError';

	let name = $state('');
	let email = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let error = $state('');
	let loading = $state(false);

	let showPassword = $state(false);
	let showConfirmPassword = $state(false);

	let fieldErrors = $state({ name: '', email: '', password: '', confirmPassword: '' });
	let shake = $state({ name: false, email: false, password: false, confirmPassword: false });

	const handleNameInput = debounce((val: string) => {
		const err = validateName(val);
		if (err) {
			fieldErrors.name = err;
			shake.name = false;
			setTimeout(() => (shake.name = true), 10);
		} else {
			fieldErrors.name = '';
			shake.name = false;
		}
	}, 500);

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

	const handlePasswordInput = debounce((val: string, confirm: string) => {
		const err = validatePassword(val);
		if (err) {
			fieldErrors.password = err;
			shake.password = false;
			setTimeout(() => (shake.password = true), 10);
		} else {
			fieldErrors.password = '';
			shake.password = false;
		}
		// Also validate confirm password if it's already filled
		if (confirm) handleConfirmPasswordInput(val, confirm);
	}, 500);

	const handleConfirmPasswordInput = debounce((pass: string, confirm: string) => {
		const err = validateConfirmPassword(pass, confirm);
		if (err) {
			fieldErrors.confirmPassword = err;
			shake.confirmPassword = false;
			setTimeout(() => (shake.confirmPassword = true), 10);
		} else {
			fieldErrors.confirmPassword = '';
			shake.confirmPassword = false;
		}
	}, 500);

	async function handleRegister(event: Event) {
		event.preventDefault();

		// Re-validate before submit
		const nameErr = validateName(name);
		const emailErr = validateEmail(email);
		const passErr = validatePassword(password);
		const confirmErr = validateConfirmPassword(password, confirmPassword);
		fieldErrors.name = nameErr;
		fieldErrors.email = emailErr;
		fieldErrors.password = passErr;
		fieldErrors.confirmPassword = confirmErr;
		if (nameErr || emailErr || passErr || confirmErr) return;

		loading = true;
		error = '';

		if (password !== confirmPassword) {
			error = 'Password does not match.';
			loading = false;
			return;
		}

		try {
			await AuthAPI.register({ name, email, password });
			goto('/auth/login');
		} catch (err: any) {
			error = parseApiError(err, 'Registration failed');
		} finally {
			loading = false;
		}
	}

	async function loginWithGoogle() {
		try {
			await AuthAPI.startGoogleOAuth();
		} catch (err: any) {
			error = parseApiError(err, 'Failed to start Google login');
		}
	}
</script>

<svelte:head>
	<title>Register | Netra Uptime</title>
</svelte:head>

<main
	class="min-h-screen flex items-center justify-center p-sm md:p-md text-foreground antialiased w-full relative auth-bg"
>
	<div
		class="w-full max-w-112 bg-card rounded-xl border border-border p-lg shadow-sm animate-fade-in"
	>
		<form class="space-y-md" onsubmit={handleRegister}>
			<div class="flex flex-col items-center mb-lg">
				<div class="mb-sm">
					<img alt="Netra Uptime Logo" class="h-10 w-auto" src={Logo} />
				</div>
				<h1
					class="font-headline-lg text-[24px] font-semibold text-foreground tracking-tight"
					style="text-wrap: balance;"
				>
					Create your account
				</h1>
				<p class="font-body-md text-muted-foreground text-center mt-xs">
					Start monitoring your systems today.
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
						<label class="font-label-md text-muted-foreground" for="name"> Full Name </label>
						{#if fieldErrors.name}
							<span class="text-xs text-destructive font-medium">{fieldErrors.name}</span>
						{/if}
					</div>
					<input
						bind:value={name}
						oninput={() => handleNameInput(name)}
						class="w-full rounded-lg px-3 py-2.5 bg-background border border-border text-foreground font-body-md placeholder:text-muted-foreground/50 focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary transition-colors {fieldErrors.name
							? 'border-destructive/70 focus:border-destructive focus:ring-destructive'
							: ''} {shake.name ? 'animate-shake' : ''}"
						id="name"
						name="name"
						placeholder="John Doe"
						required
						type="text"
					/>
				</div>
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
				<div class="flex flex-col gap-1">
					<div class="flex items-center gap-3">
						<label class="font-label-md text-muted-foreground" for="password"> Password </label>
						{#if fieldErrors.password}
							<span class="text-xs text-destructive font-medium">{fieldErrors.password}</span>
						{/if}
					</div>
					<div class="relative">
						<input
							bind:value={password}
							oninput={() => handlePasswordInput(password, confirmPassword)}
							class="w-full rounded-lg px-3 py-2.5 bg-background border border-border text-foreground font-body-md placeholder:text-muted-foreground/50 focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary transition-colors pr-10 {fieldErrors.password
								? 'border-destructive/70 focus:border-destructive focus:ring-destructive'
								: ''} {shake.password ? 'animate-shake' : ''}"
							id="password"
							name="password"
							placeholder="••••••••"
							required
							type={showPassword ? 'text' : 'password'}
						/>
						<button
							type="button"
							class="absolute inset-y-0 right-0 pr-3 flex items-center text-muted-foreground hover:text-foreground transition-colors focus:outline-none"
							onclick={() => (showPassword = !showPassword)}
							tabindex="-1"
						>
							{#if showPassword}
								<EyeOff class="w-5 h-5" />
							{:else}
								<Eye class="w-5 h-5" />
							{/if}
						</button>
					</div>
				</div>
				<div class="flex flex-col gap-1">
					<div class="flex items-center gap-3">
						<label class="font-label-md text-muted-foreground" for="confirmPassword">
							Confirm Password
						</label>
						{#if fieldErrors.confirmPassword}
							<span class="text-xs text-destructive font-medium">{fieldErrors.confirmPassword}</span
							>
						{/if}
					</div>
					<div class="relative">
						<input
							bind:value={confirmPassword}
							oninput={() => handleConfirmPasswordInput(password, confirmPassword)}
							class="w-full rounded-lg px-3 py-2.5 bg-background border border-border text-foreground font-body-md placeholder:text-muted-foreground/50 focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary transition-colors pr-10 {fieldErrors.confirmPassword
								? 'border-destructive/70 focus:border-destructive focus:ring-destructive'
								: ''} {shake.confirmPassword ? 'animate-shake' : ''}"
							id="confirmPassword"
							name="confirmPassword"
							placeholder="••••••••"
							required
							type={showConfirmPassword ? 'text' : 'password'}
						/>
						<button
							type="button"
							class="absolute inset-y-0 right-0 pr-3 flex items-center text-muted-foreground hover:text-foreground transition-colors focus:outline-none"
							onclick={() => (showConfirmPassword = !showConfirmPassword)}
							tabindex="-1"
						>
							{#if showConfirmPassword}
								<EyeOff class="w-5 h-5" />
							{:else}
								<Eye class="w-5 h-5" />
							{/if}
						</button>
					</div>
				</div>
			</div>

			<!-- Submit -->
			<div class="pt-2">
				<button
					class="w-full bg-primary hover:bg-primary/90 text-primary-foreground font-label-md font-semibold py-2.5 px-4 rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 focus:ring-offset-background"
					type="submit"
					disabled={loading}
				>
					{#if loading}Loading...{:else}Create Account{/if}
				</button>
			</div>

			{#if !isSelfHosted()}
				<!-- Separator -->
				<div class="relative py-2 flex items-center">
					<div class="grow h-px bg-border"></div>
					<span class="shrink-0 mx-4 text-label-md text-muted-foreground">or</span>
					<div class="grow h-px bg-border"></div>
				</div>

				<!-- Google Auth Button -->
				<button
					type="button"
					onclick={loginWithGoogle}
					class="w-full bg-transparent hover:bg-accent text-foreground font-label-md py-2.5 px-4 rounded-lg border border-border transition-colors flex items-center justify-center gap-2 focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 focus:ring-offset-background"
				>
					<svg class="w-5 h-5" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
						<path
							d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
							fill="#4285F4"
						/>
						<path
							d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
							fill="#34A853"
						/>
						<path
							d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
							fill="#FBBC05"
						/>
						<path
							d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
							fill="#EA4335"
						/>
					</svg>
					<span>Continue with Google</span>
				</button>
			{/if}
		</form>

		<p class="mt-lg text-center font-body-md text-sm text-muted-foreground">
			Already have an account? <a
				class="text-primary hover:underline underline-offset-4"
				href="/auth/login">Log in</a
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
