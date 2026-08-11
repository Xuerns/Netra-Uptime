<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { toast } from 'svelte-sonner';
	import { Loader2, AlertTriangle, ArrowLeft, ChevronDown, Check } from 'lucide-svelte';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';
	import { OrgAPI } from '$lib/api/organization';

	const orgId = $page.params.id as string;
	const currentUserId = $page.data.user?.id;

	let isLoading = $state(true);
	let isUpdating = $state(false);
	let isDisbanding = $state(false);
	let isDeleteDialogOpen = $state(false);
	let hasAccess = $state(false);
	let confirmOrgName = $state('');

	let currentUserRole = $state('');
	let eligibleMembers = $state<any[]>([]);
	
	let isTransferModalOpen = $state(false);
	let selectedNewOwnerId = $state('');
	let isDropdownOpen = $state(false);
	let selectedMember = $derived(eligibleMembers.find(m => m.UserID === selectedNewOwnerId));
	
	let isLeaveModalOpen = $state(false);

	function getAvatarColor(name: string) {
		const colors = [
			'bg-blue-500/10 text-blue-500',
			'bg-emerald-500/10 text-emerald-500',
			'bg-purple-500/10 text-purple-500',
			'bg-amber-500/10 text-amber-500',
			'bg-rose-500/10 text-rose-500',
			'bg-cyan-500/10 text-cyan-500'
		];
		let hash = 0;
		for (let i = 0; i < name.length; i++) {
			hash = name.charCodeAt(i) + ((hash << 5) - hash);
		}
		return colors[Math.abs(hash) % colors.length];
	}

	function getInitials(name: string) {
		return name
			.split(' ')
			.map((n) => n[0])
			.join('')
			.toUpperCase()
			.slice(0, 2);
	}

	let formData = $state({
		name: '',
		slug: ''
	});
	
	let originalOrgName = $state('');

	onMount(async () => {
		try {
			// Fetch members to check role
			const membersRes = await OrgAPI.getMembers(orgId);
			const members = membersRes.data || [];
			const currentUser = members.find((m: any) => m.UserID === currentUserId);
			
			if (currentUser) {
				hasAccess = true;
				currentUserRole = currentUser.Role;
				eligibleMembers = members.filter((m: any) => m.UserID !== currentUserId);
				
				if (currentUserRole === 'owner') {
					// Fetch org details only if owner
					const orgsRes = await OrgAPI.getAll();
					const org = orgsRes.data?.find((o: any) => o.ID === orgId);
					if (org) {
						formData.name = org.Name;
						formData.slug = org.Slug;
						originalOrgName = org.Name;
					}
				}
			}
		} catch (error) {
			console.error('Failed to load organization data:', error);
			toast.error('Failed to load organization settings');
		} finally {
			isLoading = false;
		}
	});

	async function updateOrganization() {
		if (!formData.name || !formData.slug) {
			toast.error('Name and slug cannot be empty');
			return;
		}
		
		const slugRegex = /^[a-z0-9-]+$/;
		if (!slugRegex.test(formData.slug)) {
			toast.error('Slug can only contain lowercase letters, numbers, and hyphens');
			return;
		}

		if (formData.slug.length < 6) {
			toast.error('Slug must be at least 6 characters long');
			return;
		}

		isUpdating = true;
		try {
			await OrgAPI.update(orgId, {
				organization_id: orgId,
				name: formData.name,
				slug: formData.slug
			});
			originalOrgName = formData.name;
			toast.success('Organization updated successfully');
		} catch (error: any) {
			toast.error(error.message || 'Failed to update organization');
		} finally {
			isUpdating = false;
		}
	}

	async function disbandOrganization() {
		if (confirmOrgName !== originalOrgName) {
			toast.error('Organization name does not match');
			return;
		}
		
		isDisbanding = true;
		try {
			await OrgAPI.disband(orgId);
			toast.success('Organization disbanded successfully');
			goto('/dashboard');
		} catch (error: any) {
			toast.error(error.message || 'Failed to disband organization');
			isDisbanding = false;
		}
	}

	async function transferOwnership() {
		if (!selectedNewOwnerId) {
			toast.error('Please select a member to transfer ownership');
			return;
		}
		try {
			await OrgAPI.transferOwnership(orgId, selectedNewOwnerId);
			toast.success('Ownership transferred successfully');
			goto('/dashboard');
		} catch (error: any) {
			toast.error(error.message || 'Failed to transfer ownership');
		} finally {
			isTransferModalOpen = false;
		}
	}

	async function leaveOrganization() {
		try {
			await OrgAPI.leaveOrganization(orgId);
			toast.success('You have left the organization');
			goto('/dashboard');
		} catch (error: any) {
			toast.error(error.message || 'Failed to leave organization');
		} finally {
			isLeaveModalOpen = false;
		}
	}
</script>

<svelte:head>
	<title>Organization Settings - PulseWatch</title>
</svelte:head>

<div class="p-6 max-w-4xl mx-auto space-y-6">
	<div class="flex items-center gap-4 mb-2">
		<a
			href={`/dashboard/organization/${orgId}`}
			class="flex items-center justify-center size-9 rounded-lg border bg-white/2 border-white/10 text-on-surface-variant hover:bg-white/7 hover:text-on-surface transition-colors"
			aria-label="Back to Organization"
		>
			<ArrowLeft class="size-4.5" />
		</a>
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-white">Organization Settings</h1>
			<p class="text-slate-400">Manage your organization details and dangerous actions.</p>
		</div>
	</div>

	{#if isLoading}
		<div class="flex items-center justify-center py-20">
			<Loader2 class="size-8 text-primary animate-spin" />
		</div>
	{:else if !hasAccess}
		<Card.Root class="bg-[#131314] border-red-900/50">
			<Card.Content class="p-8 text-center space-y-4">
				<div class="mx-auto size-12 bg-red-500/10 rounded-full flex items-center justify-center">
					<AlertTriangle class="size-6 text-red-500" />
				</div>
				<h2 class="text-xl font-semibold text-white">Access Denied</h2>
				<p class="text-slate-400 max-w-md mx-auto">
					You do not have permission to view this page. You must be an active member of this organization.
				</p>
				<div class="pt-4">
					<Button
						href={`/dashboard`}
						class="bg-[#1A1A1B] border border-[#2A2A2B] text-primary hover:bg-primary hover:border-primary hover:text-white transition-colors"
					>
						Back to Dashboard
					</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{:else}
		<!-- General Settings (Owner Only) -->
		{#if currentUserRole === 'owner'}
			<Card.Root class="bg-[#131314] border-[#2A2A2B]">
				<Card.Header>
					<Card.Title class="text-white">General Information</Card.Title>
					<Card.Description class="text-slate-400">
						Update your organization's display name and URL slug.
					</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-6">
					<form
						onsubmit={(e) => {
							e.preventDefault();
							updateOrganization();
						}}
						class="space-y-4"
					>
						<div class="space-y-2">
							<Label for="name" class="text-slate-300">Organization Name</Label>
							<Input
								id="name"
								bind:value={formData.name}
								placeholder="e.g. Acme Corp"
								class="bg-[#1A1A1B] border-[#2A2A2B] text-white placeholder:text-slate-500 focus-visible:ring-emerald-500 h-11 transition-colors"
							/>
						</div>
						<div class="space-y-2">
							<Label for="slug" class="text-slate-300">Organization Slug</Label>
							<div class="flex flex-col sm:flex-row gap-3">
								<div class="flex-1 flex items-center bg-[#151516] border border-[#2A2A2B] rounded-md h-11 px-3 text-slate-400 font-mono text-sm overflow-hidden">
									<span class="truncate">https://status.pulsewatch.com/</span>
								</div>
								<Input
									id="slug"
									bind:value={formData.slug}
									placeholder="acme-corp"
									class="bg-[#1A1A1B] border-[#2A2A2B] text-white placeholder:text-slate-500 focus-visible:ring-emerald-500 h-11 transition-colors flex-1"
								/>
							</div>
							<p class="text-xs text-slate-500">Must be at least 6 characters. Only lowercase letters, numbers, and hyphens.</p>
						</div>

						<div class="flex justify-end pt-4">
							<Button
								type="submit"
								disabled={isUpdating}
								class="bg-[#1A1A1B] border border-[#2A2A2B] text-emerald-500 hover:bg-emerald-600 hover:border-emerald-600 hover:text-white active:scale-[0.98] transition-all duration-200 h-11 px-6 font-medium shadow-sm"
							>
								{#if isUpdating}
									<Loader2 class="w-4 h-4 mr-2 animate-spin" />
									Saving...
								{:else}
									Save Changes
								{/if}
							</Button>
						</div>
					</form>
				</Card.Content>
			</Card.Root>
		{/if}

		<!-- Danger Zone -->
		<Card.Root class="bg-[#131314] border border-red-900/30">
			<Card.Header>
				<Card.Title class="text-red-500 flex items-center gap-2">
					<AlertTriangle class="size-5" />
					Danger Zone
				</Card.Title>
				<Card.Description class="text-slate-400">
					Irreversible and destructive actions for your organization.
				</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-6">
				{#if currentUserRole === 'owner'}
					<!-- Transfer Ownership -->
					<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 py-4 border-b border-[#2A2A2B]/50">
						<div class="space-y-1">
							<h4 class="font-medium text-white">Transfer Ownership</h4>
							<p class="text-sm text-slate-400">Transfer this organization to another member.</p>
						</div>
						<Button
							variant="outline"
							onclick={() => (isTransferModalOpen = true)}
							class="bg-[#1A1A1B] border-[#2A2A2B] text-slate-400 hover:bg-[#2A2A2B] h-11 hover:text-white transition-colors"
						>
							Transfer
						</Button>
					</div>

					<!-- Disband Organization -->
					<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 py-2">
						<div class="space-y-1">
							<h4 class="font-medium text-white">Disband Organization</h4>
							<p class="text-sm text-slate-400">Permanently delete this organization and all of its data.</p>
						</div>
						<Button
							variant="outline"
							onclick={() => (isDeleteDialogOpen = true)}
							class="bg-[#1A1A1B] border-[#2A2A2B] text-red-500 hover:bg-red-950/50 hover:border-red-900 hover:text-red-400 active:scale-[0.98] h-11 transition-all duration-200 font-medium shadow-sm"
						>
							Disband Organization
						</Button>
					</div>
				{:else}
					<!-- Leave Organization -->
					<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 py-2">
						<div class="space-y-1">
							<h4 class="font-medium text-white">Leave Organization</h4>
							<p class="text-sm text-slate-400">Remove yourself from this organization. You will lose access to all resources.</p>
						</div>
						<Button
							variant="outline"
							onclick={() => (isLeaveModalOpen = true)}
							class="bg-[#1A1A1B] border-[#2A2A2B] text-red-500 hover:bg-red-950/50 hover:border-red-900 hover:text-red-400 active:scale-[0.98] h-11 transition-all duration-200 font-medium shadow-sm"
						>
							Leave Organization
						</Button>
					</div>
				{/if}
			</Card.Content>
		</Card.Root>
	{/if}
</div>

<!-- Modals -->

<ConfirmDialog
	bind:open={isDeleteDialogOpen}
	title="Disband Organization"
	confirmText="Disband Organization"
	destructive={true}
	onConfirm={disbandOrganization}
>
	<div class="space-y-4">
		<p class="text-slate-300">
			Are you absolutely sure? This will permanently delete <strong>{originalOrgName}</strong> and all associated monitors, status pages, and members.
		</p>
		<div class="bg-red-500/10 border border-red-500/20 p-3 rounded-md text-red-400 text-sm">
			This action cannot be undone.
		</div>
		<div class="space-y-2 pt-2">
			<Label for="confirmName" class="text-slate-300">
				Please type <strong>{originalOrgName}</strong> to confirm.
			</Label>
			<Input
				id="confirmName"
				bind:value={confirmOrgName}
				class="bg-[#1A1A1B] border-[#2A2A2B] text-white focus-visible:ring-red-500 transition-colors"
			/>
		</div>
	</div>
</ConfirmDialog>

<ConfirmDialog
	bind:open={isTransferModalOpen}
	title="Transfer Ownership"
	confirmText="Transfer Ownership"
	destructive={true}
	onConfirm={transferOwnership}
>
	<div class="space-y-4">
		<p class="text-slate-300">
			You are about to transfer the ownership of this organization. You will immediately be demoted to an admin and will lose the ability to disband the organization or transfer it again.
		</p>
		<div class="bg-red-500/10 border border-red-500/20 p-3 rounded-md text-red-400 text-sm">
			This action is immediate and cannot be reversed by you.
		</div>
		<div class="space-y-2 pt-2 relative">
			<Label for="newOwner" class="text-slate-300">Select the new owner</Label>
			
			<button
				type="button"
				class="w-full flex items-center justify-between bg-[#1A1A1B] border border-[#2A2A2B] text-white rounded-md h-12 px-3 focus:ring-2 focus:ring-red-500 outline-none transition-colors"
				onclick={() => isDropdownOpen = !isDropdownOpen}
			>
				{#if selectedMember}
					<div class="flex items-center gap-2">
						{#if selectedMember.User?.AvatarURL}
							<img src={selectedMember.User.AvatarURL} alt="" class="size-6 rounded-full object-cover" />
						{:else}
							<div class="size-6 rounded-full flex items-center justify-center font-semibold text-[10px] {getAvatarColor(selectedMember.User?.Name || 'U')}">
								{getInitials(selectedMember.User?.Name || 'U')}
							</div>
						{/if}
						<span class="text-sm font-medium">{selectedMember.User?.Name || 'Unknown'}</span>
						<span class="text-[10px] text-slate-400 capitalize bg-white/5 px-1.5 py-0.5 rounded-full">{selectedMember.Role}</span>
					</div>
				{:else}
					<span class="text-slate-400 text-sm">Select a member...</span>
				{/if}
				<ChevronDown class="size-4 text-slate-400" />
			</button>
			
			{#if isDropdownOpen}
				<!-- svelte-ignore a11y_click_events_have_key_events -->
				<!-- svelte-ignore a11y_no_static_element_interactions -->
				<div class="fixed inset-0 z-40" onclick={() => isDropdownOpen = false}></div>
				
				<div class="absolute z-50 w-full mt-1 bg-[#1A1A1B] border border-[#2A2A2B] rounded-md shadow-lg max-h-60 overflow-auto">
					{#each eligibleMembers as member}
						<button
							type="button"
							class="w-full flex items-center gap-3 px-3 py-2 hover:bg-white/5 transition-colors text-left"
							onclick={() => {
								selectedNewOwnerId = member.UserID;
								isDropdownOpen = false;
							}}
						>
							{#if member.User?.AvatarURL}
								<img src={member.User.AvatarURL} alt="" class="size-8 rounded-full object-cover shrink-0" />
							{:else}
								<div class="size-8 shrink-0 rounded-full flex items-center justify-center font-semibold text-xs {getAvatarColor(member.User?.Name || 'U')}">
									{getInitials(member.User?.Name || 'U')}
								</div>
							{/if}
							<div class="flex-1 flex flex-col min-w-0">
								<span class="text-sm font-medium text-white truncate">{member.User?.Name || 'Unknown'}</span>
								<span class="text-xs text-slate-400 capitalize">{member.Role}</span>
							</div>
							{#if selectedNewOwnerId === member.UserID}
								<Check class="size-4 text-emerald-500 shrink-0" />
							{/if}
						</button>
					{/each}
					{#if eligibleMembers.length === 0}
						<div class="px-3 py-3 text-sm text-slate-400 text-center">No eligible members found.</div>
					{/if}
				</div>
			{/if}
		</div>
	</div>
</ConfirmDialog>

<ConfirmDialog
	bind:open={isLeaveModalOpen}
	title="Leave Organization"
	confirmText="Leave Organization"
	destructive={true}
	onConfirm={leaveOrganization}
>
	<div class="space-y-4">
		<p class="text-slate-300">
			Are you sure you want to leave this organization? You will no longer have access to any monitors or status pages, and this action cannot be undone unless you are invited back.
		</p>
	</div>
</ConfirmDialog>
