<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import {
		AlertTemplateAPI,
		type AlertTemplate,
		type CreateAlertTemplateRequest
	} from '$lib/api/alertTemplate';
	import { toast } from 'svelte-sonner';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { Layers } from 'lucide-svelte';

	interface Props {
		open?: boolean;
		orgId: string;
		onSuccess?: (template: AlertTemplate) => void;
	}

	let { open = $bindable(false), orgId, onSuccess }: Props = $props();

	let templateName = $state('');
	let templateDesc = $state('');
	let isDefault = $state(false);
	let isSaving = $state(false);

	$effect(() => {
		if (open) {
			templateName = '';
			templateDesc = '';
			isDefault = false;
			isSaving = false;
		}
	});

	async function handleSaveTemplate(e: Event) {
		e.preventDefault();
		if (!templateName.trim() || !orgId) return;

		isSaving = true;
		try {
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
		} catch (err: unknown) {
			toast.error(parseApiError(err, 'Failed to create template'));
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
					<Layers class="size-4" />
				</div>
				<Dialog.Title class="text-lg font-semibold tracking-tight text-foreground">
					Create Alert Template
				</Dialog.Title>
			</div>
			<Dialog.Description class="text-xs text-muted-foreground mt-1">
				Define a new template profile. You will be directed to configure specific alert rules once
				created.
			</Dialog.Description>
		</Dialog.Header>

		<div class="h-px w-full bg-border my-2"></div>

		<form onsubmit={handleSaveTemplate} class="space-y-4">
			<div class="space-y-1.5">
				<Label for="create-template-name" class="text-xs font-medium">Template Name</Label>
				<Input
					id="create-template-name"
					bind:value={templateName}
					placeholder="e.g. Standard Web Service"
					required
					class="bg-background border-border text-foreground h-10"
				/>
			</div>

			<div class="space-y-1.5">
				<Label for="create-template-desc" class="text-xs font-medium">
					Description <span class="text-muted-foreground font-normal">(optional)</span>
				</Label>
				<textarea
					id="create-template-desc"
					bind:value={templateDesc}
					placeholder="Briefly describe what services or endpoints this template is intended for..."
					rows={3}
					class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary"
				></textarea>
			</div>

			<div
				class="flex items-start gap-2.5 p-3 rounded-lg border border-border bg-white/2 cursor-pointer"
			>
				<input
					type="checkbox"
					id="create-is-default"
					bind:checked={isDefault}
					class="rounded border-border bg-background text-primary focus:ring-primary size-4 mt-0.5 cursor-pointer"
				/>
				<Label for="create-is-default" class="text-xs cursor-pointer flex-1">
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
					disabled={isSaving || !templateName.trim()}
					class="inline-flex items-center justify-center h-9 px-4 rounded-lg bg-primary text-primary-foreground text-xs font-medium hover:bg-primary/90 disabled:opacity-50 cursor-pointer shadow-sm"
				>
					{isSaving ? 'Creating...' : 'Create & Configure Rules'}
				</button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>
