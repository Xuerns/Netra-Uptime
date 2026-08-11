<script lang="ts">
	import Logo from '$lib/assets/Logo.svg';
	import { ArrowLeft, Eye, EyeOff } from 'lucide-svelte';
	import { env } from '$env/dynamic/public';
	import { validateEmail, validatePassword, debounce } from '$lib/utils/validation';

	let email = $state('');
	let password = $state('');
	let showPassword = $state(false);
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

	const API_Route = `${env.PUBLIC_API_URL}`;

	async function handleLogin(event: Event) {
		event.preventDefault();
		loading = true;
		error = '';

		try {
			const response = await fetch(`${API_Route}/auth/login`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ email, password }),
				credentials: 'include'
			});

			const data = await response.json();

			if (!response.ok) {
				throw new Error(data.message || 'Gagal Login');
			}

			window.location.href = '/dashboard/monitor';
		} catch (err: any) {
			error = err.message;
		} finally {
			loading = false;
		}
	}

	function loginWithGoogle() {
		const clientId = env.PUBLIC_GOOGLE_CLIENT_ID;
		const redirectUri = env.PUBLIC_GOOGLE_REDIRECT_URL;
		const params = new URLSearchParams({
			client_id: clientId,
			redirect_uri: redirectUri,
			response_type: 'code',
			scope: 'email profile'
		});
		window.location.href = `https://accounts.google.com/o/oauth2/v2/auth?${params.toString()}`;
	}
</script>

<svelte:head>
	<title>Login • PulseWatch</title>
</svelte:head>

<main
	class="min-h-screen flex items-center justify-center p-sm md:p-md text-on-surface antialiased w-full relative"
>
	<a
		href="/"
		class="absolute top-6 left-6 md:top-8 md:left-8 flex items-center gap-2 text-on-surface-variant hover:text-on-surface transition-colors font-body-md group focus:outline-none focus:ring-2 focus:ring-primary rounded-lg px-2 py-1"
	>
		<ArrowLeft class="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
		<span>Back</span>
	</a>
	<div class="w-full max-w-[448px] bg-[#131314] rounded-xl border border-[#2a2a2b] p-lg shadow-sm animate-fade-in">
		<form class="space-y-md" onsubmit={handleLogin}>
			<div class="flex flex-col items-center mb-lg">
				<div class="mb-sm">
					<img alt="Pulsewatch Logo" class="h-10 w-auto" src={Logo} />
				</div>
				<h1
					class="font-headline-lg text-[24px] font-semibold text-on-surface tracking-tight"
					style="text-wrap: balance;"
				>
					Welcome back
				</h1>
				<p class="font-body-md text-on-surface-variant text-center mt-xs">
					Log in to monitor your systems.
				</p>
			</div>

			<!-- Fields -->
			<div class="space-y-sm">
				<div class="flex flex-col gap-1">
					<div class="flex items-center gap-3">
						<label class="font-label-md text-on-surface-variant" for="email"> Work Email </label>
						{#if fieldErrors.email}
							<span class="text-xs text-error font-medium">{fieldErrors.email}</span>
						{/if}
					</div>
					<input
						bind:value={email}
						oninput={() => handleEmailInput(email)}
						class="w-full rounded-lg px-3 py-2.5 bg-[#0e0e0f] border border-[#2a2a2b] text-on-surface font-body-md placeholder:text-on-surface-variant/50 focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary transition-colors {fieldErrors.email ? 'border-error/70 focus:border-error focus:ring-error' : ''} {shake.email ? 'animate-shake' : ''}"
						id="email"
						name="email"
						placeholder="john@company.com"
						required
						type="email"
					/>
				</div>
				<div class="flex flex-col gap-1">
					<div class="flex justify-between items-center">
						<div class="flex items-center gap-3">
							<label class="font-label-md text-on-surface-variant" for="password"> Password </label>
							{#if fieldErrors.password}
								<span class="text-xs text-error font-medium">{fieldErrors.password}</span>
							{/if}
						</div>
						<a
							href="/auth/forgot"
							class="text-sm font-label-md text-primary hover:underline underline-offset-4"
							>Forgot password?</a
						>
					</div>
					<div class="relative">
						<input
							bind:value={password}
							oninput={() => handlePasswordInput(password)}
							class="w-full rounded-lg px-3 py-2.5 bg-[#0e0e0f] border border-[#2a2a2b] text-on-surface font-body-md placeholder:text-on-surface-variant/50 focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary transition-colors pr-10 {fieldErrors.password ? 'border-error/70 focus:border-error focus:ring-error' : ''} {shake.password ? 'animate-shake' : ''}"
							id="password"
							name="password"
							placeholder="••••••••"
							required
							type={showPassword ? 'text' : 'password'}
						/>
						<button
							type="button"
							class="absolute inset-y-0 right-0 pr-3 flex items-center text-on-surface-variant hover:text-on-surface transition-colors focus:outline-none"
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
			</div>

			<!-- Submit -->
			<div class="pt-2">
				<button
					class="w-full bg-primary hover:bg-primary-container text-on-primary font-label-md font-semibold py-2.5 px-4 rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 focus:ring-offset-[#131314]"
					type="submit"
					onclick={handleLogin}
					disabled={loading}
				>
					{#if loading}Loading...{:else}Login{/if}
				</button>
			</div>

			<!-- Separator -->
			<div class="relative py-2 flex items-center">
				<div class="grow h-px bg-[#2a2a2b]"></div>
				<span class="shrink-0 mx-4 text-label-md text-on-surface-variant">or</span>
				<div class="grow h-px bg-[#2a2a2b]"></div>
			</div>

			<!-- Google Auth Button -->
			<button
				type="button"
				onclick={loginWithGoogle}
				class="w-full bg-transparent hover:bg-white/5 text-on-surface font-label-md py-2.5 px-4 rounded-lg border border-[#2a2a2b] transition-colors flex items-center justify-center gap-2 focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 focus:ring-offset-[#131314]"
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
		</form>

		<p class="mt-lg text-center font-body-md text-sm text-on-surface-variant">
			Don't have an account? <a
				class="text-primary hover:underline underline-offset-4"
				href="/auth/register">Sign up</a
			>
		</p>
	</div>
</main>

<style>
	main {
		background-color: #0e0e0f;
		background-image: radial-gradient(rgba(1, 101, 53, 0.15) 1.5px, transparent 1.5px);
		background-size: 24px 24px;
	}
</style>
