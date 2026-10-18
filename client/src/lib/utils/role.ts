export type OrgRole = 'owner' | 'admin' | 'member' | string;

/**
 * Extracts current user ID regardless of casing (`id` vs `ID`).
 */
export function getCurrentUserId(userData?: { id?: string; ID?: string } | null): string {
	return userData?.id || userData?.ID || '';
}

/**
 * Finds member matching user ID in members collection handling various property naming conventions.
 */
export function findMemberByUserId<T extends Record<string, any>>(
	members?: T[] | null,
	userId?: string
): T | undefined {
	if (!members || !userId) return undefined;
	return members.find((m) => (m.user_id ?? m.UserID ?? m.id ?? m.ID) === userId);
}

/**
 * Extracts role string from member object with fallback.
 */
export function getMemberRole(
	member?: Record<string, any> | null,
	defaultRole: OrgRole = 'member'
): OrgRole {
	return member?.Role ?? member?.role ?? defaultRole;
}

/**
 * Checks whether role has administrative management permissions in organization.
 */
export function canManageOrg(role?: OrgRole | null): boolean {
	return role === 'owner' || role === 'admin';
}
