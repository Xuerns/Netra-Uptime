import {
	ScrollText,
	Activity,
	Building2,
	Users,
	Globe,
	Wrench,
	FolderTree,
	KeyRound,
	Bell,
	Send,
	AlertTriangle
} from 'lucide-svelte';
import type { AuditCategory } from './AuditLogToolbar.svelte';
import type { ActionFilterGroup } from './AuditLogTable.svelte';

export const AUDIT_CATEGORIES: AuditCategory[] = [
	{ id: 'ALL', label: 'All Categories', icon: ScrollText, entityTypes: '' },
	{ id: 'MONITOR', label: 'Monitors', icon: Activity, entityTypes: 'monitor' },
	{ id: 'ORGANIZATION', label: 'Organization', icon: Building2, entityTypes: 'organization' },
	{
		id: 'MEMBER',
		label: 'Members & Invites',
		icon: Users,
		entityTypes: 'organization_member,organization_invitation'
	},
	{ id: 'STATUSPAGE', label: 'Status Pages', icon: Globe, entityTypes: 'status_page' },
	{
		id: 'MAINTENANCE',
		label: 'Maintenance',
		icon: Wrench,
		entityTypes: 'scheduled_maintenance'
	},
	{
		id: 'COMPONENT_GROUP',
		label: 'Component Groups',
		icon: FolderTree,
		entityTypes: 'component_group'
	},
	{ id: 'API_KEY', label: 'API Keys', icon: KeyRound, entityTypes: 'api_key' },
	{
		id: 'ALERT_TEMPLATE',
		label: 'Alert Templates & Rules',
		icon: Bell,
		entityTypes: 'alert_template,alert_rule'
	},
	{
		id: 'NOTIFICATION',
		label: 'Notifications',
		icon: Send,
		entityTypes: 'notification_channel'
	},
	{ id: 'INCIDENT', label: 'Incidents', icon: AlertTriangle, entityTypes: 'incident' }
];

export const AUDIT_ACTION_GROUPS: ActionFilterGroup[] = [
	{
		categoryId: 'MONITOR',
		label: 'Monitors',
		icon: Activity,
		actions: [
			{ value: 'MONITOR_CREATE', label: 'Monitor Created' },
			{ value: 'MONITOR_UPDATE', label: 'Monitor Updated' },
			{ value: 'MONITOR_DELETE', label: 'Monitor Deleted' }
		]
	},
	{
		categoryId: 'ORGANIZATION',
		label: 'Organization',
		icon: Building2,
		actions: [
			{ value: 'ORG_UPDATE', label: 'Organization Updated' },
			{ value: 'ORG_DISBAND', label: 'Organization Disbanded' },
			{ value: 'ORG_TRANSFER_OWNERSHIP', label: 'Ownership Transferred' }
		]
	},
	{
		categoryId: 'MEMBER',
		label: 'Members & Invites',
		icon: Users,
		actions: [
			{ value: 'ORG_MEMBER_INVITE', label: 'Member Invited' },
			{ value: 'ORG_MEMBER_JOIN', label: 'Member Joined' },
			{ value: 'ORG_MEMBER_ACCEPT', label: 'Invite Accepted' },
			{ value: 'ORG_INVITE_REJECT', label: 'Invite Rejected' },
			{ value: 'ORG_MEMBER_PROMOTE', label: 'Member Promoted' },
			{ value: 'ORG_MEMBER_DEMOTE', label: 'Member Demoted' },
			{ value: 'ORG_MEMBER_KICK', label: 'Member Kicked' },
			{ value: 'ORG_MEMBER_LEAVE', label: 'Member Left' }
		]
	},
	{
		categoryId: 'STATUSPAGE',
		label: 'Status Pages',
		icon: Globe,
		actions: [
			{ value: 'STATUSPAGE_CREATE', label: 'Status Page Created' },
			{ value: 'STATUSPAGE_UPDATE', label: 'Status Page Updated' },
			{ value: 'STATUSPAGE_DELETE', label: 'Status Page Deleted' }
		]
	},
	{
		categoryId: 'MAINTENANCE',
		label: 'Scheduled Maintenance',
		icon: Wrench,
		actions: [
			{ value: 'MAINTENANCE_CREATE', label: 'Maintenance Created' },
			{ value: 'MAINTENANCE_UPDATE', label: 'Maintenance Updated' },
			{ value: 'MAINTENANCE_DELETE', label: 'Maintenance Deleted' }
		]
	},
	{
		categoryId: 'COMPONENT_GROUP',
		label: 'Component Groups',
		icon: FolderTree,
		actions: [
			{ value: 'COMPONENT_GROUP_CREATE', label: 'Component Group Created' },
			{ value: 'COMPONENT_GROUP_UPDATE', label: 'Component Group Updated' },
			{ value: 'COMPONENT_GROUP_DELETE', label: 'Component Group Deleted' }
		]
	},
	{
		categoryId: 'API_KEY',
		label: 'API Keys',
		icon: KeyRound,
		actions: [
			{ value: 'API_KEY_CREATE', label: 'API Key Created' },
			{ value: 'API_KEY_REVOKE', label: 'API Key Revoked' }
		]
	},
	{
		categoryId: 'ALERT_TEMPLATE',
		label: 'Alert Templates & Rules',
		icon: Bell,
		actions: [
			{ value: 'ALERT_TEMPLATE_CREATE', label: 'Template Created' },
			{ value: 'ALERT_TEMPLATE_UPDATE', label: 'Template Updated' },
			{ value: 'ALERT_TEMPLATE_DELETE', label: 'Template Deleted' },
			{ value: 'ALERT_RULE_CREATE', label: 'Alert Rule Created' },
			{ value: 'ALERT_RULE_UPDATE', label: 'Alert Rule Updated' },
			{ value: 'ALERT_RULE_DELETE', label: 'Alert Rule Deleted' }
		]
	},
	{
		categoryId: 'NOTIFICATION',
		label: 'Notification Channels',
		icon: Send,
		actions: [
			{ value: 'NOTIFICATION_CHANNEL_CREATE', label: 'Channel Created' },
			{ value: 'NOTIFICATION_CHANNEL_UPDATE', label: 'Channel Updated' },
			{ value: 'NOTIFICATION_CHANNEL_DELETE', label: 'Channel Deleted' }
		]
	},
	{
		categoryId: 'INCIDENT',
		label: 'Incidents',
		icon: AlertTriangle,
		actions: [
			{ value: 'INCIDENT_UPDATE', label: 'Incident Updated' },
			{ value: 'INCIDENT_ASSIGNEE_JOIN', label: 'Assignee Joined' }
		]
	}
];

export function getActionStyle(action: string): string {
	const act = action.toUpperCase();
	if (act.includes('CREATE') || act.includes('JOIN') || act.includes('ACCEPT')) {
		return 'text-emerald-400';
	}
	if (
		act.includes('UPDATE') ||
		act.includes('INVITE') ||
		act.includes('PROMOTE') ||
		act.includes('DEMOTE')
	) {
		return 'text-sky-400';
	}
	if (
		act.includes('DELETE') ||
		act.includes('KICK') ||
		act.includes('DISBAND') ||
		act.includes('TRANSFER') ||
		act.includes('REJECT') ||
		act.includes('LEAVE') ||
		act.includes('REVOKE')
	) {
		return 'text-rose-400';
	}
	return 'text-muted-foreground';
}
