<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Avatar from '$lib/components/ui/avatar/index.js';
	import { Loader2 } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import { invalidateAll } from '$app/navigation';
	import { AuthAPI } from '$lib/api/auth';
	import { parseApiError } from '$lib/utils/parseApiError';

	interface Props {
		user: any;
	}

	let { user }: Props = $props();

	let isUpdatingProfile = $state(false);
	let formData = $state({
		name: ''
	});

	$effect(() => {
		if (user?.name && !formData.name) {
			formData.name = user.name;
		}
	});

	async function updateProfile() {
		if (!formData.name.trim()) {
			toast.error('Name cannot be empty');
			return;
		}
		isUpdatingProfile = true;
		try {
			const res = await AuthAPI.updateProfile({ name: formData.name.trim() });
			toast.success(res.message || 'Profile updated successfully');
			await invalidateAll();
		} catch (error: unknown) {
			toast.error(parseApiError(error, 'Failed to update profile'));
		} finally {
			isUpdatingProfile = false;
		}
	}
</script>

<Card.Root class="bg-card border-border">
	<Card.Header>
		<Card.Title class="text-foreground flex items-center gap-2">Personal Information</Card.Title>
		<Card.Description class="text-muted-foreground">
			Update your display name and view your registered email.
		</Card.Description>
	</Card.Header>
	<Card.Content class="space-y-6">
		<div class="flex items-center gap-6">
			<Avatar.Root class="w-20 h-20 border border-border">
				<Avatar.Image src={user?.avatar} alt={user?.name} />
				<Avatar.Fallback class="bg-primary/20 text-primary text-xl font-bold">
					{user?.name?.charAt(0)?.toUpperCase() ?? 'U'}
				</Avatar.Fallback>
			</Avatar.Root>
			<div class="space-y-1">
				<h3 class="font-medium text-foreground text-lg">{user?.name}</h3>
				<p class="text-sm text-muted-foreground">{user?.email}</p>
			</div>
		</div>

		<form
			onsubmit={(e) => {
				e.preventDefault();
				updateProfile();
			}}
			class="space-y-4"
		>
			<div class="space-y-2">
				<Label for="name" class="text-muted-foreground">Display Name</Label>
				<Input
					id="name"
					bind:value={formData.name}
					placeholder="Your full name"
					class="bg-secondary border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-primary h-11 transition-colors"
				/>
			</div>
			<div class="space-y-2">
				<Label for="email" class="text-muted-foreground">Email Address</Label>
				<Input
					id="email"
					value={user?.email}
					disabled
					class="bg-muted border-border text-muted-foreground cursor-not-allowed h-11"
				/>
				<p class="text-xs text-muted-foreground">Email address cannot be changed at this moment.</p>
			</div>

			<div class="flex justify-end pt-2">
				<Button
					type="submit"
					disabled={isUpdatingProfile}
					class="bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-foreground active:scale-[0.98] transition-all duration-200 h-11 px-6 font-medium shadow-sm cursor-pointer"
				>
					{#if isUpdatingProfile}
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
