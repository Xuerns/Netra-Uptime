<script lang="ts">
	import Logo from '$lib/assets/Logo.svg';
	import { ArrowLeft, Eye, EyeOff, Mail, KeyRound, CheckCircle2, Loader2 } from 'lucide-svelte';
	import { env } from '$env/dynamic/public';
	import { page } from '$app/stores';
	import { validateEmail, validatePassword, validateConfirmPassword } from '$lib/utils/validation';
	import { debounce } from '$lib/helper/utils';
	import { parseApiError } from '$lib/utils/parseApiError';

	const API_Route = `${env.PUBLIC_API_URL}`;

	// Determine flow from URL token
	let token = $derived($page.url.searchParams.get('token'));
	let isResetMode = $derived(!!token);

	// Request flow state
	let email = $state('');
	let requestLoading = $state(false);
	let requestError = $state('');
	let requestSuccess = $state(false);

	let fieldErrors = $state({ email: '', newPassword: '', confirmPassword: '' });
	let shake = $state({ email: false, newPassword: false, confirmPassword: false });

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

	// Reset flow state
	let newPassword = $state('');
	let confirmPassword = $state('');
	let showPassword = $state(false);
	let showConfirmPassword = $state(false);
	let resetLoading = $state(false);
	let resetError = $state('');
	let resetSuccess = $state(false);

	const handleNewPasswordInput = debounce((val: string, confirm: string) => {
		const err = validatePassword(val);
		if (err) {
			fieldErrors.newPassword = err;
			shake.newPassword = false;
			setTimeout(() => (shake.newPassword = true), 10);
		} else {
			fieldErrors.newPassword = '';
			shake.newPassword = false;
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

	async function handleRequestReset(event: Event) {
		event.preventDefault();
		requestLoading = true;
		requestError = '';

		try {
			const response = await fetch(`${API_Route}/auth/forgotpassword`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ email })
			});

			const data = await response.json();

			if (!response.ok) {
				throw new Error(data.error || 'Failed to send reset link.');
			}

			requestSuccess = true;
		} catch (err: any) {
			requestError = parseApiError(err, 'Failed to send reset link.');
		} finally {
			requestLoading = false;
		}
	}

	async function handleResetPassword(event: Event) {
		event.preventDefault();
		resetError = '';

		if (newPassword !== confirmPassword) {
			resetError = 'Passwords do not match.';
			return;
		}

		if (newPassword.length < 8) {
			resetError = 'Password must be at least 8 characters.';
			return;
		}

		resetLoading = true;

		try {
			const response = await fetch(`${API_Route}/auth/resetpassword`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ token, password: newPassword })
			});

			const data = await response.json();

			if (!response.ok) {
				throw new Error(data.error || 'Failed to reset password.');
			}

			resetSuccess = true;
		} catch (err: any) {
			resetError = parseApiError(err, 'Failed to reset password.');
		} finally {
			resetLoading = false;
		}
	}
</script>

<svelte:head>
	<title>{isResetMode ? 'Reset Password' : 'Forgot Password'} | Netra Uptime</title>
</svelte:head>

<main
	class="min-h-screen flex items-center justify-center p-sm md:p-md text-foreground antialiased w-full relative auth-bg"
>
	<a
		href="/auth/login"
		class="absolute top-6 left-6 md:top-8 md:left-8 flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors font-body-md group focus:outline-none focus:ring-2 focus:ring-primary rounded-lg px-2 py-1"
	>
		<ArrowLeft class="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
		<span>Back to login</span>
	</a>

	<div class="w-full max-w-[448px] bg-card rounded-xl border border-border p-lg shadow-sm">
		{#if isResetMode}
			<!-- ═══════════════════════════════════════ -->
			<!-- RESET PASSWORD FLOW                     -->
			<!-- ═══════════════════════════════════════ -->
			{#if resetSuccess}
				<!-- Success state -->
				<div class="flex flex-col items-center text-center py-md space-y-md animate-fade-in">
					<div class="w-14 h-14 rounded-full bg-primary/10 flex items-center justify-center">
						<CheckCircle2 class="w-7 h-7 text-primary" />
					</div>
					<div>
						<h1
							class="font-headline-lg text-[24px] font-semibold text-foreground tracking-tight"
							style="text-wrap: balance;"
						>
							Password updated
						</h1>
						<p class="font-body-md text-muted-foreground mt-xs">
							Your password has been changed successfully. You can now log in with your new
							password.
						</p>
					</div>
					<a
						href="/auth/login"
						class="w-full bg-primary hover:bg-primary/90 text-primary-foreground font-label-md font-semibold py-2.5 px-4 rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 focus:ring-offset-background text-center block"
					>
						Back to login
					</a>
				</div>
			{:else}
				<!-- Reset form -->
				<form class="space-y-md" onsubmit={handleResetPassword}>
					<div class="flex flex-col items-center mb-lg">
						<div class="mb-sm">
							<div class="w-14 h-14 rounded-full bg-primary/10 flex items-center justify-center">
								<KeyRound class="w-7 h-7 text-primary" />
							</div>
						</div>
						<h1
							class="font-headline-lg text-[24px] font-semibold text-foreground tracking-tight"
							style="text-wrap: balance;"
						>
							Set new password
						</h1>
						<p class="font-body-md text-muted-foreground text-center mt-xs">
							Choose a strong password for your account.
						</p>
					</div>

					<!-- Error -->
					{#if resetError}
						<div
							class="flex items-start gap-2 p-3 rounded-lg bg-destructive/10 border border-destructive/20 animate-fade-in"
						>
							<p class="font-body-md text-sm text-destructive">{resetError}</p>
						</div>
					{/if}

					<!-- Fields -->
					<div class="space-y-sm">
						<div class="flex flex-col gap-1">
							<div class="flex items-center gap-3">
								<label class="font-label-md text-muted-foreground" for="new-password">
									New Password
								</label>
								{#if fieldErrors.newPassword}
									<span class="text-xs text-destructive font-medium">{fieldErrors.newPassword}</span
									>
								{/if}
							</div>
							<div class="relative">
								<input
									bind:value={newPassword}
									oninput={() => handleNewPasswordInput(newPassword, confirmPassword)}
									class="w-full rounded-lg px-3 py-2.5 bg-background border border-border text-foreground font-body-md placeholder:text-muted-foreground/50 focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary transition-colors pr-10 {fieldErrors.newPassword
										? 'border-destructive/70 focus:border-destructive focus:ring-destructive'
										: ''} {shake.newPassword ? 'animate-shake' : ''}"
									id="new-password"
									name="new-password"
									placeholder="••••••••"
									required
									minlength="8"
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
								<label class="font-label-md text-muted-foreground" for="confirm-password">
									Confirm Password
								</label>
								{#if fieldErrors.confirmPassword}
									<span class="text-xs text-destructive font-medium"
										>{fieldErrors.confirmPassword}</span
									>
								{/if}
							</div>
							<div class="relative">
								<input
									bind:value={confirmPassword}
									oninput={() => handleConfirmPasswordInput(newPassword, confirmPassword)}
									class="w-full rounded-lg px-3 py-2.5 bg-background border border-border text-foreground font-body-md placeholder:text-muted-foreground/50 focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary transition-colors pr-10 {fieldErrors.confirmPassword
										? 'border-destructive/70 focus:border-destructive focus:ring-destructive'
										: ''} {shake.confirmPassword ? 'animate-shake' : ''}"
									id="confirm-password"
									name="confirm-password"
									placeholder="••••••••"
									required
									minlength="8"
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
							class="w-full bg-primary hover:bg-primary/90 text-primary-foreground font-label-md font-semibold py-2.5 px-4 rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 focus:ring-offset-background flex items-center justify-center gap-2"
							type="submit"
							disabled={resetLoading}
						>
							{#if resetLoading}
								<Loader2 class="w-4 h-4 animate-spin" />
								Resetting...
							{:else}
								Reset password
							{/if}
						</button>
					</div>
				</form>
			{/if}
		{:else}
			<!-- ═══════════════════════════════════════ -->
			<!-- REQUEST RESET FLOW                      -->
			<!-- ═══════════════════════════════════════ -->
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
							<span class="text-foreground font-medium">{email}</span>. Check your inbox and follow
							the link to reset your password.
						</p>
					</div>
					<div class="w-full space-y-3 pt-xs">
						<button
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
		{/if}
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
