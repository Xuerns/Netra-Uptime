<script lang="ts">
	import type { AlertTemplate } from '$lib/api/alertTemplate';
	import AlertTemplateModal from './AlertTemplateModal.svelte';

	interface Props {
		open?: boolean;
		orgId: string;
		onSuccess?: (template: AlertTemplate) => void;
	}

	let { open = $bindable(false), orgId, onSuccess }: Props = $props();
</script>

<AlertTemplateModal bind:open {orgId} mode="create" {onSuccess} />
