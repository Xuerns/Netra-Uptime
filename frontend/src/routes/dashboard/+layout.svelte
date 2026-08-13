<script lang="ts">
	import type { LayoutData } from './$types';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import AppSidebar from '$lib/components/appSidebar.svelte';

	import { Bell } from 'lucide-svelte';
	import { fade } from 'svelte/transition';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { OrgAPI } from '$lib/api/organization';

	let { data, children } = $props();

	let hasUnreadNotifications = $state(false);

	onMount(async () => {
		try {
			const res = await OrgAPI.getInvitations();
			const invitations = res.data || res.message || res || [];
			hasUnreadNotifications = invitations.some((inv: any) => inv.Status === 'pending');
		} catch (e) {
			// ignore
		}
	});
</script>

<Sidebar.Provider>
	<AppSidebar user={data?.user} />
	<Sidebar.Inset class="flex flex-col min-h-screen">
		<header
			class="flex h-12 shrink-0 items-center justify-between gap-2 border-b px-4 border-[rgba(255,255,255,0.08)]!"
		>
			<Sidebar.Trigger class="-ml-1" />

			<a
				href="/dashboard/notification"
				class="flex relative items-center justify-center size-9 rounded-lg text-on-surface-variant hover:text-on-surface transition-colors"
				aria-label="Notifications"
			>
				<Bell class="size-4.5" />
				{#if hasUnreadNotifications}
					<span
						class="absolute top-2.25 right-2.25 size-2 rounded-full bg-primary border border-background"
					></span>
				{/if}
			</a>
		</header>

		<main class="flex-1 relative flex flex-col min-h-0">
			{#key $page.url.pathname}
				<div
					in:fade={{ duration: 150, delay: 150 }}
					out:fade={{ duration: 150 }}
					class="flex-1 p-6 overflow-y-auto"
				>
					{@render children()}
				</div>
			{/key}
		</main>
	</Sidebar.Inset>
</Sidebar.Provider>

<style>
	main {
		background-color: #131314;
		opacity: 1;
		background-image:
			conic-gradient(
				from 60deg at 56.25% calc(425% / 6),
				transparent 0deg,
				transparent 0.5deg,
				rgba(255, 255, 255, 0.015) 0.5deg,
				rgba(255, 255, 255, 0.015) 119.5deg,
				transparent 119.5deg,
				transparent 120deg
			),
			conic-gradient(
				from 60deg at 56.25% calc(425% / 6),
				transparent 0deg,
				transparent 0.5deg,
				rgba(255, 255, 255, 0.015) 0.5deg,
				rgba(255, 255, 255, 0.015) 119.5deg,
				transparent 119.5deg,
				transparent 120deg
			),
			conic-gradient(
				from 180deg at 43.75% calc(425% / 6),
				transparent 0deg,
				transparent 0.5deg,
				rgba(255, 255, 255, 0.015) 0.5deg,
				rgba(255, 255, 255, 0.015) 119.5deg,
				transparent 119.5deg,
				transparent 120deg
			),
			conic-gradient(
				from 180deg at 43.75% calc(425% / 6),
				transparent 0deg,
				transparent 0.5deg,
				rgba(255, 255, 255, 0.015) 0.5deg,
				rgba(255, 255, 255, 0.015) 119.5deg,
				transparent 119.5deg,
				transparent 120deg
			),
			conic-gradient(
				from -60deg at 50% calc(175% / 12),
				transparent 0deg,
				transparent 0.5deg,
				rgba(255, 255, 255, 0.015) 0.5deg,
				rgba(255, 255, 255, 0.015) 119.5deg,
				transparent 119.5deg,
				transparent 120deg
			),
			conic-gradient(
				from -60deg at 50% calc(175% / 12),
				transparent 0deg,
				transparent 0.5deg,
				rgba(255, 255, 255, 0.015) 0.5deg,
				rgba(255, 255, 255, 0.015) 119.5deg,
				transparent 119.5deg,
				transparent 120deg
			);
		background-position:
			0 0,
			6px 10px,
			0 0,
			6px 10px,
			6px 0,
			0 10px;
		background-size: 12px 20px;
	}
</style>
