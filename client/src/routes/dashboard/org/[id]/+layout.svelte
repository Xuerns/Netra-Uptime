<script lang="ts">
	import type { LayoutData } from './$types';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import OrgSidebar from '$lib/components/orgSidebar.svelte';
	import OrgMembersPanel from '$lib/components/orgMembersPanel.svelte';

	import { Bell } from 'lucide-svelte';
	import { fade } from 'svelte/transition';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { OrgAPI } from '$lib/api/organization';
	import { invalidateAll } from '$app/navigation';
	import { parseApiError } from '$lib/utils/parseApiError';

	let { data, children } = $props();

	let hasUnreadNotifications = $state(false);
	let members = $state<any[]>([]);
	let isLoadingMembers = $state(false);
	let membersError = $state('');

	$effect(() => {
		members = data?.members ?? [];
	});

	const currentUserId = $derived(data?.user?.id || data?.user?.ID);

	onMount(async () => {
		try {
			const res = await OrgAPI.getInvitations();
			const invitations = res.data || res.message || res || [];
			hasUnreadNotifications = invitations.some((inv: any) => inv.Status === 'pending');
		} catch (e) {
			// ignore
		}
	});

	async function refreshMembers() {
		try {
			const res = await OrgAPI.getMembers(data.orgId);
			members = res.data ?? [];
		} catch (err: any) {
			membersError = parseApiError(err, 'Failed to reload members');
		}
	}
</script>

<Sidebar.Provider>
	<OrgSidebar user={data?.user} org={data?.org} allOrgs={data?.allOrgs ?? []} />
	<Sidebar.Inset class="flex flex-col min-h-screen">
		<header
			class="flex h-12 shrink-0 items-center justify-between gap-2 border-b px-4 border-border!"
		>
			<Sidebar.Trigger class="-ml-1" />

			<a
				href="/dashboard/notification"
				class="flex relative items-center justify-center size-9 rounded-lg text-muted-foreground hover:text-foreground transition-colors"
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

			<!-- Members Panel (Right Side) -->
			<OrgMembersPanel
				{members}
				{isLoadingMembers}
				{membersError}
				currentUserId={currentUserId ?? ''}
				orgId={data?.orgId ?? ''}
				onMembersUpdated={refreshMembers}
			/>
		</div>
	</Sidebar.Inset>
</Sidebar.Provider>
