<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import {
		AlertTemplateAPI,
		type AlertTemplate,
		type CreateAlertTemplateRequest,
		type UpdateAlertTemplateRequest
	} from '$lib/api/alertTemplate';
	import { toast } from 'svelte-sonner';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { Layers, Pencil } from 'lucide-svelte';

	interface Props {
		open?: boolean;
		orgId: string;
		mode?: 'create' | 'edit';
		template?: AlertTemplate | null;
		onSuccess?: (template: AlertTemplate) => void;
	}

	let { open = $bindable(false), orgId, mode, template = null, onSuccess }: Props = $props();

	const isEdit = $derived(mode ? mode === 'edit' : !!template);

	let templateName = $state('');
	let templateDesc = $state('');
	let isDefault = $state(false);
	let isSaving = $state(false);

	$effect(() => {
		if (open) {
			if (isEdit && template) {
				templateName = template.name;
				templateDesc = template.description || '';
				isDefault = template.is_default;
			} else {
				templateName = '';
				templateDesc = '';
				isDefault = false;
			}
			isSaving = false;
		}
	});

	async function handleSaveTemplate(e: Event) {
		e.preventDefault();
		if (!templateName.trim() || !orgId) return;

		isSaving = true;
		try {
			if (isEdit) {
				if (!template?.id) return;
				const payload: UpdateAlertTemplateRequest = {
					name: templateName.trim(),
					description: templateDesc.trim(),
					is_default: isDefault
				};
				const res = await AlertTemplateAPI.update(orgId, template.id, payload);
				toast.success('Template updated successfully');
				open = false;
				onSuccess?.(res.data);
			} else {
				const payload: CreateAlertTemplateRequest = {
					name: templateName.trim(),
					description: templateDesc.trim(),
					is_default: isDefault,
					rules: []
				};
				const res = await AlertTemplateAPI.create(orgId, payload);
				toast.success('Alert template created');
				open = false;
				onSuccess?.(res.data);
			}
		} catch (err: unknown) {
			toast.error(
				parseApiError(err, isEdit ? 'Failed to update template' : 'Failed to create template')
			);
		} finally {
			isSaving = false;
		}
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content
		class="w-[calc(100%-2rem)] sm:max-w-135! bg-card border border-border shadow-2xl p-6 rounded-2xl text-foreground"
	>
		<Dialog.Header>
			<div class="flex items-center gap-2.5">
				<div
					class="size-8 rounded-lg bg-primary/10 border border-primary/20 text-primary flex items-center justify-center shrink-0"
				>
					{#if isEdit}
						<Pencil class="size-4" />
					{:else}
						<Layers class="size-4" />
					{/if}
				</div>
				<Dialog.Title class="text-lg font-semibold tracking-tight text-foreground">
					{isEdit ? 'Edit Alert Template' : 'Create Alert Template'}
				</Dialog.Title>
			</div>
			<Dialog.Description class="text-xs text-muted-foreground mt-1">
				{isEdit
					? 'Update profile name and description for this alert template.'
					: 'Define a new template profile. You will be directed to configure specific alert rules once created.'}
			</Dialog.Description>
		</Dialog.Header>

		<div class="h-px w-full bg-border my-2"></div>

		<form onsubmit={handleSaveTemplate} class="space-y-4">
			<div class="space-y-1.5">
				<Label for="template-name" class="text-xs font-medium">Template Name</Label>
				<Input
					id="template-name"
					bind:value={templateName}
					placeholder={isEdit ? 'e.g. Production Web Services' : 'e.g. Standard Web Service'}
					required
					class="bg-background border-border text-foreground h-10"
				/>
			</div>

			<div class="space-y-1.5">
				<Label for="template-desc" class="text-xs font-medium">
					Description {#if !isEdit}<span class="text-muted-foreground font-normal">(optional)</span
						>{/if}
				</Label>
				<Textarea
					id="template-desc"
					bind:value={templateDesc}
					placeholder={isEdit
						? 'Summary of services this template is intended for...'
						: 'Briefly describe what services or endpoints this template is intended for...'}
					rows={3}
					class="bg-background border-border text-foreground"
				/>
			</div>

			<label
				for="template-is-default"
				class="flex items-start gap-3 p-3 rounded-lg border border-border bg-background cursor-pointer select-none"
			>
				<Checkbox id="template-is-default" bind:checked={isDefault} class="mt-0.5" />
				<div class="space-y-0.5 text-xs">
					<p class="font-medium text-foreground">Set as organization default</p>
					<p class="text-muted-foreground text-[11px] leading-relaxed">
						Automatically pre-select this template when creating new monitors.
					</p>
				</div>
			</label>

			<Dialog.Footer class="gap-2 sm:gap-0 pt-3">
				<button
					type="button"
					onclick={() => (open = false)}
					class="inline-flex items-center justify-center h-9 px-4 rounded-lg border border-border text-xs font-medium text-muted-foreground hover:text-foreground cursor-pointer"
				>
					Cancel
				</button>
				<button
					type="submit"
					disabled={isSaving || !templateName.trim()}
					class="inline-flex items-center justify-center h-9 px-4 rounded-lg bg-primary text-primary-foreground text-xs font-medium hover:bg-primary/90 disabled:opacity-50 cursor-pointer shadow-sm"
				>
					{#if isSaving}
						{isEdit ? 'Saving...' : 'Creating...'}
					{:else}
						{isEdit ? 'Save Changes' : 'Create & Configure Rules'}
					{/if}
				</button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>
