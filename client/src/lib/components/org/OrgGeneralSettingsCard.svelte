<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Loader2 } from 'lucide-svelte';
	import { OrgAPI } from '$lib/api/organization';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { toast } from 'svelte-sonner';

	interface Props {
		orgId: string;
		initialName: string;
		initialSlug: string;
		onUpdated?: (name: string, slug: string) => void;
	}

	let { orgId, initialName, initialSlug, onUpdated }: Props = $props();

	let name = $state('');
	let slug = $state('');
	let isUpdating = $state(false);

	$effect(() => {
		name = initialName;
		slug = initialSlug;
	});

	async function handleUpdateOrganization(e: Event) {
		e.preventDefault();
		if (!name || !slug) {
			toast.error('Name and slug cannot be empty');
			return;
		}

		const slugRegex = /^[a-z0-9-]+$/;
		if (!slugRegex.test(slug)) {
			toast.error('Slug can only contain lowercase letters, numbers, and hyphens');
			return;
		}

		if (slug.length < 6) {
			toast.error('Slug must be at least 6 characters long');
			return;
		}

		isUpdating = true;
		try {
			await OrgAPI.update(orgId, {
				organization_id: orgId,
				name,
				slug
			});
			toast.success('Organization updated successfully');
			if (onUpdated) {
				onUpdated(name, slug);
			}
		} catch (error: unknown) {
			toast.error(parseApiError(error, 'Failed to update organization'));
		} finally {
			isUpdating = false;
		}
	}
</script>

<Card.Root class="bg-card border-border">
	<Card.Header>
		<Card.Title class="text-foreground">General Information</Card.Title>
		<Card.Description class="text-muted-foreground">
			Update your organization's display name and URL slug.
		</Card.Description>
	</Card.Header>
	<Card.Content class="space-y-6">
		<form onsubmit={handleUpdateOrganization} class="space-y-4">
			<div class="space-y-2">
				<Label for="name" class="text-muted-foreground">Organization Name</Label>
				<Input
					id="name"
					bind:value={name}
					placeholder="e.g. Acme Corp"
					class="bg-secondary border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-primary h-11 transition-colors"
				/>
			</div>
			<div class="space-y-2">
				<Label for="slug" class="text-muted-foreground">Organization Slug</Label>
				<div class="flex flex-col sm:flex-row gap-3">
					<div
						class="flex-1 flex items-center bg-muted border border-border rounded-md h-11 px-3 text-muted-foreground font-mono text-sm overflow-hidden"
					>
						<span class="truncate">https://status.netrauptime.com/</span>
					</div>
					<Input
						id="slug"
						bind:value={slug}
						placeholder="acme-corp"
						class="bg-secondary border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-primary h-11 transition-colors flex-1"
					/>
				</div>
				<p class="text-xs text-muted-foreground">
					Must be at least 6 characters. Only lowercase letters, numbers, and hyphens.
				</p>
			</div>

			<div class="flex justify-end pt-4">
				<Button
					type="submit"
					disabled={isUpdating}
					class="bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground active:scale-[0.98] transition-all duration-200 h-11 px-6 font-medium cursor-pointer"
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
