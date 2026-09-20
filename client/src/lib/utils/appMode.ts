import { env } from '$env/dynamic/public';

/**
 * Checks if the application is running in Self-Hosted mode.
 * Defaults to true if PUBLIC_APP_MODE is not specified.
 * Supports both 'selfhosted' and user alias 'selfhoster'.
 */
export function isSelfHosted(): boolean {
	const mode = (env.PUBLIC_APP_MODE || 'selfhosted').trim().toLowerCase();
	return mode.startsWith('selfhost');
}
