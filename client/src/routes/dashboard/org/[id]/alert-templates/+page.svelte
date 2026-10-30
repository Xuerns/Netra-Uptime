<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { headerStore } from '$lib/stores/headerStore.svelte';
	import { Layers, Plus, AlertTriangle, Search } from 'lucide-svelte';
	import { AlertTemplateAPI, type AlertTemplate } from '$lib/api/alertTemplate';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { DataSkeleton } from '$lib/components/common';
	import {
		AlertTemplateRow,
		CreateAlertTemplateModal
	} from '$lib/components/alert-template/index.js';

	const orgId = $derived($page.params.id as string);

	let templates = $state<AlertTemplate[]>([]);
	let isLoading = $state(true);
	let errorMsg = $state('');
	let isCreateModalOpen = $state(false);

	async function loadTemplates() {
		isLoading = true;
		errorMsg = '';
		try {
			const res = await AlertTemplateAPI.list(orgId);
			templates = res.data || [];
		} catch (err: unknown) {
			errorMsg = parseApiError(err, 'Failed to load alert templates');
		} finally {
			isLoading = false;
		}
	}

	onMount(() => {
		if (orgId) {
			loadTemplates();
			if (typeof window !== 'undefined') {
				const params = new URLSearchParams(window.location.search);
				if (params.get('create') === 'true') {
					isCreateModalOpen = true;
				}
			}
		}
	});

	// Synchronize top navigation header
	$effect(() => {
		headerStore.set({
			title: 'Alert Templates',
			showSearch: true,
			searchPlaceholder: 'Search templates by name or description...',
			actionType: 'button',
			actionLabel: 'Create Template',
			onAction: () => {
				isCreateModalOpen = true;
			},
			badge: {
				text: `${templates.length} Template${templates.length !== 1 ? 's' : ''}`,
				variant: 'slate'
			}
		});
	});

	let filteredTemplates = $derived(
		templates.filter((tmpl) => {
			const q = headerStore.searchQuery.toLowerCase().trim();
			if (!q) return true;
			return (
				tmpl.name.toLowerCase().includes(q) ||
				(tmpl.description && tmpl.description.toLowerCase().includes(q))
			);
		})
	);

	function handleCreated(newTemplate: AlertTemplate) {
		if (newTemplate?.id) {
			goto(`/dashboard/org/${orgId}/alert-templates/${newTemplate.id}`);
		} else {
			loadTemplates();
		}
	}
</script>

<svelte:head>
	<title>Alert Templates | Netra Uptime</title>
</svelte:head>

<div class="text-foreground antialiased space-y-6">
	<!-- ── Page Header (Single button placement via headerStore like other pages) ── -->
	<header class="flex flex-wrap items-start justify-between gap-4 mb-6">
		<div>
			<h1 class="text-[32px] font-semibold leading-tight tracking-tight text-foreground mb-1.5">
				Alert Templates
			</h1>
			<p class="text-base leading-relaxed text-muted-foreground max-w-[60ch]">
				Define reusable alert profiles and link them to monitors. Changes to a template
				automatically apply to all linked monitors.
			</p>
		</div>
	</header>

	<!-- ── Templates Content ──────────────────────────────────────────────────── -->
	{#if isLoading}
		<DataSkeleton count={4} />
	{:else if errorMsg}
		<div
			class="rounded-lg border border-destructive/20 bg-destructive/10 p-6 text-center text-destructive"
		>
			<AlertTriangle class="h-8 w-8 mx-auto mb-2 opacity-80" />
			<p class="font-medium">{errorMsg}</p>
			<button
				onclick={loadTemplates}
				class="mt-4 inline-flex items-center gap-1.5 bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground h-9 px-4 rounded-lg font-medium text-xs cursor-pointer"
			>
				Try Again
			</button>
		</div>
	{:else if templates.length === 0}
		<!-- Empty State with create action button -->
		<div
			class="flex flex-col items-center text-center py-20 px-4 border border-border rounded-lg bg-card"
		>
			<div
				class="flex items-center justify-center size-14 rounded-lg bg-primary/10 border border-primary/20 mb-4 text-primary"
			>
				<Layers class="size-6" />
			</div>
			<h3 class="text-xl font-semibold text-foreground mb-1.5">No alert templates yet</h3>
			<p class="text-sm text-muted-foreground max-w-[44ch] mb-6 leading-relaxed">
				Create reusable templates with customized rule thresholds (status codes, latency limits,
				timeouts) to centrally manage incident alerts across all monitors.
			</p>
			<button
				onclick={() => (isCreateModalOpen = true)}
				class="inline-flex items-center gap-1.5 bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground active:scale-[0.98] transition-all duration-200 h-9 px-4 rounded-lg font-medium text-xs cursor-pointer focus:outline-none"
			>
				<Plus class="size-3.5 stroke-[2.5]" />
				<span>Create Template</span>
			</button>
		</div>
	{:else if filteredTemplates.length === 0}
		<!-- No Search Results -->
		<div
			class="flex flex-col items-center text-center py-16 px-4 border border-border rounded-lg bg-card"
		>
			<Search class="size-8 text-muted-foreground/60 mb-2" />
			<h3 class="text-base font-semibold text-foreground mb-1">No templates found</h3>
			<p class="text-xs text-muted-foreground max-w-[36ch]">
				No templates matched your search query "{headerStore.searchQuery}".
			</p>
		</div>
	{:else}
		<!-- ── List / Table View ───────────────────────────────────────────────── -->
		<div class="border border-border rounded-md overflow-hidden">
			<!-- Table Header -->
			<div
				class="hidden md:grid md:grid-cols-12 gap-4 px-4 py-2 border-b border-border bg-table-header text-[11px] font-semibold tracking-wider uppercase text-muted-foreground"
			>
				<div class="col-span-5">Template</div>
				<div class="col-span-2">Rules</div>
				<div class="col-span-2">Monitors</div>
				<div class="col-span-2">Last Updated</div>
				<div class="col-span-1 text-right"></div>
			</div>

			<!-- Rows -->
			<div class="divide-y divide-border/60 bg-card">
				{#each filteredTemplates as tmpl (tmpl.id)}
					<AlertTemplateRow template={tmpl} {orgId} />
				{/each}
			</div>
		</div>
	{/if}
</div>

<!-- ── Create Template Modal Component ──────────────────────────────────────── -->
<CreateAlertTemplateModal bind:open={isCreateModalOpen} {orgId} onSuccess={handleCreated} />
