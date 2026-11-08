<script lang="ts">
	import * as Sidebar from '$lib/components/ui/sidebar';
	import OrgSidebar from '$lib/components/orgSidebar.svelte';
	import OrgHeaderBar from '$lib/components/org/OrgHeaderBar.svelte';
	import { headerStore } from '$lib/stores/headerStore.svelte';
	import { activeOrgId, organizations } from '$lib/stores/organization';
	import { fade } from 'svelte/transition';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { invitationStore } from '$lib/stores/invitation.svelte';
	import type { LayoutData } from './$types';

	let { data, children }: { data: LayoutData; children: import('svelte').Snippet } = $props();

	const orgName = $derived(data?.org?.Name || data?.org?.name || 'Workspace');

	const pageTitle = $derived.by(() => {
		if (headerStore.title) return headerStore.title;
		const path = $page.url.pathname;
		if (path.includes('/monitors')) return 'Monitors';
		if (path.includes('/alert-templates')) return 'Alert Templates';
		if (path.includes('/members')) return 'Members';
		if (path.includes('/incidents')) return 'Incidents';
		if (path.includes('/statuspage')) return 'Status Pages';
		if (path.includes('/logs')) return 'Logs';
		if (path.includes('/notifications')) return 'Notifications';
		if (path.includes('/api-keys')) return 'API Keys';
		if (path.includes('/settings')) return 'Settings';
		return 'Dashboard';
	});

	// Safe client-side state sync
	$effect(() => {
		if (data?.allOrgs) {
			organizations.set(
				data.allOrgs.map(
					(o: {
						ID?: string;
						id?: string;
						Name?: string;
						name?: string;
						Slug?: string;
						slug?: string;
					}) => ({
						id: o.ID ?? o.id ?? '',
						name: o.Name ?? o.name ?? '',
						slug: o.Slug ?? o.slug ?? ''
					})
				)
			);
		}
		if (data?.orgId) {
			activeOrgId.set(data.orgId);
		}
	});

	onMount(() => {
		invitationStore.fetch();
	});
</script>

<Sidebar.Provider>
	<OrgSidebar
		user={data?.user}
		org={data?.org}
		allOrgs={data?.allOrgs ?? []}
		hasUnreadNotifications={invitationStore.hasPending}
	/>
	<Sidebar.Inset class="flex flex-col min-h-screen">
		<!-- ── Top Header Bar ─────────────────────────────────────────── -->
		<OrgHeaderBar
			{orgName}
			{pageTitle}
			showSearch={headerStore.showSearch}
			bind:searchQuery={headerStore.searchQuery}
			searchPlaceholder={headerStore.searchPlaceholder}
			actionType={headerStore.actionType}
			actionLabel={headerStore.actionLabel}
			onAction={headerStore.onAction}
		/>

		<!-- ── Main View Area (Full width) ─────────────────────────────────── -->
		<div class="flex-1 relative flex min-h-0">
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
		</div>
	</Sidebar.Inset>
</Sidebar.Provider>
