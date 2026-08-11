<script lang="ts">
	import { onMount } from 'svelte';
	import {
		Globe,
		Plus,
		TriangleAlert,
		Activity,
		MoreVertical,
		Edit2,
		Trash2,
		ExternalLink
	} from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import { StatusPageAPI, type StatusPage } from '$lib/api/statuspage';
	import type { Monitor } from '$lib/api/monitor';
	import { toast } from 'svelte-sonner';
	
	import AddStatusPageModal from './AddStatusPageModal.svelte';
	import EditStatusPageModal from './EditStatusPageModal.svelte';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';
	
	let { orgId, monitors, currentUserRole, isAddModalOpen = $bindable(false) } = $props<{ orgId: string, monitors: Monitor[], currentUserRole: string | undefined, isAddModalOpen?: boolean }>();

	let statusPages = $state<StatusPage[]>([]);
	let isLoading = $state(true);
	let fetchError = $state('');

	let isEditModalOpen = $state(false);
	let selectedStatusPage = $state<StatusPage | null>(null);

	async function loadStatusPages() {
		isLoading = true;
		fetchError = '';
		try {
			const res = await StatusPageAPI.getAll(orgId);
			statusPages = res.data ?? [];
		} catch (err: any) {
			fetchError = err.message || 'Failed to load status pages.';
		} finally {
			isLoading = false;
		}
	}

	onMount(() => {
		loadStatusPages();
	});

	let isConfirmDeleteOpen = $state(false);
	let pageToDelete = $state<string | null>(null);

	function openDeleteConfirm(id: string) {
		pageToDelete = id;
		isConfirmDeleteOpen = true;
	}

	async function confirmDelete() {
		if (!pageToDelete) return;
		try {
			await StatusPageAPI.delete(orgId, pageToDelete);
			toast.success('Status page deleted successfully');
			loadStatusPages();
		} catch (err: any) {
			toast.error(err.message || 'Failed to delete status page');
		} finally {
			pageToDelete = null;
		}
	}
</script>

<div class="flex flex-col gap-6 h-full pt-2">

	{#if isLoading}
		<div class="flex items-center justify-center h-32">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
		</div>
	{:else if fetchError}
		<div class="flex flex-col items-center justify-center gap-3 py-16 border border-white/8 rounded-lg bg-white/3">
			<TriangleAlert class="size-8 text-error" />
			<p class="text-sm text-on-surface-variant">{fetchError}</p>
			<Button class="bg-[#1A1A1B] border border-[#2A2A2B] text-slate-300 hover:text-white hover:bg-[#2A2A2B] active:scale-[0.98] transition-all duration-200 h-11 px-6 font-medium shadow-sm" onclick={loadStatusPages}>
				Try again
			</Button>
		</div>
	{:else if statusPages.length === 0}
		<div class="flex flex-col items-center justify-center text-center py-24 px-6 gap-3.5 bg-white/3 border border-white/8 rounded-lg backdrop-blur-md">
			<div class="flex items-center justify-center size-[60px] rounded-lg bg-primary/8 text-primary">
				<Globe class="size-6" />
			</div>
			<h2 class="text-[19px] font-semibold text-on-surface m-0">No status pages yet</h2>
			<p class="text-sm leading-relaxed text-on-surface-variant max-w-[360px] m-0">
				{#if currentUserRole === 'owner'}
					Create a public status page to let your customers know when services are down.
				{:else}
					There are no status pages in this organization yet.
				{/if}
			</p>
			{#if currentUserRole === 'owner'}
			<Button class="mt-2 bg-[#1A1A1B] border border-[#2A2A2B] text-primary hover:bg-primary hover:border-primary hover:text-white active:scale-[0.98] transition-all duration-200 h-11 px-6 font-medium shadow-sm" onclick={() => isAddModalOpen = true}>
				<Plus class="size-4 mr-2" />
				Create Status Page
			</Button>
			{/if}
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			{#each statusPages as sp}
				<div class="bg-white/3 border border-white/8 rounded-xl p-5 flex flex-col gap-4">
					<div class="flex justify-between items-start">
						<div>
							<h3 class="font-medium text-white">{sp.Title}</h3>
							<p class="text-sm font-mono text-on-surface-variant mt-1">{sp.Subdomain}.pulsewatch.com</p>
						</div>
						<DropdownMenu.Root>
							<DropdownMenu.Trigger class="p-1.5 hover:bg-white/10 rounded-md text-on-surface-variant hover:text-white transition-colors">
								<MoreVertical class="size-4" />
							</DropdownMenu.Trigger>
							<DropdownMenu.Content align="end" class="bg-[#161618] border-white/8 text-white min-w-[160px] p-1.5">
								<DropdownMenu.Item class="hover:bg-white/10 cursor-pointer rounded-md p-2" onclick={() => window.open(`/status/${sp.Subdomain}`, '_blank')}>
									<ExternalLink class="mr-2 size-4 text-blue-400" /> View Public Page
								</DropdownMenu.Item>
								<DropdownMenu.Item class="hover:bg-white/10 cursor-pointer rounded-md p-2" onclick={() => { selectedStatusPage = sp; isEditModalOpen = true; }}>
									<Edit2 class="mr-2 size-4 text-primary" /> Edit Details
								</DropdownMenu.Item>
								<DropdownMenu.Separator class="bg-white/10" />
								<DropdownMenu.Item class="hover:bg-error/20 hover:text-error cursor-pointer rounded-md p-2 text-error" onclick={() => openDeleteConfirm(sp.ID)}>
									<Trash2 class="mr-2 size-4" /> Delete
								</DropdownMenu.Item>
							</DropdownMenu.Content>
						</DropdownMenu.Root>
					</div>
					
					<div class="flex items-center gap-2 mt-auto pt-4 border-t border-white/8">
						<span class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium {sp.IsPublic ? 'bg-primary/20 text-primary' : 'bg-white/10 text-on-surface-variant'}">
							<div class="size-1.5 rounded-full {sp.IsPublic ? 'bg-primary' : 'bg-on-surface-variant'}"></div>
							{sp.IsPublic ? 'Public' : 'Private'}
						</span>
						<span class="text-xs text-on-surface-variant ml-auto">
							Created {new Date(sp.CreatedAt).toLocaleDateString()}
						</span>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<AddStatusPageModal bind:isOpen={isAddModalOpen} {orgId} {monitors} onSuccess={loadStatusPages} />
<EditStatusPageModal bind:isOpen={isEditModalOpen} statusPage={selectedStatusPage} {orgId} {monitors} onSuccess={loadStatusPages} />

<ConfirmDialog
	bind:open={isConfirmDeleteOpen}
	title="Delete status page?"
	description="Are you sure you want to delete this status page? This action cannot be undone."
	confirmText="Delete"
	cancelText="Cancel"
	destructive={true}
	onConfirm={confirmDelete}
/>
