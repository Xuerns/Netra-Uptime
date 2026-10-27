<script lang="ts">
	import { goto } from '$app/navigation';
	import { Label } from '$lib/components/ui/label/index.js';
	import * as Alert from '$lib/components/ui/alert/index.js';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import * as Command from '$lib/components/ui/command/index.js';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';
	import { AlertTriangle, ChevronDown, Check } from 'lucide-svelte';
	import { OrgAPI } from '$lib/api/organization';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { getInitials, getAvatarColor } from '$lib/utils/string';
	import { toast } from 'svelte-sonner';

	export interface MemberUser {
		ID?: string;
		Name?: string;
		AvatarURL?: string;
	}

	export interface Member {
		ID?: string;
		OrganizationID?: string;
		UserID: string;
		Role?: string;
		User?: MemberUser;
	}

	interface Props {
		open?: boolean;
		orgId: string;
		eligibleMembers: Member[];
		onSuccess?: () => void;
	}

	let { open = $bindable(false), orgId, eligibleMembers, onSuccess }: Props = $props();

	let selectedNewOwnerId = $state('');
	let isDropdownOpen = $state(false);
	let selectedMember = $derived(eligibleMembers.find((m) => m.UserID === selectedNewOwnerId));

	async function handleTransferOwnership() {
		if (!selectedNewOwnerId) {
			toast.error('Please select a member to transfer ownership');
			return;
		}
		try {
			await OrgAPI.transferOwnership(orgId, selectedNewOwnerId);
			toast.success('Ownership transferred successfully');
			open = false;
			selectedNewOwnerId = '';
			if (onSuccess) {
				onSuccess();
			} else {
				goto('/dashboard/org');
			}
		} catch (error: unknown) {
			toast.error(parseApiError(error, 'Failed to transfer ownership'));
		}
	}
</script>

<ConfirmDialog
	bind:open
	title="Transfer Ownership"
	confirmText="Transfer Ownership"
	destructive={true}
	onConfirm={handleTransferOwnership}
>
	<div class="space-y-4">
		<p class="text-muted-foreground">
			You are about to transfer the ownership of this organization. You will immediately be demoted
			to an admin and will lose the ability to disband the organization or transfer it again.
		</p>
		<Alert.Root variant="destructive" class="border-destructive/30 bg-destructive/10">
			<AlertTriangle class="size-4" />
			<Alert.Description class="text-xs font-medium">
				This action is immediate and cannot be reversed by you.
			</Alert.Description>
		</Alert.Root>
		<div class="space-y-2 pt-2">
			<Label for="newOwner" class="text-muted-foreground">Select the new owner</Label>

			<Popover.Root bind:open={isDropdownOpen}>
				<Popover.Trigger
					class="w-full flex items-center justify-between bg-secondary border border-border text-foreground rounded-md h-12 px-3 focus:ring-2 focus:ring-red-500 outline-none transition-colors cursor-pointer"
				>
					{#if selectedMember}
						<div class="flex items-center gap-2 min-w-0">
							{#if selectedMember.User?.AvatarURL}
								<img
									src={selectedMember.User.AvatarURL}
									alt=""
									class="size-6 rounded-full object-cover shrink-0"
								/>
							{:else}
								<div
									class="size-6 rounded-full flex items-center justify-center font-semibold text-[10px] shrink-0 {getAvatarColor(
										selectedMember.User?.Name || 'U'
									)}"
								>
									{getInitials(selectedMember.User?.Name || 'U')}
								</div>
							{/if}
							<span class="text-sm font-medium truncate"
								>{selectedMember.User?.Name || 'Unknown'}</span
							>
							<span
								class="text-[10px] text-muted-foreground capitalize bg-white/5 px-1.5 py-0.5 rounded-full shrink-0"
								>{selectedMember.Role}</span
							>
						</div>
					{:else}
						<span class="text-muted-foreground text-sm">Select a member...</span>
					{/if}
					<ChevronDown class="size-4 text-muted-foreground shrink-0 opacity-60" />
				</Popover.Trigger>

				<Popover.Content
					class="w-[var(--bits-popover-anchor-width)] min-w-[320px] p-0"
					align="start"
				>
					<Command.Root>
						<Command.Input placeholder="Search member by name..." />
						<Command.List class="max-h-60 overflow-y-auto">
							<Command.Empty>No eligible members found.</Command.Empty>
							<Command.Group>
								{#each eligibleMembers as member (member.UserID)}
									<Command.Item
										value="{member.User?.Name || ''} {member.Role || ''}"
										onSelect={() => {
											selectedNewOwnerId = member.UserID;
											isDropdownOpen = false;
										}}
										class="flex items-center gap-3 px-3 py-2 cursor-pointer"
									>
										{#if member.User?.AvatarURL}
											<img
												src={member.User.AvatarURL}
												alt=""
												class="size-8 rounded-full object-cover shrink-0"
											/>
										{:else}
											<div
												class="size-8 shrink-0 rounded-full flex items-center justify-center font-semibold text-xs {getAvatarColor(
													member.User?.Name || 'U'
												)}"
											>
												{getInitials(member.User?.Name || 'U')}
											</div>
										{/if}
										<div class="flex-1 flex flex-col min-w-0">
											<span class="text-sm font-medium text-foreground truncate"
												>{member.User?.Name || 'Unknown'}</span
											>
											<span class="text-xs text-muted-foreground capitalize">{member.Role}</span>
										</div>
										{#if selectedNewOwnerId === member.UserID}
											<Check class="size-4 text-primary shrink-0 ml-auto" />
										{/if}
									</Command.Item>
								{/each}
							</Command.Group>
						</Command.List>
					</Command.Root>
				</Popover.Content>
			</Popover.Root>
		</div>
	</div>
</ConfirmDialog>
