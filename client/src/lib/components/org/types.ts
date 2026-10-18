export interface OrgData {
	ID: string;
	Name: string;
	Slug: string;
	CreatedAt: string;
	Monitors?: Array<{ ID?: string; Status: string }>;
	OrganizationMembers?: Array<any>;
}
