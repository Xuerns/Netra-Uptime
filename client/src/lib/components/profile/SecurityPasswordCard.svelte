<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { toast } from 'svelte-sonner';
	import { Loader2, ChevronDown, Eye, EyeOff } from 'lucide-svelte';
	import { slide } from 'svelte/transition';
	import { AuthAPI } from '$lib/api/auth';
	import { parseApiError } from '$lib/utils/parseApiError';

	interface Props {
		user: any;
	}

	let { user }: Props = $props();

	let isSecurityOpen = $state(false);
	let isChangingPassword = $state(false);
	let passwordStep = $state<'init' | 'verify'>('init');

	let passwordData = $state({
		currentPassword: '',
		newPassword: '',
		confirmPassword: '',
		code: ''
	});

	let showCurrentPassword = $state(false);
	let showNewPassword = $state(false);
	let showConfirmPassword = $state(false);

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
			await AuthAPI.initPasswordChange({
				current_password: passwordData.currentPassword
			});
			toast.success('Verification code sent to your email');
			passwordStep = 'verify';
		} catch (error: unknown) {
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
			await AuthAPI.verifyPasswordChange({
				code: passwordData.code,
				new_password: passwordData.newPassword
			});
			toast.success('Password changed successfully');
			passwordStep = 'init';
			passwordData = { currentPassword: '', newPassword: '', confirmPassword: '', code: '' };
			isSecurityOpen = false;
		} catch (error: unknown) {
			toast.error(parseApiError(error, 'Failed to verify password change'));
		} finally {
			isChangingPassword = false;
		}
	}
</script>

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
								<span class="font-semibold text-foreground capitalize">{user.auth_provider}</span>.
								Password changes are managed by your provider.
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
									<Label for="currentPassword" class="text-muted-foreground">Current Password</Label
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
											class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground cursor-pointer"
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
											class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground cursor-pointer"
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
											class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground cursor-pointer"
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
										class="bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-foreground active:scale-[0.98] transition-all duration-200 h-11 px-6 font-medium shadow-sm cursor-pointer"
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
										class="text-muted-foreground hover:text-foreground h-11 cursor-pointer"
									>
										Cancel
									</Button>
									<Button
										type="submit"
										disabled={isChangingPassword || passwordData.code.length !== 6}
										class="bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-foreground active:scale-[0.98] transition-all duration-200 h-11 px-6 font-medium shadow-sm cursor-pointer"
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
