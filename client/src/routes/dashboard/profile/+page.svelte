<script lang="ts">
	import { ProfileInfoCard, SecurityPasswordCard, DangerZoneCard } from '$lib/components/profile';

	let { data } = $props();
	let user = $derived(data.user);
</script>

<svelte:head>
	<title>Profile | Netra Uptime</title>
</svelte:head>

<div class="p-6 max-w-4xl mx-auto space-y-6 text-foreground antialiased">
	<div>
		<h1 class="text-2xl font-bold tracking-tight text-foreground mb-1">Profile Settings</h1>
		<p class="text-sm text-muted-foreground">
			Manage your personal information and account security.
		</p>
	</div>

	<!-- 1. Personal Information Section -->
	<ProfileInfoCard {user} />

	<!-- 2. Account Security & Password Management Section -->
	<SecurityPasswordCard {user} />

	<!-- 3. Danger Zone Section -->
	<DangerZoneCard />
</div>
