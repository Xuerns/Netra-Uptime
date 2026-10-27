<script lang="ts">
	import * as Sidebar from '$lib/components/ui/sidebar';
	import AppSidebar from '$lib/components/appSidebar.svelte';

	import { Mail } from 'lucide-svelte';
	import { fade } from 'svelte/transition';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { invitationStore } from '$lib/stores/invitation.svelte';
	import type { LayoutData } from './$types';

	let { data, children }: { data: LayoutData; children: import('svelte').Snippet } = $props();

	const isOrgScoped = $derived(
		$page.url.pathname.startsWith('/dashboard/org/') &&
			$page.url.pathname.split('/').filter(Boolean).length > 2
	);

	onMount(() => {
		invitationStore.fetch();
	});
</script>

{#if isOrgScoped}
	{@render children()}
{:else}
	<Sidebar.Provider>
		<AppSidebar user={data?.user} />
		<Sidebar.Inset class="flex flex-col min-h-screen">
			<header
				class="flex h-12 shrink-0 items-center justify-between gap-2 border-b bg-card px-4 border-border!"
			>
				<Sidebar.Trigger class="-ml-1" />

				<a
					href="/dashboard/notification"
					class="flex relative items-center justify-center size-9 rounded-lg text-muted-foreground hover:text-foreground hover:bg-white/5 transition-colors"
					aria-label="Invitations"
					title="Invitations"
				>
					<Mail class="size-4.5" />
					{#if invitationStore.hasPending}
						<span class="absolute top-2 right-2 flex size-2">
							<span
								class="absolute inline-flex size-full animate-ping rounded-full bg-primary opacity-75"
							></span>
							<span class="relative inline-flex size-2 rounded-full bg-primary ring-2 ring-card"
							></span>
						</span>
					{/if}
				</a>
			</header>

			<main class="flex-1 relative flex flex-col min-h-0 dashboard-bg">
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
{/if}
