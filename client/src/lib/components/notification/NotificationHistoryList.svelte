<script lang="ts">
	import { CheckCircle, AlertCircle, Clock } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import type { NotificationLog } from '$lib/api/notification';
	import { truncateErrorDetail } from '$lib/utils/parseApiError';

	interface Props {
		logs: NotificationLog[];
		logsTotal: number;
		isLoadingLogs: boolean;
		formatDate: (iso: string) => string;
		onLoadMore: () => void;
	}

	let { logs, logsTotal, isLoadingLogs, formatDate, onLoadMore }: Props = $props();

	let expandedErrors = $state<Record<string, boolean>>({});

	function toggleErrorDetail(id: string) {
		expandedErrors[id] = !expandedErrors[id];
	}

	function channelTypeLabel(type: string): string {
		return { email: 'Email', slack: 'Slack', webhook: 'Webhook', discord: 'Discord' }[type] ?? type;
	}

	function channelTypeBadge(type: string): string {
		return (
			{
				email: 'bg-blue-500/10 text-blue-400 border-blue-500/20',
				slack: 'bg-purple-500/10 text-purple-400 border-purple-500/20',
				webhook: 'bg-orange-500/10 text-orange-400 border-orange-500/20',
				discord: 'bg-indigo-500/10 text-indigo-400 border-indigo-500/20'
			}[type] ?? 'bg-white/5 text-foreground/60'
		);
	}
</script>

{#if logs.length === 0}
	<div
		class="flex flex-col items-center justify-center text-center py-20 px-6 gap-3.5 bg-card border border-border rounded-lg"
	>
		<div
			class="flex items-center justify-center size-14 rounded-lg bg-white/5 text-muted-foreground/60"
		>
			<Clock class="size-6" />
		</div>
		<h2 class="text-[18px] font-semibold text-foreground m-0">No notification logs</h2>
		<p class="text-sm leading-relaxed text-muted-foreground max-w-95 m-0">
			History will appear here when monitor alerts are triggered and dispatched to your channels.
		</p>
	</div>
{:else}
	<div class="flex flex-col gap-2.5">
		{#each logs as log (log.id)}
			<div
				class="flex items-start gap-3 p-4 rounded-lg bg-card border border-border hover:border-primary/30 transition-colors"
			>
				<div class="mt-0.5 shrink-0">
					{#if log.status === 'sent'}
						<CheckCircle class="size-4 text-primary" />
					{:else}
						<AlertCircle class="size-4 text-destructive" />
					{/if}
				</div>

				<div class="flex flex-col min-w-0 flex-1">
					<div class="flex items-center gap-2 flex-wrap">
						<span class="text-sm font-medium text-foreground truncate">{log.message}</span>
						<span
							class="text-[11px] font-medium px-2 py-0.5 rounded-full border {channelTypeBadge(
								log.channel
							)}"
						>
							{channelTypeLabel(log.channel)}
						</span>
					</div>

					{#if log.error_detail}
						<div
							class="mt-2 rounded-lg border border-destructive/20 bg-destructive/5 px-3 py-2 text-xs"
						>
							<div class="flex items-center justify-between gap-2">
								<span class="font-medium text-destructive flex items-center gap-1.5">
									<AlertCircle class="size-3.5 shrink-0" />
									Error Details
								</span>
								{#if log.error_detail.length > 120}
									<button
										type="button"
										onclick={() => toggleErrorDetail(log.id)}
										class="text-[11px] text-muted-foreground hover:text-foreground underline transition-colors cursor-pointer"
									>
										{expandedErrors[log.id] ? 'Show less' : 'Show details'}
									</button>
								{/if}
							</div>
							<p
								class="mt-1 font-mono text-[11px] leading-relaxed text-foreground/80 wrap-break-word whitespace-pre-wrap select-text"
							>
								{expandedErrors[log.id] || log.error_detail.length <= 120
									? log.error_detail
									: truncateErrorDetail(log.error_detail, 120)}
							</p>
						</div>
					{/if}

					<span class="text-xs text-muted-foreground/60 mt-1">{formatDate(log.sent_at)}</span>
				</div>

				<span
					class="shrink-0 text-[10px] font-semibold px-2.5 py-0.5 rounded-full uppercase tracking-wider {log.status ===
					'sent'
						? 'bg-primary/10 text-primary'
						: 'bg-destructive/10 text-destructive'}"
				>
					{log.status}
				</span>
			</div>
		{/each}
	</div>

	{#if logs.length < logsTotal}
		<Button
			type="button"
			onclick={onLoadMore}
			disabled={isLoadingLogs}
			class="self-center mt-3 bg-secondary border border-border text-muted-foreground hover:text-foreground h-9 px-6 text-xs cursor-pointer"
		>
			{isLoadingLogs ? 'Loading…' : 'Load more'}
		</Button>
	{/if}
{/if}
