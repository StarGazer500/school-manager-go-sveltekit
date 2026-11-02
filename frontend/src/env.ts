import { defineEnvVars } from '@sveltejs/kit/env';

// Environment variables are declared here and imported from
// `$app/env/private` (server-only) or `$app/env/public`.
export const variables = defineEnvVars({
	API_URL: {
		description: 'Base URL of the Go API, including the /api/v1 prefix',
		schema: (value) => value || 'http://localhost:8080/api/v1'
	}
});
