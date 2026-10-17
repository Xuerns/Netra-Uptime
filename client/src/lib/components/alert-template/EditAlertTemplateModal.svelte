<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import {
		AlertTemplateAPI,
		type AlertTemplate,
		type UpdateAlertTemplateRequest
	} from '$lib/api/alertTemplate';
	import { toast } from 'svelte-sonner';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { Pencil } from 'lucide-svelte';

	interface Props {
		open?: boolean;
		orgId: string;
		template: AlertTemplate | null;
		onSuccess?: (updated: AlertTemplate) => void;
	}

	let { open = $bindable(false), orgId, template, onSuccess }: Props = $props();

	let editName = $state('');
	let editDesc = $state('');
	let editIsDefault = $state(false);
	let isSaving = $state(false);

	$effect(() => {
		if (open && template) {
			editName = template.name;
			editDesc = template.description || '';
			editIsDefault = template.is_default;
			isSaving = false;
		}
	});

	async function handleSaveTemplate(e: Event) {
		e.preventDefault();
		if (!editName.trim() || !orgId || !template?.id) return;

		isSaving = true;
		try {
			const payload: UpdateAlertTemplateRequest = {
				name: editName.trim(),
				description: editDesc.trim(),
				is_default: editIsDefault
			};
			const res = await AlertTemplateAPI.update(orgId, template.id, payload);
			toast.success('Template updated successfully');
			open = false;
			onSuccess?.(res.data);
		} catch (err: unknown) {
			toast.error(parseApiError(err, 'Failed to update template'));
		} finally {
			isSaving = false;
		}
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content
		class="w-[calc(100%-2rem)] sm:max-w-[540px]! bg-card border border-border shadow-2xl p-6 rounded-2xl text-foreground"
	>
		<Dialog.Header>
			<div class="flex items-center gap-2.5">
				<div
					class="size-8 rounded-lg bg-primary/10 border border-primary/20 text-primary flex items-center justify-center shrink-0"
				>
					<Pencil class="size-4" />
				</div>
				<Dialog.Title class="text-lg font-semibold tracking-tight text-foreground">
					Edit Alert Template
				</Dialog.Title>
			</div>
			<Dialog.Description class="text-xs text-muted-foreground mt-1">
				Update profile name and description for this alert template.
			</Dialog.Description>
		</Dialog.Header>

		<div class="h-px w-full bg-border my-2"></div>

		<form onsubmit={handleSaveTemplate} class="space-y-4">
			<div class="space-y-1.5">
				<Label for="edit-template-name" class="text-xs font-medium">Template Name</Label>
				<Input
					id="edit-template-name"
					bind:value={editName}
					placeholder="e.g. Production Web Services"
					required
					class="bg-background border-border text-foreground h-10"
				/>
			</div>

			<div class="space-y-1.5">
				<Label for="edit-template-desc" class="text-xs font-medium">Description</Label>
				<textarea
					id="edit-template-desc"
					bind:value={editDesc}
					placeholder="Summary of services this template is intended for..."
					rows={3}
					class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary"
				></textarea>
			</div>

			<div
				class="flex items-start gap-2.5 p-3 rounded-lg border border-border bg-white/2 cursor-pointer"
			>
				<input
					type="checkbox"
					id="edit-is-default"
					bind:checked={editIsDefault}
					class="rounded border-border bg-background text-primary focus:ring-primary size-4 mt-0.5 cursor-pointer"
				/>
				<Label for="edit-is-default" class="text-xs cursor-pointer flex-1">
					<span class="font-medium text-foreground">Set as organization default</span>
					<span class="block text-muted-foreground text-[11px] mt-0.5 leading-relaxed">
						Automatically pre-select this template when creating new monitors.
					</span>
				</Label>
			</div>

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
					disabled={isSaving || !editName.trim()}
					class="inline-flex items-center justify-center h-9 px-4 rounded-lg bg-primary text-primary-foreground text-xs font-medium hover:bg-primary/90 disabled:opacity-50 cursor-pointer shadow-sm"
				>
					{isSaving ? 'Saving...' : 'Save Changes'}
				</button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>
