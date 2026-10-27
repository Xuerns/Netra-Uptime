<script lang="ts">
	import { CheckCircle2, KeyRound, Loader2 } from 'lucide-svelte';
	import { AuthAPI } from '$lib/api/auth';
	import { validatePassword, validateConfirmPassword } from '$lib/utils/validation';
	import { debounce } from '$lib/utils';
	import { parseApiError } from '$lib/utils/parseApiError';
	import PasswordInput from './PasswordInput.svelte';

	interface Props {
		token: string;
		onSuccess?: () => void;
	}

	let { token, onSuccess }: Props = $props();

	let newPassword = $state('');
	let confirmPassword = $state('');
	let resetLoading = $state(false);
	let resetError = $state('');
	let resetSuccess = $state(false);

	let fieldErrors = $state({ newPassword: '', confirmPassword: '' });
	let shake = $state({ newPassword: false, confirmPassword: false });

	const handleNewPasswordInput = debounce((val: string, confirm: string) => {
		const err = validatePassword(val);
		if (err) {
			fieldErrors.newPassword = err;
			shake.newPassword = false;
			setTimeout(() => (shake.newPassword = true), 10);
		} else {
			fieldErrors.newPassword = '';
			shake.newPassword = false;
		}
		if (confirm) handleConfirmPasswordInput(val, confirm);
	}, 500);

	const handleConfirmPasswordInput = debounce((pass: string, confirm: string) => {
		const err = validateConfirmPassword(pass, confirm);
		if (err) {
			fieldErrors.confirmPassword = err;
			shake.confirmPassword = false;
			setTimeout(() => (shake.confirmPassword = true), 10);
		} else {
			fieldErrors.confirmPassword = '';
			shake.confirmPassword = false;
		}
	}, 500);

	async function handleResetPassword(event: Event) {
		event.preventDefault();
		resetError = '';

		const passErr = validatePassword(newPassword);
		const confirmErr = validateConfirmPassword(newPassword, confirmPassword);
		fieldErrors.newPassword = passErr;
		fieldErrors.confirmPassword = confirmErr;
		if (passErr || confirmErr) return;

		if (newPassword !== confirmPassword) {
			resetError = 'Passwords do not match.';
			return;
		}

		if (newPassword.length < 8) {
			resetError = 'Password must be at least 8 characters.';
			return;
		}

		resetLoading = true;

		try {
			await AuthAPI.resetPassword(token, newPassword);
			resetSuccess = true;
			onSuccess?.();
		} catch (err: unknown) {
			resetError = parseApiError(err, 'Failed to reset password.');
		} finally {
			resetLoading = false;
		}
	}
</script>

{#if resetSuccess}
	<!-- Success state -->
	<div class="flex flex-col items-center text-center py-md space-y-md animate-fade-in">
		<div class="w-14 h-14 rounded-full bg-primary/10 flex items-center justify-center">
			<CheckCircle2 class="w-7 h-7 text-primary" />
		</div>
		<div>
			<h1
				class="font-headline-lg text-[24px] font-semibold text-foreground tracking-tight"
				style="text-wrap: balance;"
			>
				Password updated
			</h1>
			<p class="font-body-md text-muted-foreground mt-xs">
				Your password has been changed successfully. You can now log in with your new password.
			</p>
		</div>
		<a
			href="/auth/login"
			class="w-full bg-primary hover:bg-primary/90 text-primary-foreground font-label-md font-semibold py-2.5 px-4 rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 focus:ring-offset-background text-center block"
		>
			Back to login
		</a>
	</div>
{:else}
	<!-- Reset form -->
	<form class="space-y-md" onsubmit={handleResetPassword}>
		<div class="flex flex-col items-center mb-lg">
			<div class="mb-sm">
				<div class="w-14 h-14 rounded-full bg-primary/10 flex items-center justify-center">
					<KeyRound class="w-7 h-7 text-primary" />
				</div>
			</div>
			<h1
				class="font-headline-lg text-[24px] font-semibold text-foreground tracking-tight"
				style="text-wrap: balance;"
			>
				Set new password
			</h1>
			<p class="font-body-md text-muted-foreground text-center mt-xs">
				Choose a strong password for your account.
			</p>
		</div>

		<!-- Error -->
		{#if resetError}
			<div
				class="flex items-start gap-2 p-3 rounded-lg bg-destructive/10 border border-destructive/20 animate-fade-in"
			>
				<p class="font-body-md text-sm text-destructive">{resetError}</p>
			</div>
		{/if}

		<!-- Fields -->
		<div class="space-y-sm">
			<PasswordInput
				id="new-password"
				name="new-password"
				label="New Password"
				bind:value={newPassword}
				error={fieldErrors.newPassword}
				shake={shake.newPassword}
				minlength={8}
				autocomplete="new-password"
				oninput={() => handleNewPasswordInput(newPassword, confirmPassword)}
			/>

			<PasswordInput
				id="confirm-password"
				name="confirm-password"
				label="Confirm Password"
				bind:value={confirmPassword}
				error={fieldErrors.confirmPassword}
				shake={shake.confirmPassword}
				minlength={8}
				autocomplete="new-password"
				oninput={() => handleConfirmPasswordInput(newPassword, confirmPassword)}
			/>
		</div>

		<!-- Submit -->
		<div class="pt-2">
			<button
				class="w-full bg-primary hover:bg-primary/90 text-primary-foreground font-label-md font-semibold py-2.5 px-4 rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 focus:ring-offset-background flex items-center justify-center gap-2"
				type="submit"
				disabled={resetLoading}
			>
				{#if resetLoading}
					<Loader2 class="w-4 h-4 animate-spin" />
					Resetting...
				{:else}
					Reset password
				{/if}
			</button>
		</div>
	</form>
{/if}
