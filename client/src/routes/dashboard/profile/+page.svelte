<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Avatar from '$lib/components/ui/avatar/index.js';
	import { toast } from 'svelte-sonner';
	import { User, Shield, Loader2, AlertTriangle, ChevronDown, Eye, EyeOff } from 'lucide-svelte';
	import { slide } from 'svelte/transition';
	import { apiFetch } from '$lib/helper/api';
	import { env } from '$env/dynamic/public';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';
	import { parseApiError } from '$lib/utils/parseApiError';

	let { data } = $props();
	let user = $state(data.user);

	let isUpdatingProfile = $state(false);
	let isChangingPassword = $state(false);

	let formData = $state({
		name: user?.name || ''
	});

	let passwordData = $state({
		currentPassword: '',
		newPassword: '',
		confirmPassword: '',
		code: ''
	});
	let passwordStep = $state<'init' | 'verify'>('init');

	let showCurrentPassword = $state(false);
	let showNewPassword = $state(false);
	let showConfirmPassword = $state(false);

	let isDeleteDialogOpen = $state(false);
	let isSecurityOpen = $state(false);

	async function updateProfile() {
		if (!formData.name) {
			toast.error('Name cannot be empty');
			return;
		}
		isUpdatingProfile = true;
		try {
			const res = await apiFetch(`${env.PUBLIC_API_URL}/auth/profile`, {
				method: 'PUT',
				body: JSON.stringify(formData),
				credentials: 'include'
			});

			if (!res.ok) {
				const error = await res.json();
				throw new Error(error.error || 'Failed to update profile');
			}

			const data = await res.json();
			if (data.user) {
				user = data.user;
			} else {
				user.name = formData.name;
			}
			toast.success(data.message || 'Profile updated successfully');
		} catch (error: any) {
			toast.error(parseApiError(error, 'Failed to update profile'));
		} finally {
			isUpdatingProfile = false;
		}
	}

	async function initPasswordChange() {
		if (passwordData.newPassword !== passwordData.confirmPassword) {
			toast.error('New passwords do not match');
			return;
		}
		if (passwordData.newPassword.length < 8) {
			toast.error('Password must be at least 8 characters');
			return;
		}
		isChangingPassword = true;
		try {
			const res = await apiFetch(`${env.PUBLIC_API_URL}/auth/password/change/init`, {
				method: 'POST',
				body: JSON.stringify({
					current_password: passwordData.currentPassword
				}),
				credentials: 'include'
			});

			if (!res.ok) {
				const error = await res.json();
				throw new Error(error.error || 'Failed to initiate password change');
			}

			toast.success('Verification code sent to your email');
			passwordStep = 'verify';
		} catch (error: any) {
			toast.error(parseApiError(error, 'Failed to initiate password change'));
		} finally {
			isChangingPassword = false;
		}
	}

	async function verifyPasswordChange() {
		if (passwordData.code.length !== 6) {
			toast.error('Code must be 6 digits');
			return;
		}
		isChangingPassword = true;
		try {
			const res = await apiFetch(`${env.PUBLIC_API_URL}/auth/password/change/verify`, {
				method: 'POST',
				body: JSON.stringify({
					code: passwordData.code,
					new_password: passwordData.newPassword
				}),
				credentials: 'include'
			});

			if (!res.ok) {
				const error = await res.json();
				throw new Error(error.error || 'Failed to verify password change');
			}

			toast.success('Password changed successfully');
			passwordStep = 'init';
			passwordData = { currentPassword: '', newPassword: '', confirmPassword: '', code: '' };
			isSecurityOpen = false;
		} catch (error: any) {
			toast.error(parseApiError(error, 'Failed to verify password change'));
		} finally {
			isChangingPassword = false;
		}
	}

	async function deleteAccount() {
		try {
			const res = await apiFetch(`${env.PUBLIC_API_URL}/auth/account`, {
				method: 'DELETE',
				credentials: 'include'
			});

			if (!res.ok) {
				const error = await res.json();
				throw new Error(error.error || 'Failed to delete account');
			}

			toast.success('Account deleted successfully');
			// Redirect to home/login
			window.location.href = '/auth/login';
		} catch (error: any) {
			toast.error(parseApiError(error, 'Failed to delete account'));
		}
	}
</script>

<svelte:head>
	<title>Profile | PulseWatch</title>
</svelte:head>

<div class="p-6 max-w-4xl mx-auto space-y-3">
	<div>
		<h1 class="text-2xl font-bold tracking-tight text-foreground mb-2">Profile Settings</h1>
		<p class="text-muted-foreground">Manage your personal information and account security.</p>
	</div>

	<!-- Personal Information -->
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
						{user?.name?.charAt(0)?.toUpperCase()}
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
					<p class="text-xs text-muted-foreground">
						Email address cannot be changed at this moment.
					</p>
				</div>

				<div class="flex justify-end pt-2">
					<Button
						type="submit"
						disabled={isUpdatingProfile}
						class="bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-foreground active:scale-[0.98] transition-all duration-200 h-11 px-6 font-medium shadow-sm"
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

	<!-- Account Security -->
	<Card.Root class="bg-card border-border p-0 gap-0">
		<button
			type="button"
			class="w-full text-left flex items-center justify-between px-6 py-4 cursor-pointer hover:bg-accent transition-colors rounded-t-xl"
			onclick={() => (isSecurityOpen = !isSecurityOpen)}
		>
			<div class="space-y-1.5">
				<Card.Title class="text-foreground flex items-center gap-2">Account Security</Card.Title>
				<Card.Description class="text-muted-foreground">
					Update your password to keep your account secure.
				</Card.Description>
			</div>
			<ChevronDown
				class="size-5 text-muted-foreground transition-transform duration-200 {isSecurityOpen
					? 'rotate-180'
					: ''}"
			/>
		</button>

		{#if isSecurityOpen}
			<div transition:slide={{ duration: 200 }}>
				<Card.Content class="pt-0 border-t border-border">
					<div class="pt-6">
						{#if user?.auth_provider === 'google' || user?.auth_provider === 'github'}
							<div class="bg-secondary border border-border rounded-lg p-4 text-center">
								<p class="text-muted-foreground">
									You are signed in using
									<span class="font-semibold text-foreground capitalize">{user.auth_provider}</span
									>. Password changes are managed by your provider.
								</p>
							</div>
						{:else}
							{#if passwordStep === 'init'}
								<form
									onsubmit={(e) => {
										e.preventDefault();
										initPasswordChange();
									}}
									class="space-y-4"
								>
									<div class="space-y-2">
										<Label for="currentPassword" class="text-muted-foreground"
											>Current Password</Label
										>
										<div class="relative">
											<Input
												id="currentPassword"
												type={showCurrentPassword ? 'text' : 'password'}
												bind:value={passwordData.currentPassword}
												placeholder="Enter current password"
												required
												class="bg-secondary border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-primary h-11 pr-10 transition-colors"
											/>
											<button
												type="button"
												class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
												onclick={() => (showCurrentPassword = !showCurrentPassword)}
											>
												{#if showCurrentPassword}
													<EyeOff class="w-4 h-4" />
												{:else}
													<Eye class="w-4 h-4" />
												{/if}
											</button>
										</div>
									</div>
									<div class="space-y-2">
										<Label for="newPassword" class="text-muted-foreground">New Password</Label>
										<div class="relative">
											<Input
												id="newPassword"
												type={showNewPassword ? 'text' : 'password'}
												bind:value={passwordData.newPassword}
												placeholder="Minimum 8 characters"
												required
												class="bg-secondary border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-primary h-11 pr-10 transition-colors"
											/>
											<button
												type="button"
												class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
												onclick={() => (showNewPassword = !showNewPassword)}
											>
												{#if showNewPassword}
													<EyeOff class="w-4 h-4" />
												{:else}
													<Eye class="w-4 h-4" />
												{/if}
											</button>
										</div>
									</div>
									<div class="space-y-2">
										<div class="flex justify-between items-center">
											<Label for="confirmPassword" class="text-muted-foreground"
												>Confirm New Password</Label
											>
											{#if passwordData.confirmPassword && passwordData.newPassword !== passwordData.confirmPassword}
												<span class="text-xs font-medium text-destructive"
													>Passwords do not match</span
												>
											{/if}
										</div>
										<div class="relative">
											<Input
												id="confirmPassword"
												type={showConfirmPassword ? 'text' : 'password'}
												bind:value={passwordData.confirmPassword}
												placeholder="Confirm new password"
												required
												class="bg-secondary border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-primary h-11 pr-10 transition-colors"
											/>
											<button
												type="button"
												class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
												onclick={() => (showConfirmPassword = !showConfirmPassword)}
											>
												{#if showConfirmPassword}
													<EyeOff class="w-4 h-4" />
												{:else}
													<Eye class="w-4 h-4" />
												{/if}
											</button>
										</div>
									</div>
									<div class="flex justify-end pb-4">
										<Button
											type="submit"
											disabled={isChangingPassword}
											class="bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-foreground active:scale-[0.98] transition-all duration-200 h-11 px-6 font-medium shadow-sm"
										>
											{#if isChangingPassword}
												<Loader2 class="w-4 h-4 mr-2 animate-spin" />
												Sending Code...
											{:else}
												Continue
											{/if}
										</Button>
									</div>
								</form>
							{:else}
								<form
									onsubmit={(e) => {
										e.preventDefault();
										verifyPasswordChange();
									}}
									class="space-y-4"
								>
									<div class="bg-indigo-500/10 border border-indigo-500/20 rounded-lg p-4 mb-4">
										<p class="text-primary text-sm">
											We've sent a 6-digit verification code to your email. Please enter it below to
											confirm your password change.
										</p>
									</div>
									<div class="space-y-2">
										<Label for="code" class="text-muted-foreground">Verification Code</Label>
										<Input
											id="code"
											type="text"
											bind:value={passwordData.code}
											placeholder="123456"
											maxlength={6}
											required
											class="bg-secondary border-border text-foreground text-center text-xl tracking-widest placeholder:text-muted-foreground focus-visible:ring-primary h-12 transition-colors uppercase font-mono"
										/>
									</div>
									<div class="flex justify-end gap-3 pt-2">
										<Button
											type="button"
											variant="ghost"
											onclick={() => {
												passwordStep = 'init';
												passwordData.code = '';
											}}
											class="text-muted-foreground hover:text-foreground h-11"
										>
											Cancel
										</Button>
										<Button
											type="submit"
											disabled={isChangingPassword || passwordData.code.length !== 6}
											class="bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-foreground active:scale-[0.98] transition-all duration-200 h-11 px-6 font-medium shadow-sm"
										>
											{#if isChangingPassword}
												<Loader2 class="w-4 h-4 mr-2 animate-spin" />
												Verifying...
											{:else}
												Verify & Update
											{/if}
										</Button>
									</div>
								</form>
							{/if}
						{/if}
					</div>
				</Card.Content>
			</div>
		{/if}
	</Card.Root>

	<!-- Danger Zone -->
	<Card.Root class="bg-card border border-destructive/30">
		<Card.Header>
			<Card.Title class="text-destructive flex items-center gap-2">Danger Zone</Card.Title>
			<Card.Description class="text-muted-foreground">
				Irreversible and destructive actions for your account.
			</Card.Description>
		</Card.Header>
		<Card.Content>
			<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pt-4">
				<div class="space-y-1">
					<h4 class="font-medium text-foreground">Delete Account</h4>
					<p class="text-sm text-muted-foreground">
						Permanently delete your account and all associated data. This action cannot be undone.
					</p>
				</div>
				<Button
					class="shrink-0 h-11 px-6 bg-secondary border border-border text-destructive hover:bg-destructive hover:border-destructive hover:text-foreground transition-all duration-200 active:scale-[0.98] font-medium shadow-sm"
					onclick={() => (isDeleteDialogOpen = true)}
				>
					Delete Account
				</Button>
			</div>
		</Card.Content>
	</Card.Root>

	<ConfirmDialog
		bind:open={isDeleteDialogOpen}
		title="Delete Account"
		description="Are you absolutely sure you want to delete your account? This action cannot be undone and will permanently delete your data."
		confirmText="Yes, delete my account"
		destructive={true}
		onConfirm={deleteAccount}
	/>
</div>
