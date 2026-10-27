/**
 * Centralized application logger.
 * Only logs to console when running in development mode (import.meta.env.DEV).
 */

export function debug(...args: unknown[]): void {
	if (import.meta.env.DEV) {
		console.debug(...args);
	}
}

export function info(...args: unknown[]): void {
	if (import.meta.env.DEV) {
		console.info(...args);
	}
}

export function warn(...args: unknown[]): void {
	if (import.meta.env.DEV) {
		console.warn(...args);
	}
}

export function error(...args: unknown[]): void {
	if (import.meta.env.DEV) {
		console.error(...args);
	}
}

export const logger = {
	debug,
	info,
	warn,
	error
};

export default logger;
