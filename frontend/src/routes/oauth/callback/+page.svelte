<script lang="ts">
	import { onMount } from 'svelte';
	import { env } from '$env/dynamic/public';

	let error = $state('');

	onMount(async () => {
		const urlParams = new URLSearchParams(window.location.search);
		const code = urlParams.get('code');
		const provider = 'google'; // Currently only supporting Google

		if (!code) {
			error = 'Authorization code not found in URL.';
			return;
		}

		try {
			const API_Route = `${env.PUBLIC_API_URL}`;
			const response = await fetch(`${API_Route}/auth`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ provider, code }),
				credentials: 'include'
			});

			const data = await response.json();

			if (!response.ok) {
				throw new Error(data.error || data.message || 'Failed to authenticate with provider');
			}

			// Redirect to dashboard on success
			window.location.href = '/dashboard/monitor';
		} catch (err: any) {
			console.error(err);
			error = err.message || 'An error occurred during authentication';
		}
	});
</script>

<svelte:head>
	<title>Authenticating... • PulseWatch</title>
</svelte:head>

<main
	class="min-h-screen flex items-center justify-center p-sm md:p-md text-on-surface antialiased w-full"
>
	<div
		class="w-full max-w-[448px] bg-[#131314] rounded-xl border border-[#2a2a2b] p-lg shadow-sm text-center"
	>
		{#if error}
			<div class="flex flex-col items-center gap-4">
				<div
					class="w-12 h-12 rounded-full bg-red-500/20 flex items-center justify-center text-red-500"
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
				<h1 class="font-headline-sm font-semibold text-on-surface">Authentication Failed</h1>
				<p class="font-body-md text-on-surface-variant">{error}</p>
				<a
					href="/auth/login"
					class="mt-4 px-4 py-2 bg-primary text-on-primary rounded-lg font-label-md hover:bg-primary-container transition-colors"
				>
					Return to Login
				</a>
			</div>
		{:else}
			<div class="flex flex-col items-center gap-4">
				<div
					class="w-12 h-12 border-4 border-[#2a2a2b] border-t-primary rounded-full animate-spin"
				></div>
				<h1 class="font-headline-sm font-semibold text-on-surface">Authenticating</h1>
				<p class="font-body-md text-on-surface-variant">
					Please wait while we verify your credentials...
				</p>
			</div>
		{/if}
	</div>
</main>

<style>
	main {
		background-color: #0e0e0f;
		background-image: radial-gradient(rgba(1, 101, 53, 0.462) 1.5px, transparent 1.5px);
		background-size: 24px 24px;
	}
</style>
