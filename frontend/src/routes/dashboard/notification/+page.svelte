<script lang="ts">
	import { onMount } from 'svelte';
	import { OrgAPI } from '$lib/api/organization';
	import { Check, X, Bell } from 'lucide-svelte';
	import { invalidateAll } from '$app/navigation';
	import NotificationSkeleton from '$lib/components/skeletons/NotificationSkeleton.svelte';
	import { toast } from 'svelte-sonner';

	interface User {
		id: string;
		name: string;
		email: string;
		avatar_url: string;
	}

	interface Organization {
		id: string;
		name: string;
		slug: string;
	}

	interface Invitation {
		ID: string;
		OrganizationID: string;
		Role: string;
		Status: string;
		CreatedAt: string;
		ExpiresAt: string;
		Organization: Organization;
		Inviter: User;
	}

	let invitations: Invitation[] = $state([]);
	let isLoading = $state(true);
	let errorMsg = $state('');

	onMount(async () => {
		await loadInvitations();
	});

	async function loadInvitations() {
		isLoading = true;
		errorMsg = '';
		try {
			const res = await OrgAPI.getInvitations();
			const raw = Array.isArray(res.data) ? res.data : [];
			invitations = raw.map((inv: any) => ({
				ID: inv.id ?? inv.ID,
				Role: inv.role ?? inv.Role,
				Status: inv.status ?? inv.Status,
				Organization: {
					id: inv.organization?.id,
					name: inv.organization?.name ?? inv.Organization?.name,
					slug: inv.organization?.slug ?? inv.Organization?.slug
				},
				Inviter: {
					id: inv.inviter?.id,
					name: inv.inviter?.name ?? inv.Inviter?.name,
					avatar_url: inv.inviter?.avatar_url ?? inv.Inviter?.avatar_url
				}
			}));
		} catch (err: any) {
			errorMsg = err.message || 'Failed to load invitations';
		} finally {
			isLoading = false;
		}
	}

	async function handleAction(id: string, action: 'accept' | 'reject') {
		try {
			await OrgAPI.actionInvitation(id, action);
			// Refresh list after action
			await loadInvitations();
			// Also invalidate layout data to refresh org list in sidebar if accepted
			if (action === 'accept') {
				invalidateAll();
				toast.success('Invitation accepted');
			} else {
				toast.success('Invitation rejected');
			}
		} catch (err: any) {
			errorMsg = err.message || `Failed to ${action} invitation`;
			toast.error(errorMsg);
		}
	}

	function avatarColor(name: string): string {
		const colors = [
			'bg-red-500/20 text-red-500',
			'bg-blue-500/20 text-blue-500',
			'bg-green-500/20 text-green-500',
			'bg-yellow-500/20 text-yellow-500',
			'bg-purple-500/20 text-purple-500',
			'bg-pink-500/20 text-pink-500',
			'bg-indigo-500/20 text-indigo-500'
		];
		let hash = 0;
		for (let i = 0; i < name.length; i++) {
			hash = name.charCodeAt(i) + ((hash << 5) - hash);
		}
		return colors[Math.abs(hash) % colors.length];
	}

	function initials(name: string): string {
		if (!name) return 'U';
		const parts = name.trim().split(/\s+/);
		if (parts.length >= 2) {
			return (parts[0][0] + parts[1][0]).toUpperCase();
		}
		return name.substring(0, 2).toUpperCase();
	}
</script>

<svelte:head>
	<title>Notification • PulseWatch</title>
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
			<NotificationSkeleton />
		{:else if !invitations || invitations.length === 0}
			<div
				class="flex flex-col items-center justify-center py-12 px-4 rounded-xl border border-border border-dashed bg-white/1"
			>
				<div
					class="flex items-center justify-center size-12 rounded-full bg-white/5 text-muted-foreground/50 mb-4"
				>
					<Bell class="size-5" />
				</div>
				<p class="text-[14px] font-medium text-foreground">No pending invitations</p>
				<p class="text-[13px] text-muted-foreground mt-1 text-center">
					When you are invited to join an organization, it will appear here.
				</p>
			</div>
		{:else}
			{#each invitations as invite (invite.ID)}
				<div
					class="flex flex-wrap items-center justify-between gap-6 p-5 rounded-xl bg-white/2 border border-border hover:bg-white/4 transition-colors"
				>
					<div class="flex items-center gap-4 min-w-0">
						<!-- Inviter Avatar -->
						{#if invite.Inviter?.avatar_url}
							<img
								src={invite.Inviter.avatar_url}
								alt={invite.Inviter.name}
								class="size-10 rounded-full object-cover shrink-0"
							/>
						{:else}
							<div
								class="size-10 rounded-full flex items-center justify-center font-semibold text-[14px] shrink-0 {avatarColor(
									invite.Inviter?.name || 'U'
								)}"
							>
								{initials(invite.Inviter?.name || 'U')}
							</div>
						{/if}

						<div class="flex flex-col min-w-0">
							<p class="text-[14.5px] text-foreground m-0 leading-snug">
								<span class="font-semibold">{invite.Inviter?.name || 'Someone'}</span>
								<span class="text-muted-foreground">invited you to join</span>
								<span class="font-semibold text-primary"
									>{invite.Organization?.name || 'an organization'}</span
								>
							</p>
							<p class="text-[12.5px] text-muted-foreground m-0 mt-1">
								Role: <span class="capitalize text-foreground">{invite.Role}</span>
							</p>
						</div>
					</div>

					{#if invite.Status === 'pending'}
						<div class="flex items-center gap-2 shrink-0 ml-auto">
							<button
								class="flex items-center justify-center size-9 rounded-lg border border-border bg-white/2 text-muted-foreground hover:bg-destructive/15 hover:text-destructive hover:border-destructive/20 transition-all"
								onclick={() => handleAction(invite.ID, 'reject')}
								aria-label="Reject"
								title="Reject invitation"
							>
								<X class="size-[18px]" />
							</button>
							<button
								class="flex items-center gap-2 h-9 px-4 rounded-lg bg-primary text-primary-foreground font-medium text-[13.5px] hover:bg-[#12c98d] transition-colors"
								onclick={() => handleAction(invite.ID, 'accept')}
							>
								<Check class="size-sm" />
								<span>Accept</span>
							</button>
						</div>
					{:else}
						<div class="shrink-0 ml-auto">
							<span
								class="text-[12px] font-medium px-3 py-1 rounded-full uppercase tracking-wider
								{invite.Status === 'accepted'
									? 'bg-primary/10 text-primary'
									: 'bg-destructive/10 text-destructive'}"
							>
								{invite.Status}
							</span>
						</div>
					{/if}
				</div>
			{/each}
		{/if}
	</div>
</div>
