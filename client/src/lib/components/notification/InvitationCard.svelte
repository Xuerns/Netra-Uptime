<script lang="ts">
	import { Check, X } from 'lucide-svelte';
	import { getInitials as initials, getAvatarColor as avatarColor } from '$lib/utils/string';

	export interface InvitationUser {
		name?: string;
		avatar_url?: string;
		[key: string]: unknown;
	}

	export interface InvitationOrganization {
		name?: string;
		[key: string]: unknown;
	}

	export interface Invitation {
		ID: string;
		Role: string;
		Status: string;
		Organization?: InvitationOrganization;
		Inviter?: InvitationUser;
		[key: string]: unknown;
	}

	interface Props {
		invitation: Invitation;
		onAccept: (id: string) => void;
		onReject: (id: string) => void;
	}

	let { invitation, onAccept, onReject }: Props = $props();
</script>

<div
	class="flex flex-wrap items-center justify-between gap-6 p-5 rounded-md bg-card border border-border hover:border-primary/30 transition-colors"
>
	<div class="flex items-center gap-4 min-w-0">
		<!-- Inviter Avatar -->
		{#if invitation.Inviter?.avatar_url}
			<img
				src={invitation.Inviter.avatar_url}
				alt={invitation.Inviter.name || 'Inviter'}
				class="size-10 rounded-full object-cover shrink-0"
			/>
		{:else}
			<div
				class="size-10 rounded-full flex items-center justify-center font-semibold text-[14px] shrink-0 {avatarColor(
					invitation.Inviter?.name || 'U'
				)}"
			>
				{initials(invitation.Inviter?.name || 'U')}
			</div>
		{/if}

		<div class="flex flex-col min-w-0">
			<p class="text-[14.5px] text-foreground m-0 leading-snug">
				<span class="font-semibold">{invitation.Inviter?.name || 'Someone'}</span>
				<span class="text-muted-foreground">invited you to join</span>
				<span class="font-semibold text-primary"
					>{invitation.Organization?.name || 'an organization'}</span
				>
			</p>
			<p class="text-[12.5px] text-muted-foreground m-0 mt-1">
				Role: <span class="capitalize text-foreground">{invitation.Role}</span>
			</p>
		</div>
	</div>

	{#if invitation.Status === 'pending'}
		<div class="flex items-center gap-2 shrink-0 ml-auto">
			<button
				class="flex items-center justify-center size-9 rounded-lg border border-border bg-white/2 text-muted-foreground hover:bg-destructive/15 hover:text-destructive hover:border-destructive/20 transition-all cursor-pointer"
				onclick={() => onReject(invitation.ID)}
				aria-label="Reject"
				title="Reject invitation"
			>
				<X class="size-4.5" />
			</button>
			<button
				class="flex items-center gap-2 h-9 px-4 rounded-lg bg-primary text-primary-foreground font-medium text-[13.5px] hover:bg-[#12c98d] transition-colors cursor-pointer"
				onclick={() => onAccept(invitation.ID)}
			>
				<Check class="size-4" />
				<span>Accept</span>
			</button>
		</div>
	{:else}
		<div class="shrink-0 ml-auto">
			<span
				class="text-[12px] font-medium px-3 py-1 rounded-full uppercase tracking-wider
				{invitation.Status === 'accepted'
					? 'bg-primary/10 text-primary'
					: 'bg-destructive/10 text-destructive'}"
			>
				{invitation.Status}
			</span>
		</div>
	{/if}
</div>
