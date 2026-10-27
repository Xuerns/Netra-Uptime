export interface AlertRule {
	id: string;
	organization_id: string;
	monitor_id?: string;
	name: string;
	metric: 'status_code' | 'response_time_ms' | 'timeout';
	operator: '==' | '!=' | '>' | '>=' | '<' | '<=' | 'in' | 'not_in';
	target_value: string;
	consecutive: number;
	severity: 'critical' | 'warning';
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

export interface CreateAlertRuleRequest {
	name: string;
	metric: 'status_code' | 'response_time_ms' | 'timeout';
	operator: '==' | '!=' | '>' | '>=' | '<' | '<=' | 'in' | 'not_in';
	target_value: string;
	consecutive: number;
	severity: 'critical' | 'warning';
	is_active?: boolean;
}

export interface UpdateAlertRuleRequest {
	name?: string;
	metric?: 'status_code' | 'response_time_ms' | 'timeout';
	operator?: '==' | '!=' | '>' | '>=' | '<' | '<=' | 'in' | 'not_in';
	target_value?: string;
	consecutive?: number;
	severity?: 'critical' | 'warning';
	is_active?: boolean;
}
