<script lang="ts">
	import { untrack } from 'svelte';
	import { StatusPageAPI, type Maintenance, type StatusPage } from '$lib/api/statuspage';
	import type { Monitor } from '$lib/api/monitor';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Label } from '$lib/components/ui/label';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Alert from '$lib/components/ui/alert';
	import { toast } from 'svelte-sonner';
	import {
		Wrench,
		Plus,
		Trash2,
		CheckCircle2,
		Play,
		Calendar,
		AlertCircle,
		Pencil
	} from 'lucide-svelte';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { formatDateTimeShort } from '$lib/utils/datetime';

	let {
		isOpen = $bindable(false),
		orgId,
		statusPage,
		monitors = []
	}: {
		isOpen: boolean;
		orgId: string;
		statusPage: StatusPage | null;
		monitors: Monitor[];
	} = $props();

	let maintenances = $state<Maintenance[]>([]);
	let isLoading = $state(false);
	let isSubmitting = $state(false);
	let isFormOpen = $state(false);
	let editingMaintenance = $state<Maintenance | null>(null);

	// Form fields
	let title = $state('');
	let description = $state('');
	let editStatus = $state<'scheduled' | 'in_progress' | 'completed'>('scheduled');
	let scheduledStart = $state('');
	let scheduledEnd = $state('');
	let selectedMonitors = $state<string[]>([]);
	let formError = $state('');

	let lastOpenedId = $state<string | null>(null);

	$effect(() => {
		if (isOpen && statusPage && orgId) {
			if (lastOpenedId !== statusPage.ID) {
				lastOpenedId = statusPage.ID;
				untrack(() => {
					loadMaintenances();
					isFormOpen = false;
					editingMaintenance = null;
					resetForm();
				});
			}
		} else if (!isOpen) {
			lastOpenedId = null;
		}
	});

	function toLocalDatetimeString(dtStr: string): string {
		if (!dtStr) return '';
		const d = new Date(dtStr);
		if (isNaN(d.getTime())) return '';
		const pad = (n: number) => n.toString().padStart(2, '0');
		return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
	}

	function resetForm() {
		title = '';
		description = '';
		editStatus = 'scheduled';
		const now = new Date();
		const oneHourLater = new Date(now.getTime() + 60 * 60 * 1000);
		scheduledStart = toLocalDatetimeString(now.toISOString());
		scheduledEnd = toLocalDatetimeString(oneHourLater.toISOString());
		selectedMonitors = [];
		formError = '';
	}

	function openCreateForm() {
		editingMaintenance = null;
		resetForm();
		isFormOpen = true;
	}

	function openEditForm(m: Maintenance) {
		editingMaintenance = m;
		title = m.title;
		description = m.description || '';
		editStatus = (m.status as 'scheduled' | 'in_progress' | 'completed') || 'scheduled';
		scheduledStart = toLocalDatetimeString(m.scheduled_start);
		scheduledEnd = toLocalDatetimeString(m.scheduled_end);

		if (m.affected_monitor_ids && m.affected_monitor_ids.length > 0) {
			selectedMonitors = [...m.affected_monitor_ids];
		} else if (m.affected_monitors && m.affected_monitors.length > 0) {
			selectedMonitors = monitors
				.filter((mon: Monitor) => m.affected_monitors.includes(mon.Name))
				.map((mon: Monitor) => mon.ID);
		} else {
			selectedMonitors = [];
		}

		formError = '';
		isFormOpen = true;
	}

	async function loadMaintenances() {
		if (!statusPage || !orgId) return;
		isLoading = true;
		try {
			const res = await StatusPageAPI.getMaintenances(orgId, statusPage.ID);
			maintenances = res.data ?? [];
		} catch (err: any) {
			toast.error(parseApiError(err, 'Failed to fetch maintenances.'));
		} finally {
			isLoading = false;
		}
	}

	function toggleMonitor(monitorId: string) {
		if (selectedMonitors.includes(monitorId)) {
			selectedMonitors = selectedMonitors.filter((id) => id !== monitorId);
		} else {
			selectedMonitors = [...selectedMonitors, monitorId];
		}
	}

	async function handleSaveMaintenance() {
		if (!title.trim() || !scheduledStart || !scheduledEnd) {
			formError = 'Please fill in title and scheduled start/end dates.';
			return;
		}

		if (new Date(scheduledEnd) <= new Date(scheduledStart)) {
			formError = 'Scheduled end time must be after scheduled start time.';
			return;
		}

		if (!statusPage) return;
		isSubmitting = true;
		formError = '';

		try {
			if (editingMaintenance) {
				await StatusPageAPI.updateMaintenance(orgId, statusPage.ID, editingMaintenance.id, {
					title: title.trim(),
					description: description.trim() || undefined,
					status: editStatus,
					scheduled_start: new Date(scheduledStart).toISOString(),
					scheduled_end: new Date(scheduledEnd).toISOString(),
					monitor_ids: selectedMonitors
				});
				toast.success('Scheduled maintenance updated successfully!');
			} else {
				await StatusPageAPI.createMaintenance(orgId, statusPage.ID, {
					title: title.trim(),
					description: description.trim() || undefined,
					scheduled_start: new Date(scheduledStart).toISOString(),
					scheduled_end: new Date(scheduledEnd).toISOString(),
					monitor_ids: selectedMonitors.length ? selectedMonitors : undefined
				});
				toast.success('Scheduled maintenance created successfully!');
			}

			isFormOpen = false;
			editingMaintenance = null;
			resetForm();
			await loadMaintenances();
		} catch (err: any) {
			formError = parseApiError(err, 'Failed to save maintenance.');
		} finally {
			isSubmitting = false;
		}
	}

	async function handleStatusChange(m: Maintenance, newStatus: 'in_progress' | 'completed') {
		if (!statusPage) return;
		try {
			await StatusPageAPI.updateMaintenance(orgId, statusPage.ID, m.id, {
				status: newStatus
			});
			toast.success(`Maintenance marked as ${newStatus.replace('_', ' ')}.`);
			await loadMaintenances();
		} catch (err: any) {
			toast.error(parseApiError(err, 'Failed to update maintenance.'));
		}
	}

	async function handleDelete(m: Maintenance) {
		if (!statusPage) return;
		try {
			await StatusPageAPI.deleteMaintenance(orgId, statusPage.ID, m.id);
			toast.success('Maintenance deleted.');
			await loadMaintenances();
		} catch (err: any) {
			toast.error(parseApiError(err, 'Failed to delete maintenance.'));
		}
	}

	function formatDT(dtStr: string): string {
		if (!dtStr) return '';
		return formatDateTimeShort(dtStr);
	}
</script>

<Dialog.Root bind:open={isOpen}>
	<Dialog.Content
		class="sm:max-w-[620px] max-h-[85vh] flex flex-col bg-card border border-border rounded-2xl p-0 overflow-hidden"
	>
		<!-- Header -->
		<div class="p-6 border-b border-border/60 flex items-center justify-between">
			<div class="flex items-center gap-3">
				<div>
					<Dialog.Title class="text-lg font-bold text-foreground">
						Scheduled Maintenance
					</Dialog.Title>
					<Dialog.Description class="text-xs text-muted-foreground mt-0.5">
						Manage planned maintenance windows for <span class="text-foreground font-medium"
							>{statusPage?.Title}</span
						>
					</Dialog.Description>
				</div>
			</div>

			{#if !isFormOpen}
				<Button
					size="sm"
					class="gap-1.5 text-xs bg-primary text-primary-foreground"
					onclick={openCreateForm}
				>
					<Plus class="size-3.5" />
					Schedule
				</Button>
			{/if}
		</div>

		<!-- Body (Scrollable) -->
		<div class="flex-1 overflow-y-auto p-6 space-y-5">
			{#if isFormOpen}
				<!-- Create / Edit Form -->
				<div class="rounded-xl border border-border/80 bg-white/3 p-5 space-y-4">
					<div class="flex items-center justify-between">
						<h3 class="text-sm font-semibold text-foreground flex items-center gap-2">
							<Calendar class="size-4 text-blue-400" />
							{editingMaintenance ? 'Edit Scheduled Maintenance' : 'New Scheduled Maintenance'}
						</h3>
						<Button
							variant="ghost"
							size="sm"
							class="text-xs text-muted-foreground hover:text-foreground h-7"
							onclick={() => {
								isFormOpen = false;
								editingMaintenance = null;
							}}
						>
							Cancel
						</Button>
					</div>

					{#if formError}
						<Alert.Root variant="destructive" class="border-destructive/30 bg-destructive/10">
							<AlertCircle class="size-4 shrink-0" />
							<Alert.Description class="text-xs font-medium">
								{formError}
							</Alert.Description>
						</Alert.Root>
					{/if}

					<div class="space-y-1.5">
						<Label for="m-title" class="text-xs font-medium text-foreground">Title *</Label>
						<Input
							id="m-title"
							placeholder="e.g. Database Server Upgrade"
							bind:value={title}
							class="text-xs bg-background/50 border-border"
						/>
					</div>

					<div class="space-y-1.5">
						<Label for="m-desc" class="text-xs font-medium text-foreground">Description</Label>
						<Textarea
							id="m-desc"
							rows={3}
							placeholder="Brief explanation of what is being maintained and expected impact..."
							bind:value={description}
							class="text-xs bg-background/50 border-border text-foreground"
						/>
					</div>

					{#if editingMaintenance}
						<div class="space-y-1.5">
							<Label for="m-status" class="text-xs font-medium text-foreground">Status</Label>
							<select
								id="m-status"
								bind:value={editStatus}
								class="w-full text-xs rounded-md bg-background/50 border border-border p-2 text-foreground focus:outline-none focus:ring-1 focus:ring-ring"
							>
								<option value="scheduled">Scheduled</option>
								<option value="in_progress">In Progress</option>
								<option value="completed">Completed</option>
							</select>
						</div>
					{/if}

					<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
						<div class="space-y-1.5">
							<Label for="m-start" class="text-xs font-medium text-foreground"
								>Scheduled Start *</Label
							>
							<Input
								id="m-start"
								type="datetime-local"
								bind:value={scheduledStart}
								class="text-xs bg-background/50 border-border"
							/>
						</div>
						<div class="space-y-1.5">
							<Label for="m-end" class="text-xs font-medium text-foreground">Scheduled End *</Label>
							<Input
								id="m-end"
								type="datetime-local"
								bind:value={scheduledEnd}
								class="text-xs bg-background/50 border-border"
							/>
						</div>
					</div>

					<!-- Affected Monitors -->
					{#if monitors.length > 0}
						<div class="space-y-2">
							<Label class="text-xs font-medium text-foreground">Affected Monitors (optional)</Label
							>
							<div class="flex flex-wrap gap-2 max-h-32 overflow-y-auto p-1">
								{#each monitors as mon (mon.ID)}
									{@const selected = selectedMonitors.includes(mon.ID)}
									<button
										type="button"
										class="text-xs font-mono px-2.5 py-1 rounded-md border transition-colors {selected
											? 'bg-blue-500/20 text-blue-300 border-blue-500/40'
											: 'bg-white/5 text-muted-foreground border-border/70 hover:text-foreground'}"
										onclick={() => toggleMonitor(mon.ID)}
									>
										{mon.Name}
									</button>
								{/each}
							</div>
						</div>
					{/if}

					<div class="pt-2 flex justify-end gap-2">
						<Button
							variant="ghost"
							size="sm"
							class="text-xs"
							disabled={isSubmitting}
							onclick={() => {
								isFormOpen = false;
								editingMaintenance = null;
							}}
						>
							Cancel
						</Button>
						<Button
							size="sm"
							class="text-xs bg-blue-600 hover:bg-blue-500 text-white"
							disabled={isSubmitting}
							onclick={handleSaveMaintenance}
						>
							{isSubmitting
								? 'Saving...'
								: editingMaintenance
									? 'Update Maintenance'
									: 'Save Maintenance'}
						</Button>
					</div>
				</div>
			{/if}

			<!-- List of Maintenances -->
			{#if isLoading}
				<div class="py-8 text-center text-xs text-muted-foreground">Loading maintenances...</div>
			{:else if maintenances.length === 0 && !isFormOpen}
				<div
					class="py-10 text-center space-y-2 border border-dashed border-border/80 rounded-xl p-6"
				>
					<Wrench class="size-8 text-muted-foreground/40 mx-auto" />
					<p class="text-xs font-medium text-foreground">No scheduled maintenances</p>
					<p class="text-[11px] text-muted-foreground">
						Click the "Schedule" button above to inform users about upcoming maintenance windows.
					</p>
				</div>
			{:else}
				<div class="space-y-3">
					{#each maintenances as m (m.id)}
						<div
							class="rounded-xl border border-border/70 bg-white/3 p-4 space-y-2.5 transition-colors"
						>
							<div class="flex items-start justify-between gap-3">
								<div class="space-y-1 min-w-0">
									<div class="flex items-center gap-2">
										{#if m.status === 'in_progress'}
											<span
												class="inline-flex items-center gap-1 text-[10px] font-mono uppercase px-2 py-0.5 rounded-full bg-blue-500/20 text-blue-300 border border-blue-500/30"
											>
												<span class="size-1.5 rounded-full bg-blue-400"></span>
												In Progress
											</span>
										{:else if m.status === 'completed'}
											<span
												class="inline-flex items-center gap-1 text-[10px] font-mono uppercase px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
											>
												<CheckCircle2 class="size-3" />
												Completed
											</span>
										{:else}
											<span
												class="inline-flex items-center gap-1 text-[10px] font-mono uppercase px-2 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 border border-indigo-500/30"
											>
												<Calendar class="size-3" />
												Scheduled
											</span>
										{/if}
										<span class="text-xs text-muted-foreground font-mono">
											{formatDT(m.scheduled_start)} – {formatDT(m.scheduled_end)}
										</span>
									</div>
									<h4 class="text-sm font-semibold text-foreground truncate">{m.title}</h4>
									{#if m.description}
										<p class="text-xs text-muted-foreground line-clamp-2 leading-relaxed">
											{m.description}
										</p>
									{/if}
								</div>

								<!-- Action Buttons -->
								<div class="flex items-center gap-1.5 shrink-0">
									{#if m.status === 'scheduled'}
										<Button
											variant="ghost"
											size="sm"
											class="h-7 text-xs text-blue-400 hover:text-blue-300 hover:bg-blue-500/10 gap-1 px-2"
											title="Start Maintenance Now"
											onclick={() => handleStatusChange(m, 'in_progress')}
										>
											<Play class="size-3" />
											Start
										</Button>
									{:else if m.status === 'in_progress'}
										<Button
											variant="ghost"
											size="sm"
											class="h-7 text-xs text-emerald-400 hover:text-emerald-300 hover:bg-emerald-500/10 gap-1 px-2"
											title="Mark as Completed"
											onclick={() => handleStatusChange(m, 'completed')}
										>
											<CheckCircle2 class="size-3" />
											Complete
										</Button>
									{/if}

									<!-- Edit Button -->
									<Button
										variant="ghost"
										size="sm"
										class="h-7 w-7 p-0 text-muted-foreground hover:text-blue-400 hover:bg-blue-500/10"
										title="Edit Maintenance"
										onclick={() => openEditForm(m)}
									>
										<Pencil class="size-3.5" />
									</Button>

									<!-- Delete Button -->
									<Button
										variant="ghost"
										size="sm"
										class="h-7 w-7 p-0 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
										title="Delete Maintenance"
										onclick={() => handleDelete(m)}
									>
										<Trash2 class="size-3.5" />
									</Button>
								</div>
							</div>

							{#if m.affected_monitors && m.affected_monitors.length > 0}
								<div class="pt-2 border-t border-border/40 flex flex-wrap gap-1.5 items-center">
									<span class="text-[10px] text-muted-foreground">Affected:</span>
									{#each m.affected_monitors as monName (monName)}
										<span
											class="text-[10px] font-mono px-1.5 py-0.5 rounded bg-white/5 border border-border/60 text-foreground/80"
										>
											{monName}
										</span>
									{/each}
								</div>
							{/if}
						</div>
					{/each}
				</div>
			{/if}
		</div>
	</Dialog.Content>
</Dialog.Root>
