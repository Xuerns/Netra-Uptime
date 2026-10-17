export interface HeaderBadge {
	text: string;
	variant?: 'emerald' | 'amber' | 'rose' | 'slate' | 'blue';
}

export type HeaderActionType = 'monitor' | 'member' | 'button' | 'custom' | null;

export interface HeaderConfig {
	title?: string;
	badge?: HeaderBadge | null;
	showSearch?: boolean;
	searchQuery?: string;
	searchPlaceholder?: string;
	actionType?: HeaderActionType;
	actionLabel?: string;
	onAction?: ((type?: any) => void) | null;
}

class HeaderStore {
	title = $state<string>('');
	badge = $state<HeaderBadge | null>(null);
	showSearch = $state<boolean>(false);
	searchQuery = $state<string>('');
	searchPlaceholder = $state<string>('Search...');
	actionType = $state<HeaderActionType>(null);
	actionLabel = $state<string>('');
	onAction = $state<((type?: any) => void) | null>(null);

	set(config: HeaderConfig) {
		if (config.title !== undefined) this.title = config.title;
		if (config.badge !== undefined) this.badge = config.badge;
		if (config.showSearch !== undefined) this.showSearch = config.showSearch;
		if (config.searchQuery !== undefined) this.searchQuery = config.searchQuery;
		if (config.searchPlaceholder !== undefined) this.searchPlaceholder = config.searchPlaceholder;
		if (config.actionType !== undefined) this.actionType = config.actionType;
		if (config.actionLabel !== undefined) this.actionLabel = config.actionLabel;
		if (config.onAction !== undefined) this.onAction = config.onAction;
	}

	reset() {
		this.title = '';
		this.badge = null;
		this.showSearch = false;
		this.searchQuery = '';
		this.searchPlaceholder = 'Search...';
		this.actionType = null;
		this.actionLabel = '';
		this.onAction = null;
	}
}

export const headerStore = new HeaderStore();
