<script lang="ts">
	import * as Sidebar from '$lib/components/ui/sidebar';
	import AppSidebar from '$lib/components/appSidebar.svelte';
	import OrgHeaderBar from '$lib/components/org/OrgHeaderBar.svelte';
	import { headerStore } from '$lib/stores/headerStore.svelte';
	import { beforeNavigate } from '$app/navigation';
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

	const pageTitle = $derived.by(() => {
		if (headerStore.title) return headerStore.title;
		const path = $page.url.pathname;
		if (path.startsWith('/dashboard/org')) return 'Organizations';
		if (path.startsWith('/dashboard/notification')) return 'Invitations';
		if (path.startsWith('/dashboard/profile')) return 'Profile';
		return 'Dashboard';
	});

	// Reset header state before non-org navigation so the new page starts clean.
	beforeNavigate(({ to }) => {
		if (to?.url.pathname.startsWith('/dashboard/org/')) return;
		if (to?.url.pathname !== '/dashboard/org') {
			headerStore.reset();
		}
	});

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
			<!-- Top Header Bar -->
			<OrgHeaderBar
				orgName="Workspace"
				breadcrumbHref="/dashboard/org"
				{pageTitle}
				showSearch={headerStore.showSearch}
				bind:searchQuery={headerStore.searchQuery}
				searchPlaceholder={headerStore.searchPlaceholder}
				actionType={headerStore.actionType}
				actionLabel={headerStore.actionLabel}
				onAction={headerStore.onAction}
			/>

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
