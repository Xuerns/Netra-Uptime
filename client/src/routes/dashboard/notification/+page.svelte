<script lang="ts">
	import { onMount } from 'svelte';
	import { OrgAPI, type OrganizationInvitationItem } from '$lib/api/organization';
	import { Mail } from 'lucide-svelte';
	import { invalidateAll } from '$app/navigation';
	import { DataSkeleton } from '$lib/components/common';
	import { toast } from 'svelte-sonner';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { invitationStore } from '$lib/stores/invitation.svelte';
	import { headerStore } from '$lib/stores/headerStore.svelte';
	import InvitationCard, {
		type Invitation
	} from '$lib/components/notification/InvitationCard.svelte';

	let invitations: Invitation[] = $state([]);
	let isLoading = $state(true);
	let errorMsg = $state('');

	$effect(() => {
		headerStore.reset();
	});

	onMount(async () => {
		await loadInvitations();
	});

	async function loadInvitations() {
		isLoading = true;
		errorMsg = '';
		try {
			const res = await OrgAPI.getInvitations();
			const raw: OrganizationInvitationItem[] = Array.isArray(res.data)
				? res.data
				: Array.isArray(res)
					? (res as OrganizationInvitationItem[])
					: [];
			invitations = raw.map((inv: OrganizationInvitationItem) => ({
				ID: String(inv.id ?? inv.ID ?? ''),
				Role: String(inv.role ?? inv.Role ?? 'member'),
				Status: String(inv.status ?? inv.Status ?? '').toLowerCase(),
				Organization: {
					name: String(inv.organization?.name ?? inv.Organization?.name ?? 'Organization')
				},
				Inviter: {
					name: String(inv.inviter?.name ?? inv.Inviter?.name ?? 'User'),
					avatar_url: inv.inviter?.avatar_url ?? inv.Inviter?.avatar_url ?? undefined
				}
			}));
			await invitationStore.fetch();
		} catch (err: unknown) {
			errorMsg = parseApiError(err, 'Failed to load invitations');
		} finally {
			isLoading = false;
		}
	}

	async function handleAction(id: string, action: 'accept' | 'reject') {
		try {
			await OrgAPI.actionInvitation(id, action);
			// Refresh list after action
			await loadInvitations();
			await invitationStore.fetch();
			// Also invalidate layout data to refresh org list in sidebar if accepted
			if (action === 'accept') {
				invalidateAll();
				toast.success('Invitation accepted');
			} else {
				toast.success('Invitation rejected');
			}
		} catch (err: unknown) {
			errorMsg = parseApiError(err, `Failed to ${action} invitation`);
			toast.error(errorMsg);
		}
	}
</script>

<svelte:head>
	<title>Invitations | Netra Uptime</title>
</svelte:head>

<div class="w-full text-foreground antialiased py-4">
	{#if errorMsg}
		<div
			class="p-4 mb-6 rounded-lg bg-destructive/10 border border-destructive/20 text-destructive text-[13px]"
		>
			{errorMsg}
		</div>
	{/if}

	<div class="flex flex-col gap-4">
		<h2 class="text-sm font-medium text-muted-foreground uppercase tracking-wider mb-2">
			Organization Invitations
		</h2>

		{#if isLoading}
			<DataSkeleton variant="list" count={3} />
		{:else if !invitations || invitations.length === 0}
			<div
				class="flex flex-col items-center justify-center py-12 px-4 rounded-lg border border-border border-dashed bg-card"
			>
				<div
					class="flex items-center justify-center size-12 rounded-full bg-white/5 text-muted-foreground/50 mb-4"
				>
					<Mail class="size-5" />
				</div>
				<p class="text-[14px] font-medium text-foreground">No pending invitations</p>
				<p class="text-[13px] text-muted-foreground mt-1 text-center">
					When you are invited to join an organization, it will appear here.
				</p>
			</div>
		{:else}
			{#each invitations as invite (invite.ID)}
				<InvitationCard
					invitation={invite}
					onAccept={(id) => handleAction(id, 'accept')}
					onReject={(id) => handleAction(id, 'reject')}
				/>
			{/each}
		{/if}
	</div>
</div>
