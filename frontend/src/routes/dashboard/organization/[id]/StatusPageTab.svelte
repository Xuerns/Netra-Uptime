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
	import { userTimezone } from '$lib/stores/timezone';
	import { formatDateOnly } from '$lib/utils/datetime';

	import AddStatusPageModal from './AddStatusPageModal.svelte';
	import EditStatusPageModal from './EditStatusPageModal.svelte';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';

	let {
		orgId,
		monitors,
		currentUserRole,
		isAddModalOpen = $bindable(false)
	} = $props<{
		orgId: string;
		monitors: Monitor[];
		currentUserRole: string | undefined;
		isAddModalOpen?: boolean;
	}>();

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
		<div
			class="flex flex-col items-center justify-center gap-3 py-16 border border-border rounded-lg bg-white/3"
		>
			<TriangleAlert class="size-8 text-destructive" />
			<p class="text-sm text-muted-foreground">{fetchError}</p>
			<Button
				class="bg-secondary border border-border text-muted-foreground hover:text-foreground hover:bg-accent active:scale-[0.98] transition-all duration-200 h-11 px-6 font-medium shadow-sm"
				onclick={loadStatusPages}
			>
				Try again
			</Button>
		</div>
	{:else if statusPages.length === 0}
		<div
			class="flex flex-col items-center justify-center text-center py-24 px-6 gap-3.5 bg-white/3 border border-border rounded-lg backdrop-blur-md"
		>
			<div
				class="flex items-center justify-center size-[60px] rounded-lg bg-primary/8 text-primary"
			>
				<Globe class="size-6" />
			</div>
			<h2 class="text-[19px] font-semibold text-foreground m-0">No status pages yet</h2>
			<p class="text-sm leading-relaxed text-muted-foreground max-w-[360px] m-0">
				{#if currentUserRole === 'owner'}
					Create a public status page to let your customers know when services are down.
				{:else}
					There are no status pages in this organization yet.
				{/if}
			</p>
			{#if currentUserRole === 'owner'}
				<Button
					class="mt-2 bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-foreground active:scale-[0.98] transition-all duration-200 h-11 px-6 font-medium shadow-sm"
					onclick={() => (isAddModalOpen = true)}
				>
					<Plus class="size-4 mr-2" />
					Create Status Page
				</Button>
			{/if}
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			{#each statusPages as sp}
				<div class="bg-white/3 border border-border rounded-xl p-5 flex flex-col gap-4">
					<div class="flex justify-between items-start">
						<div>
							<h3 class="font-medium text-foreground">{sp.Title}</h3>
							<p class="text-sm font-mono text-muted-foreground mt-1">
								{sp.Subdomain}.pulsewatch.com
							</p>
						</div>
						<DropdownMenu.Root>
							<DropdownMenu.Trigger
								class="p-1.5 hover:bg-white/10 rounded-md text-muted-foreground hover:text-foreground transition-colors"
							>
								<MoreVertical class="size-4" />
							</DropdownMenu.Trigger>
							<DropdownMenu.Content
								align="end"
								class="bg-card border-border text-foreground min-w-[160px] p-1.5"
							>
								<DropdownMenu.Item
									class="hover:bg-white/10 cursor-pointer rounded-md p-2"
									onclick={() => window.open(`/status/${sp.Subdomain}`, '_blank')}
								>
									<ExternalLink class="mr-2 size-4 text-blue-400" /> View Public Page
								</DropdownMenu.Item>
								<DropdownMenu.Item
									class="hover:bg-white/10 cursor-pointer rounded-md p-2"
									onclick={() => {
										selectedStatusPage = sp;
										isEditModalOpen = true;
									}}
								>
									<Edit2 class="mr-2 size-4 text-primary" /> Edit Details
								</DropdownMenu.Item>
								<DropdownMenu.Separator class="bg-white/10" />
								<DropdownMenu.Item
									class="hover:bg-destructive/20 hover:text-destructive cursor-pointer rounded-md p-2 text-destructive"
									onclick={() => openDeleteConfirm(sp.ID)}
								>
									<Trash2 class="mr-2 size-4" /> Delete
								</DropdownMenu.Item>
							</DropdownMenu.Content>
						</DropdownMenu.Root>
					</div>

					<div class="flex items-center gap-2 mt-auto pt-4 border-t border-border">
						<span
							class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium {sp.IsPublic
								? 'bg-primary/20 text-primary'
								: 'bg-white/10 text-muted-foreground'}"
						>
							<div
								class="size-1.5 rounded-full {sp.IsPublic ? 'bg-primary' : 'bg-on-surface-variant'}"
							></div>
							{sp.IsPublic ? 'Public' : 'Private'}
						</span>
						<span class="text-xs text-muted-foreground ml-auto">
							Created {formatDateOnly(sp.CreatedAt, $userTimezone)}
						</span>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<AddStatusPageModal bind:isOpen={isAddModalOpen} {orgId} {monitors} onSuccess={loadStatusPages} />
<EditStatusPageModal
	bind:isOpen={isEditModalOpen}
	statusPage={selectedStatusPage}
	{orgId}
	{monitors}
	onSuccess={loadStatusPages}
/>

<ConfirmDialog
	bind:open={isConfirmDeleteOpen}
	title="Delete status page?"
	description="Are you sure you want to delete this status page? This action cannot be undone."
	confirmText="Delete"
	cancelText="Cancel"
	destructive={true}
	onConfirm={confirmDelete}
/>
