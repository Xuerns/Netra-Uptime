<script lang="ts">
	import Logo from '$lib/assets/Logo.svg';
	import { goto } from '$app/navigation';
	import { AuthAPI } from '$lib/api/auth';
	import {
		validateEmail,
		validatePassword,
		validateName,
		validateConfirmPassword
	} from '$lib/utils/validation';
	import { debounce } from '$lib/utils';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { PasswordInput } from '$lib/components/auth';
	import type { PageData } from './$types';

	// eslint-disable-next-line @typescript-eslint/no-unused-vars
	let { data }: { data: PageData } = $props();

	let name = $state('');
	let email = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let error = $state('');
	let loading = $state(false);

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
		} catch (err: unknown) {
			error = parseApiError(err, 'Registration failed');
		} finally {
			loading = false;
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

				<PasswordInput
					id="password"
					name="password"
					label="Password"
					bind:value={password}
					error={fieldErrors.password}
					shake={shake.password}
					autocomplete="new-password"
					oninput={() => handlePasswordInput(password, confirmPassword)}
				/>

				<PasswordInput
					id="confirmPassword"
					name="confirmPassword"
					label="Confirm Password"
					bind:value={confirmPassword}
					error={fieldErrors.confirmPassword}
					shake={shake.confirmPassword}
					autocomplete="new-password"
					oninput={() => handleConfirmPasswordInput(password, confirmPassword)}
				/>
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
