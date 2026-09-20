<script lang="ts">
	import { onMount } from 'svelte';
	import { env } from '$env/dynamic/public';
	import { parseApiError } from '$lib/utils/parseApiError';

	let error = $state('');

	onMount(async () => {
		const urlParams = new URLSearchParams(window.location.search);
		const code = urlParams.get('code');
		const provider = 'google'; // Currently only supporting Google

		if (!code) {
			error = 'Authorization code not found in URL.';
			return;
		}

		// Verify OAuth state parameter
		const state = urlParams.get('state');
		const savedState = sessionStorage.getItem('oauth_state');
		if (!state || state !== savedState) {
			error = 'Invalid OAuth state. Please try logging in again.';
			return;
		}
		sessionStorage.removeItem('oauth_state');

		// Timeout handling
		const controller = new AbortController();
		const timeout = setTimeout(() => controller.abort(), 15000);

		try {
			const API_Route = `${env.PUBLIC_API_URL}`;
			const response = await fetch(`${API_Route}/auth`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ provider, code }),
				credentials: 'include',
				signal: controller.signal
			});

			clearTimeout(timeout);
			const data = await response.json();

			if (!response.ok) {
				throw new Error(data.error || data.message || 'Failed to authenticate with provider');
			}

			// Redirect to dashboard on success
			window.location.href = '/dashboard/org';
		} catch (err: any) {
			clearTimeout(timeout);
			console.error(err);
			if (err.name === 'AbortError') {
				error = 'Authentication timed out. Please try again.';
			} else {
				error = parseApiError(err, 'An error occurred during authentication');
			}
		}
	});
</script>

<svelte:head>
	<title>Authenticating... | PulseWatch</title>
</svelte:head>

<main
	class="min-h-screen flex items-center justify-center p-sm md:p-md text-foreground antialiased w-full auth-bg"
>
	<div
		class="w-full max-w-[448px] bg-card rounded-xl border border-border p-lg shadow-sm text-center"
	>
		{#if error}
			<div class="flex flex-col items-center gap-4">
				<div
					class="w-12 h-12 rounded-full bg-destructive/20 flex items-center justify-center text-destructive"
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						width="24"
						height="24"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
						><circle cx="12" cy="12" r="10" /><line x1="12" y1="8" x2="12" y2="12" /><line
							x1="12"
							y1="16"
							x2="12.01"
							y2="16"
						/></svg
					>
				</div>
				<h1 class="font-headline-sm font-semibold text-foreground">Authentication Failed</h1>
				<p class="font-body-md text-muted-foreground">{error}</p>
				<a
					href="/auth/login"
					class="mt-4 px-4 py-2 bg-primary text-primary-foreground rounded-lg font-label-md hover:bg-primary/90 transition-colors"
				>
					Return to Login
				</a>
			</div>
		{:else}
			<div class="flex flex-col items-center gap-4">
				<div
					class="w-12 h-12 border-4 border-border border-t-primary rounded-full animate-spin"
				></div>
				<h1 class="font-headline-sm font-semibold text-foreground">Authenticating</h1>
				<p class="font-body-md text-muted-foreground">
					Please wait while we verify your credentials...
				</p>
			</div>
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
		background-image: radial-gradient(rgba(1, 101, 53, 0.462) 1.5px, transparent 1.5px);
	}
</style>
