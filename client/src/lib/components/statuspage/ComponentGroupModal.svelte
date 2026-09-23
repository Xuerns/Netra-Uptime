<script lang="ts">
	import { untrack } from 'svelte';
	import { StatusPageAPI, type ComponentGroup, type StatusPage } from '$lib/api/statuspage';
	import type { Monitor } from '$lib/api/monitor';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Button } from '$lib/components/ui/button/index.js';
	import { toast } from 'svelte-sonner';
	import {
		FolderTree,
		Plus,
		Trash2,
		Pencil,
		Layers,
		AlertCircle,
		Check,
		Radio
	} from 'lucide-svelte';
	import { parseApiError } from '$lib/utils/parseApiError';

	let {
		isOpen = $bindable(false),
		orgId,
		statusPage,
		monitors = [],
		onUpdated
	} = $props<{
		isOpen: boolean;
		orgId: string;
		statusPage: StatusPage | null;
		monitors: Monitor[];
		onUpdated?: () => void;
	}>();

	let groups = $state<ComponentGroup[]>([]);
	let isLoading = $state(false);
	let isSubmitting = $state(false);
	let activeTab = $state<'groups' | 'assignment'>('groups');

	// Create/Edit Group form
	let isEditingGroup = $state<ComponentGroup | null>(null);
	let groupName = $state('');
	let groupOrder = $state(0);
	let formError = $state('');

	// Monitor to Group mapping (monitorId -> groupId)
	let monitorGroupMap = $state<Record<string, string>>({});
	let savingMonitorId = $state<string | null>(null);

	let lastOpenedId = $state<string | null>(null);

	$effect(() => {
		if (isOpen && statusPage && orgId) {
			if (lastOpenedId !== statusPage.ID) {
				lastOpenedId = statusPage.ID;
				untrack(() => {
					loadGroups();
					resetForm();
					initMonitorMap();
				});
			}
		} else if (!isOpen) {
			lastOpenedId = null;
		}
	});

	function initMonitorMap() {
		const map: Record<string, string> = {};
		if (statusPage?.StatusPageMonitors) {
			for (const spm of statusPage.StatusPageMonitors) {
				map[spm.MonitorID] = spm.GroupID || '';
			}
		}
		monitorGroupMap = map;
	}

	function resetForm() {
		isEditingGroup = null;
		groupName = '';
		groupOrder = groups.length;
		formError = '';
	}

	async function loadGroups() {
		if (!statusPage || !orgId) return;
		isLoading = true;
		try {
			const res = await StatusPageAPI.getGroups(orgId, statusPage.ID);
			groups = res.data ?? [];
			if (!isEditingGroup && groupName === '') {
				groupOrder = groups.length;
			}
		} catch (err: any) {
			toast.error(parseApiError(err, 'Failed to fetch component groups.'));
		} finally {
			isLoading = false;
		}
	}

	function startEditGroup(g: ComponentGroup) {
		isEditingGroup = g;
		groupName = g.name;
		groupOrder = g.display_order;
		formError = '';
	}

	async function handleSaveGroup() {
		if (!groupName.trim()) {
			formError = 'Please enter a group name.';
			return;
		}

		if (!statusPage) return;
		isSubmitting = true;
		formError = '';

		try {
			if (isEditingGroup) {
				await StatusPageAPI.updateGroup(orgId, statusPage.ID, isEditingGroup.id, {
					name: groupName.trim(),
					display_order: Number(groupOrder)
				});
				toast.success('Component group updated.');
			} else {
				await StatusPageAPI.createGroup(orgId, statusPage.ID, {
					name: groupName.trim(),
					display_order: Number(groupOrder)
				});
				toast.success('Component group created.');
			}

			resetForm();
			await loadGroups();
			onUpdated?.();
		} catch (err: any) {
			formError = parseApiError(err, 'Failed to save component group.');
		} finally {
			isSubmitting = false;
		}
	}

	async function handleDeleteGroup(g: ComponentGroup) {
		if (!statusPage) return;
		try {
			await StatusPageAPI.deleteGroup(orgId, statusPage.ID, g.id);
			toast.success(`Group "${g.name}" deleted.`);
			await loadGroups();
			onUpdated?.();
		} catch (err: any) {
			toast.error(parseApiError(err, 'Failed to delete component group.'));
		}
	}

	async function handleAssignMonitor(monitorId: string, groupId: string) {
		if (!statusPage) return;
		savingMonitorId = monitorId;
		try {
			await StatusPageAPI.assignMonitorGroup(
				orgId,
				statusPage.ID,
				monitorId,
				groupId ? groupId : null
			);
			monitorGroupMap[monitorId] = groupId;
			toast.success('Monitor group updated.');
			onUpdated?.();
		} catch (err: any) {
			toast.error(parseApiError(err, 'Failed to assign monitor to group.'));
		} finally {
			savingMonitorId = null;
		}
	}

	// Filter monitors assigned to this status page
	let pageMonitors = $derived.by(() => {
		if (!statusPage?.StatusPageMonitors) return [];
		const monitorIdSet = new Set(statusPage.StatusPageMonitors.map((m: any) => m.MonitorID));
		return monitors.filter((m: Monitor) => monitorIdSet.has(m.ID));
	});
</script>

<Dialog.Root bind:open={isOpen}>
	<Dialog.Content
		class="sm:max-w-[640px] max-h-[85vh] flex flex-col bg-card border border-border rounded-2xl p-0 overflow-hidden shadow-2xl"
	>
		<!-- Header -->
		<div class="p-6 border-b border-border/60 flex items-center justify-between">
			<div class="flex items-center gap-3">
				<div>
					<Dialog.Title class="text-lg font-bold text-foreground">Component Groups</Dialog.Title>
					<Dialog.Description class="text-xs text-muted-foreground mt-0.5">
						Organize services into collapsible categories on <span
							class="text-foreground font-medium">{statusPage?.Title}</span
						>
					</Dialog.Description>
				</div>
			</div>
		</div>

		<!-- Tab Navigation -->
		<div class="px-6 pt-3 pb-0 border-b border-border/40 flex gap-4">
			<button
				type="button"
				class="text-xs font-semibold pb-2.5 border-b-2 transition-colors flex items-center gap-1.5 {activeTab ===
				'groups'
					? 'border-primary text-primary'
					: 'border-transparent text-muted-foreground hover:text-foreground'}"
				onclick={() => (activeTab = 'groups')}
			>
				<Layers class="size-3.5" />
				Manage Groups ({groups.length})
			</button>
			<button
				type="button"
				class="text-xs font-semibold pb-2.5 border-b-2 transition-colors flex items-center gap-1.5 {activeTab ===
				'assignment'
					? 'border-primary text-primary'
					: 'border-transparent text-muted-foreground hover:text-foreground'}"
				onclick={() => (activeTab = 'assignment')}
			>
				<Radio class="size-3.5" />
				Assign Monitors ({pageMonitors.length})
			</button>
		</div>

		<!-- Body -->
		<div class="flex-1 overflow-y-auto p-6 space-y-5">
			{#if activeTab === 'groups'}
				<!-- Form for Add/Edit Group -->
				<div class="rounded-xl border border-border/80 bg-white/3 p-4 space-y-3">
					<h3 class="text-xs font-semibold text-foreground uppercase tracking-wider">
						{isEditingGroup ? 'Edit Group' : 'Add New Group'}
					</h3>

					{#if formError}
						<div
							class="p-2.5 rounded-lg bg-destructive/10 border border-destructive/20 text-destructive text-xs flex items-center gap-2"
						>
							<AlertCircle class="size-4 shrink-0" />
							<span>{formError}</span>
						</div>
					{/if}

					<div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
						<div class="sm:col-span-2 space-y-1">
							<Label for="g-name" class="text-xs font-medium text-foreground">Group Name *</Label>
							<Input
								id="g-name"
								placeholder="e.g. Core Services, Databases, APIs"
								bind:value={groupName}
								class="text-xs bg-background/50 border-border"
							/>
						</div>
						<div class="space-y-1">
							<Label for="g-order" class="text-xs font-medium text-foreground">Display Order</Label>
							<Input
								id="g-order"
								type="number"
								placeholder="0"
								bind:value={groupOrder}
								class="text-xs bg-background/50 border-border"
							/>
						</div>
					</div>

					<div class="flex justify-end gap-2 pt-1">
						{#if isEditingGroup}
							<Button variant="ghost" size="sm" class="text-xs h-8" onclick={resetForm}>
								Cancel
							</Button>
						{/if}
						<Button
							size="sm"
							class="text-xs h-8 bg-primary text-primary-foreground gap-1"
							disabled={isSubmitting}
							onclick={handleSaveGroup}
						>
							{#if isEditingGroup}
								<Check class="size-3.5" />
								{isSubmitting ? 'Saving...' : 'Update Group'}
							{:else}
								<Plus class="size-3.5" />
								{isSubmitting ? 'Adding...' : 'Add Group'}
							{/if}
						</Button>
					</div>
				</div>

				<!-- List of Groups -->
				{#if isLoading}
					<div class="py-8 text-center text-xs text-muted-foreground">Loading groups...</div>
				{:else if groups.length === 0}
					<div
						class="py-10 text-center space-y-2 border border-dashed border-border/80 rounded-xl p-6"
					>
						<Layers class="size-8 text-muted-foreground/40 mx-auto" />
						<p class="text-xs font-medium text-foreground">No component groups configured</p>
						<p class="text-[11px] text-muted-foreground max-w-[50%] mx-auto">
							Create groups like "Web Applications", "APIs", or "Databases" to organize monitors
							cleanly on your public status page.
						</p>
					</div>
				{:else}
					<div class="space-y-2">
						<Label class="text-xs text-muted-foreground font-medium">Existing Groups</Label>
						<div class="space-y-2">
							{#each groups as g}
								<div
									class="flex items-center justify-between p-3 rounded-lg border border-border/70 bg-white/3 transition-colors hover:border-border"
								>
									<div class="flex items-center gap-3">
										<span
											class="text-xs font-mono size-6 rounded bg-white/5 border border-border/60 flex items-center justify-center text-muted-foreground"
										>
											{g.display_order}
										</span>
										<div>
											<p class="text-sm font-semibold text-foreground">{g.name}</p>
											<p class="text-[11px] text-muted-foreground">
												{g.monitor_ids?.length || 0} monitor{g.monitor_ids?.length === 1 ? '' : 's'} assigned
											</p>
										</div>
									</div>

									<div class="flex items-center gap-1">
										<Button
											variant="ghost"
											size="sm"
											class="size-7 p-0 text-muted-foreground hover:text-foreground"
											title="Edit Group"
											onclick={() => startEditGroup(g)}
										>
											<Pencil class="size-3.5" />
										</Button>
										<Button
											variant="ghost"
											size="sm"
											class="size-7 p-0 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
											title="Delete Group"
											onclick={() => handleDeleteGroup(g)}
										>
											<Trash2 class="size-3.5" />
										</Button>
									</div>
								</div>
							{/each}
						</div>
					</div>
				{/if}
			{:else}
				<!-- Monitor Assignment Tab -->
				<div class="space-y-3">
					<p class="text-xs text-muted-foreground">
						Assign each monitored service to a component group. Services without a group will appear
						under "Ungrouped Components".
					</p>

					{#if pageMonitors.length === 0}
						<div
							class="py-8 text-center text-xs text-muted-foreground border border-dashed border-border rounded-xl"
						>
							No monitors attached to this status page yet.
						</div>
					{:else}
						<div class="space-y-2">
							{#each pageMonitors as mon}
								<div
									class="flex items-center justify-between p-3 rounded-lg border border-border/70 bg-white/3 gap-3"
								>
									<div class="min-w-0">
										<p class="text-xs font-semibold text-foreground truncate">{mon.Name}</p>
										<span class="text-[10px] font-mono text-muted-foreground uppercase"
											>{mon.Type}</span
										>
									</div>

									<div class="w-48 sm:w-56 shrink-0">
										<select
											value={monitorGroupMap[mon.ID] || ''}
											disabled={savingMonitorId === mon.ID}
											class="w-full text-xs rounded-md bg-background/80 border border-border p-1.5 text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
											onchange={(e) => handleAssignMonitor(mon.ID, e.currentTarget.value)}
										>
											<option value="">Ungrouped (Default)</option>
											{#each groups as g}
												<option value={g.id}>{g.name}</option>
											{/each}
										</select>
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			{/if}
		</div>
	</Dialog.Content>
</Dialog.Root>
