<script lang="ts">
	import { MoreVertical, ShieldCheck, Shield, Trash2, ShieldAlert, ShieldX } from 'lucide-svelte';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { getInitials as initials, getAvatarColor as avatarColor } from '$lib/utils/string';

	export interface MemberUser {
		ID: string;
		Name: string;
		AvatarURL?: string;
	}

	export interface Member {
		ID: string;
		OrganizationID: string;
		UserID: string;
		Role: string;
		User: MemberUser;
	}

	interface Props {
		member: Member;
		currentUserId: string;
		isOwner: boolean;
		canManage: boolean;
		onPromote: (id: string) => void;
		onDemote: (id: string) => void;
		onKick: (m: Member) => void;
	}

	let { member, currentUserId, isOwner, canManage, onPromote, onDemote, onKick }: Props = $props();
</script>

<div class="flex items-center justify-between p-4 hover:bg-muted/20 transition-colors">
	<!-- User Info -->
	<div class="flex items-center gap-3.5 min-w-0">
		{#if member.User?.AvatarURL}
			<img
				src={member.User.AvatarURL}
				alt={member.User?.Name || 'User'}
				class="size-9 rounded-lg object-cover border border-border shrink-0"
			/>
		{:else}
			<div
				class="size-9 rounded-lg flex items-center justify-center font-bold text-xs shrink-0 border border-border {avatarColor(
					member.User?.Name || 'User'
				)}"
			>
				{initials(member.User?.Name || 'User')}
			</div>
		{/if}

		<div class="min-w-0">
			<div class="flex items-center gap-2">
				<span class="text-sm font-semibold text-foreground truncate">
					{member.User?.Name || 'Unknown User'}
				</span>
				{#if member.UserID === currentUserId}
					<Badge
						variant="secondary"
						class="h-4.5 px-1.5 py-0 text-[10px] font-mono text-muted-foreground"
					>
						You
					</Badge>
				{/if}
			</div>
			<div class="flex items-center gap-1.5 mt-0.5">
				<span class="text-xs text-muted-foreground capitalize">{member.Role}</span>
			</div>
		</div>
	</div>

	<!-- Role Pill & Actions -->
	<div class="flex items-center gap-3 shrink-0">
		<!-- Role Indicator -->
		{#if member.Role === 'owner'}
			<Badge
				variant="outline"
				class="gap-1.5 px-2.5 py-1 text-xs font-medium text-amber-400 bg-amber-500/10 border-amber-500/20"
			>
				<ShieldCheck class="size-3.5" />
				<span>Owner</span>
			</Badge>
		{:else if member.Role === 'admin'}
			<Badge
				variant="outline"
				class="gap-1.5 px-2.5 py-1 text-xs font-medium text-emerald-400 bg-emerald-500/10 border-emerald-500/20"
			>
				<Shield class="size-3.5" />
				<span>Admin</span>
			</Badge>
		{:else}
			<Badge
				variant="secondary"
				class="gap-1.5 px-2.5 py-1 text-xs font-medium text-muted-foreground border-border"
			>
				<span>Member</span>
			</Badge>
		{/if}

		<!-- Member Action Menu -->
		{#if canManage && member.UserID !== currentUserId && member.Role !== 'owner'}
			<DropdownMenu.Root>
				<DropdownMenu.Trigger
					class="p-1.5 text-muted-foreground hover:text-foreground rounded-md hover:bg-secondary transition-colors focus:outline-none cursor-pointer"
				>
					<MoreVertical class="size-4" />
				</DropdownMenu.Trigger>
				<DropdownMenu.Content
					align="end"
					class="bg-card border-border text-foreground min-w-37.5 p-1 rounded-lg"
				>
					{#if member.Role === 'member' && isOwner}
						<DropdownMenu.Item
							class="hover:bg-white/10 cursor-pointer rounded-md p-2 text-xs"
							onclick={() => onPromote(member.UserID)}
						>
							<ShieldAlert class="mr-2 size-3.5 text-emerald-400" /> Make Admin
						</DropdownMenu.Item>
					{:else if member.Role === 'admin' && isOwner}
						<DropdownMenu.Item
							class="hover:bg-white/10 cursor-pointer rounded-md p-2 text-xs"
							onclick={() => onDemote(member.UserID)}
						>
							<ShieldX class="mr-2 size-3.5 text-amber-400" /> Demote to Member
						</DropdownMenu.Item>
					{/if}

					<DropdownMenu.Item
						class="hover:bg-destructive/10 text-destructive cursor-pointer rounded-md p-2 text-xs"
						onclick={() => onKick(member)}
					>
						<Trash2 class="mr-2 size-3.5" /> Remove from Org
					</DropdownMenu.Item>
				</DropdownMenu.Content>
			</DropdownMenu.Root>
		{/if}
	</div>
</div>
