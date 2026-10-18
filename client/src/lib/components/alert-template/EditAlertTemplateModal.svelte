<script lang="ts">
	import type { AlertTemplate } from '$lib/api/alertTemplate';
	import AlertTemplateModal from './AlertTemplateModal.svelte';

	interface Props {
		open?: boolean;
		orgId: string;
		template: AlertTemplate | null;
		onSuccess?: (updated: AlertTemplate) => void;
	}

	let { open = $bindable(false), orgId, template, onSuccess }: Props = $props();
</script>

<AlertTemplateModal bind:open {orgId} mode="edit" {template} {onSuccess} />
