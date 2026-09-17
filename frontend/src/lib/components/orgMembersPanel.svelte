<script lang="ts">
	import { MoreVertical, Trash2, ShieldAlert, ShieldX, Users, ChevronRight, UserPlus } from 'lucide-svelte';
	import MemberSkeleton from '$lib/components/skeletons/MemberSkeleton.svelte';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';
	import InviteMemberModal from '$lib/components/InviteMemberModal.svelte';
	import { OrgAPI } from '$lib/api/organization';
	import { toast } from 'svelte-sonner';

	interface MemberUser {
		ID: string;
		Name: string;
		AvatarURL?: string;
	}

	interface Member {
		ID: string;
		OrganizationID: string;
		UserID: string;
		Role: string;
		User: MemberUser;
	}

	let {
		members = [],
		isLoadingMembers = false,
		membersError = '',
		currentUserId = '',
		orgId = '',
		onMembersUpdated
	} = $props<{
		members: Member[];
		isLoadingMembers: boolean;
		membersError: string;
		currentUserId: string;
		orgId: string;
		onMembersUpdated?: () => void;
	}>();

	let panelOpen = $state(true);
	let openDropdownId = $state<string | null>(null);
	let isInviteModalOpen = $state(false);

	let owners = $derived(members.filter((m: Member) => m.Role === 'owner'));
	let admins = $derived(members.filter((m: Member) => m.Role === 'admin'));
	let regularMembers = $derived(members.filter((m: Member) => m.Role === 'member'));
	let currentUserRole = $derived(members.find((m: Member) => m.UserID === currentUserId)?.Role);

	let isConfirmKickOpen = $state(false);
	let memberToKick = $state<string | null>(null);

	function toggleDropdown(id: string) {
		openDropdownId = openDropdownId === id ? null : id;
	}

	function initials(name: string): string {
		const parts = name.trim().split(/\s+/).filter(Boolean);
		if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
		return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
	}

	function avatarColor(name: string): string {
		const colors = [
			'bg-emerald-500/20 text-emerald-400',
			'bg-blue-500/20 text-blue-400',
			'bg-violet-500/20 text-violet-400',
			'bg-orange-500/20 text-orange-400',
			'bg-rose-500/20 text-rose-400',
			'bg-cyan-500/20 text-cyan-400'
		];
		let hash = 0;
		for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash);
		return colors[Math.abs(hash) % colors.length];
	}

	async function handleKickMember(userId: string) {
		try {
			await OrgAPI.kickMember(orgId, userId);
			toast.success('Member removed successfully');
			onMembersUpdated?.();
		} catch (err: any) {
			toast.error(err.message || 'Failed to remove member');
		}
	}

	async function handlePromoteMember(userId: string) {
		try {
			await OrgAPI.promoteToAdmin(orgId, userId);
			toast.success('Member promoted to admin');
			onMembersUpdated?.();
		} catch (err: any) {
			toast.error(err.message || 'Failed to promote member');
		}
	}

	async function handleDemoteMember(userId: string) {
		try {
			await OrgAPI.demoteToMember(orgId, userId);
			toast.success('Admin demoted to member');
			onMembersUpdated?.();
		} catch (err: any) {
			toast.error(err.message || 'Failed to demote admin');
		}
	}

	async function confirmKickMember() {
		if (memberToKick) {
			await handleKickMember(memberToKick);
		}
	}

	function togglePanel() {
		panelOpen = !panelOpen;
	}
</script>

<!-- Toggle button -->
<button
	class="absolute top-1/2 -translate-y-1/2 right-0 z-10 flex items-center justify-center w-5 h-12 rounded-l-lg bg-sidebar border border-r-0 border-sidebar-border text-muted-foreground hover:text-foreground hover:bg-sidebar-accent transition-all duration-200 group"
	onclick={togglePanel}
	aria-label={panelOpen ? 'Close members panel' : 'Open members panel'}
>
	<ChevronRight
		class="size-3.5 transition-transform duration-200 {panelOpen ? 'rotate-0' : 'rotate-180'}"
	/>
</button>

<div
	class="shrink-0 overflow-hidden transition-all duration-250 ease-in-out border-l border-border bg-card
	{panelOpen ? 'w-[280px] opacity-100' : 'w-0 opacity-0 pointer-events-none'}"
>
	<div class="w-[280px] flex flex-col gap-3.5 h-full p-5 overflow-y-auto">
		<!-- Panel Header -->
		<div class="flex items-center justify-between">
			<div class="flex flex-col">
				<div class="flex items-center gap-2">
					<Users class="size-4 text-muted-foreground" />
					<span class="text-body-md font-semibold text-foreground">Members</span>
				</div>
				<span class="font-mono text-[11.5px] text-muted-foreground">
					{#if isLoadingMembers}
						Loading...
					{:else}
						{members.length} member{members.length !== 1 ? 's' : ''}
					{/if}
				</span>
			</div>

			<!-- Invite Member Button -->
			{#if currentUserRole === 'owner' || currentUserRole === 'admin'}
				<button
					type="button"
					class="flex items-center justify-center size-7 rounded-md text-muted-foreground hover:text-primary hover:bg-primary/10 transition-colors shrink-0"
					onclick={() => (isInviteModalOpen = true)}
					aria-label="Invite member"
					title="Invite member"
				>
					<UserPlus class="size-4" />
				</button>
			{/if}
		</div>

		{#if isLoadingMembers}
			<MemberSkeleton />
		{:else if membersError}
			<p class="text-[12px] text-destructive">{membersError}</p>
		{:else}
			<div class="flex flex-col gap-1">
				<!-- OWNER section -->
				{#if owners.length > 0}
					<p
						class="text-[10px] font-semibold uppercase tracking-widest text-muted-foreground/50 px-2 pt-1 pb-0.5"
					>
						Owner
					</p>

					{#each owners as member (member.ID)}
						<div
							class="flex items-center gap-2.5 px-2 py-2 rounded-lg hover:bg-white/4 transition-colors"
						>
							{#if member.User?.AvatarURL}
								<img
									src={member.User.AvatarURL}
									alt={member.User.Name}
									class="size-8 rounded-full object-cover shrink-0"
								/>
							{:else}
								<div
									class="size-8 rounded-full flex items-center justify-center font-semibold text-[12.5px] shrink-0 {avatarColor(
										member.User?.Name ?? 'U'
									)}"
								>
									{initials(member.User?.Name ?? 'U')}
								</div>
							{/if}
							<div class="flex-1 min-w-0">
								<p class="text-[13px] font-medium text-foreground truncate m-0">
									{member.User?.Name ?? 'Unknown'}
								</p>
							</div>
							<span
								class="text-[10px] font-medium px-2 py-0.5 rounded-full bg-primary/14 text-primary whitespace-nowrap shrink-0"
							>
								Owner
							</span>
						</div>
					{/each}
				{/if}

				<!-- Divider -->
				{#if owners.length > 0 && (admins.length > 0 || regularMembers.length > 0)}
					<div class="h-px bg-white/8 my-1"></div>
				{/if}

				<!-- ADMIN section -->
				{#if admins.length > 0}
					<p
						class="text-[10px] font-semibold uppercase tracking-widest text-muted-foreground/50 px-2 pt-1 pb-0.5"
					>
						Admins
					</p>

					{#each admins as member (member.ID)}
						<div
							class="flex items-center gap-2.5 px-2 py-2 rounded-lg hover:bg-white/4 transition-colors"
						>
							{#if member.User?.AvatarURL}
								<img
									src={member.User.AvatarURL}
									alt={member.User.Name}
									class="size-8 rounded-full object-cover shrink-0"
								/>
							{:else}
								<div
									class="size-8 rounded-full flex items-center justify-center font-semibold text-[12.5px] shrink-0 {avatarColor(
										member.User?.Name ?? 'U'
									)}"
								>
									{initials(member.User?.Name ?? 'U')}
								</div>
							{/if}
							<div class="flex-1 min-w-0">
								<p class="text-[13px] font-medium text-foreground truncate m-0">
									{member.User?.Name ?? 'Unknown'}
								</p>
							</div>
							<div class="flex items-center gap-1 shrink-0">
								<span
									class="text-[10px] font-medium px-2 py-0.5 rounded-full bg-blue-500/14 text-blue-500 whitespace-nowrap"
								>
									Admin
								</span>
								{#if currentUserRole === 'owner'}
									<div class="relative">
										<button
											class="flex items-center justify-center size-6 rounded-md text-muted-foreground hover:bg-white/8 hover:text-foreground transition-colors"
											aria-label="Admin options"
											onclick={(e) => {
												e.stopPropagation();
												toggleDropdown('member-' + member.ID);
											}}
										>
											<MoreVertical class="size-3.5" />
										</button>

										{#if openDropdownId === 'member-' + member.ID}
											<div
												class="absolute bottom-full right-0 mb-1.5 min-w-[170px] bg-popover border border-border rounded-lg p-1.5 z-20"
												role="presentation"
												onclick={(e) => e.stopPropagation()}
											>
												<button
													class="flex items-center gap-2 px-2.5 py-2 rounded-md text-[13px] text-foreground hover:bg-white/7 transition-colors text-left w-full whitespace-nowrap"
													onclick={(e) => {
														e.preventDefault();
														toggleDropdown('');
														handleDemoteMember(member.UserID);
													}}
												>
													<ShieldX class="size-[14px]" />
													Demote to Member
												</button>
												<button
													class="flex items-center gap-2 px-2.5 py-2 rounded-md text-[13px] text-destructive hover:bg-white/7 transition-colors text-left w-full whitespace-nowrap"
													onclick={(e) => {
														e.preventDefault();
														toggleDropdown('');
														setTimeout(() => {
															memberToKick = member.UserID;
															isConfirmKickOpen = true;
														}, 0);
													}}
												>
													<Trash2 class="size-[14px]" />
													Remove member
												</button>
											</div>
										{/if}
									</div>
								{/if}
							</div>
						</div>
					{/each}
				{/if}

				<!-- Divider -->
				{#if admins.length > 0 && regularMembers.length > 0}
					<div class="h-px bg-white/8 my-1"></div>
				{/if}

				<!-- MEMBERS section -->
				{#if regularMembers.length > 0}
					<p
						class="text-[10px] font-semibold uppercase tracking-widest text-muted-foreground/50 px-2 pt-1 pb-0.5"
					>
						Members
					</p>

					{#each regularMembers as member (member.ID)}
						<div
							class="flex items-center gap-2.5 px-2 py-2 rounded-lg hover:bg-white/4 transition-colors"
						>
							{#if member.User?.AvatarURL}
								<img
									src={member.User.AvatarURL}
									alt={member.User.Name}
									class="size-8 rounded-full object-cover shrink-0"
								/>
							{:else}
								<div
									class="size-8 rounded-full flex items-center justify-center font-semibold text-[12.5px] shrink-0 {avatarColor(
										member.User?.Name ?? 'U'
									)}"
								>
									{initials(member.User?.Name ?? 'U')}
								</div>
							{/if}
							<div class="flex-1 min-w-0">
								<p class="text-[13px] font-medium text-foreground truncate m-0">
									{member.User?.Name ?? 'Unknown'}
								</p>
							</div>
							<div class="flex items-center gap-1 shrink-0">
								<span
									class="text-[10px] font-medium px-2 py-0.5 rounded-full bg-white/7 text-muted-foreground whitespace-nowrap"
								>
									Member
								</span>
								{#if currentUserRole === 'owner' || currentUserRole === 'admin'}
									<div class="relative">
										<button
											class="flex items-center justify-center size-6 rounded-md text-muted-foreground hover:bg-white/8 hover:text-foreground transition-colors"
											aria-label="Member options"
											onclick={(e) => {
												e.stopPropagation();
												toggleDropdown('member-' + member.ID);
											}}
										>
											<MoreVertical class="size-3.5" />
										</button>

										{#if openDropdownId === 'member-' + member.ID}
											<div
												class="absolute bottom-full right-0 mb-1.5 min-w-[170px] bg-popover border border-border rounded-lg p-1.5 z-20"
												role="presentation"
												onclick={(e) => e.stopPropagation()}
											>
												<button
													class="flex items-center gap-2 px-2.5 py-2 rounded-md text-[13px] text-foreground hover:bg-white/7 transition-colors text-left w-full whitespace-nowrap"
													onclick={(e) => {
														e.preventDefault();
														toggleDropdown('');
														handlePromoteMember(member.UserID);
													}}
												>
													<ShieldAlert class="size-[14px]" />
													Promote to Admin
												</button>
												<button
													class="flex items-center gap-2 px-2.5 py-2 rounded-md text-[13px] text-destructive hover:bg-white/7 transition-colors text-left w-full whitespace-nowrap"
													onclick={(e) => {
														e.preventDefault();
														toggleDropdown('');
														setTimeout(() => {
															memberToKick = member.UserID;
															isConfirmKickOpen = true;
														}, 0);
													}}
												>
													<Trash2 class="size-[14px]" />
													Remove member
												</button>
											</div>
										{/if}
									</div>
								{/if}
							</div>
						</div>
					{/each}
				{/if}

				<!-- Empty members -->
				{#if !isLoadingMembers && members.length === 0}
					<p class="text-[12px] text-muted-foreground text-center py-4">No members found.</p>
				{/if}
			</div>
		{/if}
	</div>
</div>

<ConfirmDialog
	bind:open={isConfirmKickOpen}
	title="Remove member?"
	description="Are you sure you want to remove this member from the organization? They will no longer have access to any monitors."
	confirmText="Remove"
	cancelText="Cancel"
	destructive={true}
	onConfirm={confirmKickMember}
/>


<InviteMemberModal
	bind:open={isInviteModalOpen}
	{orgId}
	onInvited={() => onMembersUpdated?.()}
/>
