import { browser } from '$app/environment';
import en from './locales/en.json';
import id from './locales/id.json';

// Type definitions for translations
type Dictionary = typeof id;

const dictionaries: Record<string, Dictionary> = {
	en,
	id
};

// Default to Indonesian since the original site is in Indonesian
let currentLocale = $state('id');

// Initialize from localStorage if available
if (browser) {
	const storedLocale = localStorage.getItem('locale');
	if (storedLocale && dictionaries[storedLocale]) {
		currentLocale = storedLocale;
	}
}

/**
 * Access translation string using dot notation
 * e.g., t('landing.hero.title')
 */
export function t(keyPath: string): string {
	const keys = keyPath.split('.');
	let current: any = dictionaries[currentLocale] || dictionaries['id'];

	for (const key of keys) {
		if (current[key] === undefined) {
			if (import.meta.env.DEV) {
				console.warn(`Translation key not found: ${keyPath}`);
			}
			return keyPath; // Fallback to returning the key itself
		}
		current = current[key];
	}

	return current as string;
}

export function setLocale(locale: string) {
	if (dictionaries[locale]) {
		currentLocale = locale;
		if (browser) {
			localStorage.setItem('locale', locale);
		}
	}
}

export function getLocale() {
	return currentLocale;
}
