<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { AlertTemplateAPI, type AlertTemplate } from '$lib/api/alertTemplate';
	import { MonitorAPI, type Monitor } from '$lib/api/monitor';
	import { headerStore } from '$lib/stores/headerStore.svelte';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { toast } from 'svelte-sonner';
	import { ArrowLeft, Pencil, Trash2, AlertTriangle } from 'lucide-svelte';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { DataSkeleton } from '$lib/components/common';
	import {
		EditAlertTemplateModal,
		AlertRulesSection,
		LinkedMonitorsSection
	} from '$lib/components/alert-template/index.js';

	// ── Params ────────────────────────────────────────────────────────────────
	const orgId = $derived($page.params.id as string);
	const templateId = $derived($page.params.templateId as string);

	// ── State ─────────────────────────────────────────────────────────────────
	let template = $state<AlertTemplate | null>(null);
	let linkedMonitors = $state<Monitor[]>([]);
	let isLoading = $state(true);
	let errorMsg = $state('');

	// Edit Template Info Modal
	let isEditTemplateOpen = $state(false);

	// Delete Template Dialog
	let isDeleteTemplateConfirmOpen = $state(false);
	let isDeletingTemplate = $state(false);

	// ── Load Data ─────────────────────────────────────────────────────────────
	async function loadTemplateData() {
		if (!orgId || !templateId) return;
		isLoading = true;
		errorMsg = '';
		try {
			const [tmplRes, monitorsRes] = await Promise.all([
				AlertTemplateAPI.getById(orgId, templateId),
				MonitorAPI.getAll(orgId).catch(() => ({ data: [] }))
			]);

			if (tmplRes.data) {
				template = tmplRes.data;
				const allMonitors = monitorsRes.data || [];
				linkedMonitors = allMonitors.filter((m: Monitor) => m.AlertTemplateID === templateId);
			} else {
				errorMsg = 'Alert template not found';
			}
		} catch (err: unknown) {
			errorMsg = parseApiError(err, 'Failed to load alert template');
		} finally {
			isLoading = false;
		}
	}

	onMount(() => {
		loadTemplateData();
	});

	// Synchronize top navigation header
	$effect(() => {
		headerStore.set({
			title: template?.name || 'Alert Template Detail',
			showSearch: false,
			actionType: null,
			badge: undefined
		});
	});

	async function handleDeleteTemplate() {
		if (!orgId || !templateId) return;
		isDeletingTemplate = true;
		try {
			await AlertTemplateAPI.delete(orgId, templateId);
			toast.success('Alert template deleted');
			isDeleteTemplateConfirmOpen = false;
			goto(`/dashboard/org/${orgId}/alert-templates`);
		} catch (err: unknown) {
			toast.error(parseApiError(err, 'Failed to delete template'));
		} finally {
			isDeletingTemplate = false;
		}
	}
</script>

<svelte:head>
	<title>{template ? `${template.name} | Alert Templates` : 'Alert Template | Netra Uptime'}</title>
</svelte:head>

<div class="text-foreground antialiased space-y-6">
	<!-- ── Back Link ─────────────────────────────────────────────────────────── -->
	<div>
		<a
			href={`/dashboard/org/${orgId}/alert-templates`}
			class="inline-flex items-center gap-1.5 text-xs font-medium text-muted-foreground hover:text-foreground transition-colors group cursor-pointer"
		>
			<ArrowLeft class="size-3.5 transition-transform group-hover:-translate-x-0.5" />
			<span>Back to Alert Templates</span>
		</a>
	</div>

	{#if isLoading}
		<DataSkeleton count={3} />
	{:else if errorMsg}
		<div
			class="rounded-xl border border-destructive/20 bg-destructive/10 p-8 text-center text-destructive"
		>
			<AlertTriangle class="h-8 w-8 mx-auto mb-2 opacity-80" />
			<p class="font-medium">{errorMsg}</p>
			<button
				onclick={loadTemplateData}
				class="mt-4 inline-flex items-center gap-1.5 bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground h-9 px-4 rounded-lg font-medium text-xs cursor-pointer"
			>
				Try Again
			</button>
		</div>
	{:else if template}
		<!-- ── Dedicated Header ──────────────────────────────────────────────────── -->
		<header
			class="flex flex-col md:flex-row md:items-start justify-between gap-4 pb-5 border-b border-border"
		>
			<div class="space-y-1.5 min-w-0 max-w-[70ch]">
				<div class="flex items-center gap-2.5 flex-wrap">
					<h1 class="text-2xl md:text-3xl font-semibold tracking-tight text-foreground truncate">
						{template.name}
					</h1>
					{#if template.is_default}
						<Badge
							variant="outline"
							class="px-2 py-0.5 text-xs font-medium bg-primary/10 border-primary/20 text-primary"
						>
							Default
						</Badge>
					{/if}
				</div>
				{#if template.description}
					<p class="text-sm text-muted-foreground leading-relaxed">
						{template.description}
					</p>
				{/if}
			</div>

			<!-- Action Buttons -->
			<div class="flex items-center gap-2 shrink-0">
				<button
					onclick={() => (isEditTemplateOpen = true)}
					class="inline-flex items-center gap-1.5 bg-secondary border border-border text-foreground hover:text-primary hover:border-primary/50 active:scale-[0.98] transition-all duration-200 h-9 px-3.5 rounded-lg font-medium text-xs cursor-pointer"
				>
					<Pencil class="size-3.5" />
					<span>Edit Info</span>
				</button>
				<button
					onclick={() => (isDeleteTemplateConfirmOpen = true)}
					class="inline-flex items-center gap-1.5 bg-destructive/10 border border-destructive/20 text-destructive hover:bg-destructive hover:text-destructive-foreground active:scale-[0.98] transition-all duration-200 h-9 px-3.5 rounded-lg font-medium text-xs cursor-pointer"
				>
					<Trash2 class="size-3.5" />
					<span>Delete</span>
				</button>
			</div>
		</header>

		<!-- ── Section: Configured Rules (Component) ──────────────────────────────── -->
		<AlertRulesSection {template} {orgId} onRefresh={loadTemplateData} />

		<!-- ── Section: Linked Monitors (Component) ──────────────────────────────── -->
		<LinkedMonitorsSection monitors={linkedMonitors} {orgId} />
	{/if}
</div>

<!-- ── Edit Template Modal Component ────────────────────────────────────────── -->
<EditAlertTemplateModal
	bind:open={isEditTemplateOpen}
	{orgId}
	{template}
	onSuccess={loadTemplateData}
/>

<!-- ── Delete Template Confirmation ────────────────────────────────────────── -->
<ConfirmDialog
	open={isDeleteTemplateConfirmOpen}
	title="Delete Alert Template"
	description="Are you sure you want to delete this alert template? Any linked monitors will be disconnected and fall back to system default checks."
	confirmText={isDeletingTemplate ? 'Deleting...' : 'Delete Template'}
	destructive={true}
	onConfirm={handleDeleteTemplate}
	onCancel={() => (isDeleteTemplateConfirmOpen = false)}
/>
