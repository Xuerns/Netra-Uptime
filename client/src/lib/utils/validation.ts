export function validateEmail(email: string): string {
	if (!email) return 'Email is required';
	if (email.length > 50) return 'Email maximum 50 characters';
	const emailRegex =
		/^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z]{2,})+$/;
	if (!emailRegex.test(email)) return 'Invalid email format';
	return '';
}

export function validatePassword(password: string): string {
	if (!password) return 'Password is required';
	if (password.length < 8) return 'Password minimum 8 characters';
	if (password.length > 50) return 'Password maximum 50 characters';
	return '';
}

export function validateName(name: string): string {
	if (!name) return 'Name is required';
	if (name.length < 3) return 'Name minimum 3 characters';
	if (name.length > 30) return 'Name maximum 30 characters';

	// nohtml check (prevent < and >)
	const htmlRegex = /[<>]/;
	if (htmlRegex.test(name)) return 'Name cannot contain HTML tags';

	return '';
}

export function validateConfirmPassword(password: string, confirm: string): string {
	if (!confirm) return 'Confirm password is required';
	if (password !== confirm) return 'Password does not match';
	return '';
}

export function validateOrgName(name: string): string {
	if (!name) return 'Organization name is required';
	if (name.length < 3) return 'Name minimum 3 characters';
	if (name.length > 24) return 'Name maximum 24 characters';
	if (/[<>]/.test(name)) return 'Name cannot contain HTML tags';
	return '';
}

export function validateOrgSlug(slug: string): string {
	if (!slug) return 'URL slug is required';
	if (slug.length < 6) return 'Slug minimum 6 characters';
	if (slug.length > 24) return 'Slug maximum 24 characters';
	if (/[<>]/.test(slug)) return 'Slug cannot contain HTML tags';
	if (!/^[\x20-\x7E]*$/.test(slug)) return 'Slug contains invalid characters';
	return '';
}

export function validateMonitorName(name: string): string {
	if (!name) return 'Monitor name is required';
	if (name.length < 4) return 'Name minimum 4 characters';
	if (name.length > 100) return 'Name maximum 100 characters';
	if (/[<>]/.test(name)) return 'Name cannot contain HTML tags';
	return '';
}

export function validateMonitorTarget(target: string): string {
	if (!target) return 'Target is required';
	if (target.length < 4) return 'Target minimum 4 characters';
	if (target.length > 100) return 'Target maximum 100 characters';
	if (/[<>]/.test(target)) return 'Target cannot contain HTML tags';
	return '';
}

export function validateMonitorInterval(interval: number): string {
	if (interval === undefined || interval === null || Number.isNaN(interval))
		return 'Interval is required';
	if (interval < 30) return 'Interval must be at least 30 seconds';
	return '';
}

export function validateMonitorTimeout(timeout: number, interval?: number): string {
	if (timeout === undefined || timeout === null || Number.isNaN(timeout))
		return 'Timeout is required';
	if (timeout < 1) return 'Timeout must be at least 1 second';
	if (timeout > 30) return 'Timeout cannot exceed 30 seconds';
	if (interval !== undefined && interval !== null && interval > 0) {
		if (timeout >= interval)
			return (
				'Timeout must be strictly less than check interval (maximum ' + (interval - 1) + ' seconds)'
			);
	}
	return '';
}

export function validateIncidentMessage(message: string): string {
	const trimmed = message.trim();
	if (!trimmed) return 'Message is required';
	if (trimmed.length < 10) return 'Message minimum 10 characters';
	if (trimmed.length > 200) return 'Message maximum 200 characters';
	return '';
}
