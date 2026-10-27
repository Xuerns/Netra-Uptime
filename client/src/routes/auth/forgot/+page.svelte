<script lang="ts">
	import { ArrowLeft } from 'lucide-svelte';
	import { page } from '$app/stores';
	import { ForgotPasswordCard, ResetPasswordCard } from '$lib/components/auth';
	import type { PageData } from './$types';

	// eslint-disable-next-line @typescript-eslint/no-unused-vars
	let { data }: { data: PageData } = $props();

	// Determine flow from URL token
	let token = $derived($page.url.searchParams.get('token'));
	let isResetMode = $derived(!!token);
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
		{#if isResetMode && token}
			<ResetPasswordCard {token} />
		{:else}
			<ForgotPasswordCard />
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
