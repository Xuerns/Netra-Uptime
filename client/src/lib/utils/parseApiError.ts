/**
 * Utility for parsing and sanitizing API error messages in the frontend.
 * Acts as a resilient safety net to guarantee that raw Go validator errors,
 * database query dumps, or mail traces never leak to users.
 */

const ACRONYMS: Record<string, string> = {
	id: 'ID',
	url: 'URL',
	api: 'API',
	key: 'Key',
	ssl: 'SSL',
	ip: 'IP',
	pics: 'PICs',
	html: 'HTML',
	http: 'HTTP',
	uuid: 'UUID',
	dns: 'DNS',
	tcp: 'TCP',
	smtp: 'SMTP'
};

/**
 * Humanizes camelCase, PascalCase, and snake_case field names into clean labels.
 * e.g., "MonitorPICs" -> "Monitor PICs", "subdomain" -> "Subdomain", "target_value" -> "Target Value"
 */
export function humanizeField(field: string): string {
	if (!field) return 'Field';

	if (field === 'MonitorPICs' || field === 'assignees') return 'Assignees';

	const cleaned = field.replace(/[-_]/g, ' ');
	const words = cleaned
		.replace(/([a-z])([A-Z])/g, '$1 $2')
		.replace(/([A-Z]+)([A-Z][a-z])/g, '$1 $2')
		.split(/\s+/)
		.filter(Boolean);

	return words
		.map((w) => {
			const lower = w.toLowerCase();
			if (ACRONYMS[lower]) return ACRONYMS[lower];
			return w.charAt(0).toUpperCase() + w.slice(1).toLowerCase();
		})
		.join(' ');
}

/**
 * Parses raw Go validator error string if any leaked through.
 */
function parseRawValidatorString(raw: string): string | null {
	const regex = /Field validation for '([^']+)' failed on the '([^']+)' tag/gi;
	const matches = [...raw.matchAll(regex)];

	if (matches.length === 0) return null;

	const messages = matches.map((m) => {
		const field = humanizeField(m[1]);
		const tag = m[2]?.toLowerCase();

		switch (tag) {
			case 'required':
				return `${field} is required`;
			case 'min':
				return `${field} is too short`;
			case 'max':
				return `${field} is too long`;
			case 'email':
				return `${field} must be a valid email address`;
			case 'url':
				return `${field} must be a valid URL`;
			case 'nohtml':
				return `${field} cannot contain HTML tags`;
			case 'oneof':
				return `${field} has an invalid option selected`;
			case 'len':
				return `${field} has an invalid length`;
			default:
				return `${field} is invalid`;
		}
	});

	return messages.join('; ');
}

/**
 * Converts any API error response into a clean, human-friendly message.
 */
export function parseApiError(err: unknown, fallback = 'An unexpected error occurred'): string {
	if (!err) return fallback;

	let msg = '';
	if (typeof err === 'string') {
		msg = err;
	} else if (err instanceof Error) {
		msg = err.message;
	} else if (typeof err === 'object') {
		const obj = err as Record<string, unknown>;
		if (typeof obj.error === 'string') {
			msg = obj.error;
		} else if (typeof obj.message === 'string') {
			msg = obj.message;
		}
	}

	if (!msg) return fallback;

	// 1. Check for raw Go validator error format
	if (msg.includes('Field validation for') || (msg.includes('Key:') && msg.includes('Error:'))) {
		const parsed = parseRawValidatorString(msg);
		if (parsed) return parsed;
	}

	const lower = msg.toLowerCase();

	// 2. Check for raw database constraints
	if (lower.includes('duplicate key') || lower.includes('unique constraint')) {
		if (lower.includes('email')) {
			return 'An account with this email already exists.';
		}
		if (lower.includes('slug') || lower.includes('subdomain')) {
			return 'This identifier or subdomain is already taken.';
		}
		if (lower.includes('name')) {
			return 'A record with this name already exists.';
		}
		return 'A record with this information already exists.';
	}

	if (lower.includes('record not found') || lower.includes('no rows in result set')) {
		return 'The requested resource was not found.';
	}

	if (lower.includes('violates foreign key constraint')) {
		return 'Referenced resource does not exist or is currently in use.';
	}

	if (
		lower.includes('connection refused') ||
		lower.includes('driver: bad connection') ||
		lower.includes('dial tcp') ||
		lower.includes('pq:') ||
		lower.includes('sql:') ||
		lower.includes('gorm:')
	) {
		return 'A database error occurred. Please try again later.';
	}

	// 3. SMTP errors
	if (lower.includes('daily user sending limit exceeded') || lower.includes('5.4.5')) {
		return 'Email sending limit exceeded on mail server.';
	}
	if (lower.includes('smtp') && (lower.includes('failed') || lower.includes('error'))) {
		return 'Email service encountered an error. Please try again later.';
	}

	// 4. Clean out common error prefixes
	let cleaned = msg.replace(/^Error:\s*/i, '').replace(/^Failed to\s+[a-z]+:\s*/i, '');
	if (cleaned.startsWith('pq: ')) {
		cleaned = cleaned.replace(/^pq:\s*/, '');
	}

	// 5. Internal pointers or memory addresses
	if (cleaned.includes('0x') || cleaned.includes('.go:') || cleaned.includes('goroutine')) {
		return 'An unexpected server error occurred. Please try again later.';
	}

	return cleaned.trim() || fallback;
}

/**
 * Truncates an error detail string nicely for preview in notification logs.
 */
export function truncateErrorDetail(detail?: string, maxLen = 140): string {
	if (!detail) return '';
	const trimmed = detail.trim();
	if (trimmed.length <= maxLen) return trimmed;
	return trimmed.slice(0, maxLen - 3) + '...';
}
