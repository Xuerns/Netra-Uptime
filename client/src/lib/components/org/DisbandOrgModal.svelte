<script lang="ts">
	import { goto } from '$app/navigation';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import * as Alert from '$lib/components/ui/alert/index.js';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';
	import { AlertTriangle } from 'lucide-svelte';
	import { OrgAPI } from '$lib/api/organization';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { toast } from 'svelte-sonner';

	interface Props {
		open?: boolean;
		orgId: string;
		orgName: string;
		onSuccess?: () => void;
	}

	let { open = $bindable(false), orgId, orgName, onSuccess }: Props = $props();

	let confirmOrgName = $state('');

	async function handleDisband() {
		if (confirmOrgName !== orgName) {
			toast.error('Organization name does not match');
			return;
		}

		try {
			await OrgAPI.disband(orgId);
			if (typeof window !== 'undefined') {
				localStorage.removeItem('netra_last_org');
			}
			toast.success('Organization disbanded successfully');
			open = false;
			confirmOrgName = '';
			if (onSuccess) {
				onSuccess();
			} else {
				goto('/dashboard/org');
			}
		} catch (error: unknown) {
			toast.error(parseApiError(error, 'Failed to disband organization'));
		}
	}
</script>

<ConfirmDialog
	bind:open
	title="Disband Organization"
	confirmText="Disband Organization"
	destructive={true}
	onConfirm={handleDisband}
>
	<div class="space-y-4">
		<p class="text-muted-foreground">
			Are you absolutely sure? This will permanently delete <strong>{orgName}</strong> and all associated
			monitors, status pages, and members.
		</p>
		<Alert.Root variant="destructive" class="border-destructive/30 bg-destructive/10">
			<AlertTriangle class="size-4" />
			<Alert.Description class="text-xs font-medium">
				This action cannot be undone.
			</Alert.Description>
		</Alert.Root>
		<div class="space-y-2 pt-2">
			<Label for="confirmName" class="text-muted-foreground">
				Please type <strong>{orgName}</strong> to confirm.
			</Label>
			<Input
				id="confirmName"
				bind:value={confirmOrgName}
				class="bg-secondary border-border text-foreground focus-visible:ring-red-500 transition-colors"
			/>
		</div>
	</div>
</ConfirmDialog>
