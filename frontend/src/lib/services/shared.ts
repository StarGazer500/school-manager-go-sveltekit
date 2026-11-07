// Every service function returns one of these instead of throwing, the same
// idea as MyApp's services/shared.ts. Components check `success` and show
// `message`, so no component needs a try/catch.
export interface IOkResult<T> {
	success: true;
	message: string;
	data: T;
}

export interface IFailureResult {
	success: false;
	message: string;
	/** HTTP status from the API, e.g. 404. Missing when the API could not be reached. */
	status?: number;
}

export type CallResult<T> = IOkResult<T> | IFailureResult;

// The app is a static SPA, so these requests come from the browser. They go to
// the same origin under /api/v1: in dev the Vite proxy forwards them to Go
// (vite.config.ts), in production nginx does (nginx.conf). No CORS needed.
const BASE = '/api/v1';

type RequestOptions = Omit<RequestInit, 'body'> & { json?: unknown };

export async function request<T>(path: string, { json, ...init }: RequestOptions = {}): Promise<CallResult<T>> {
	try {
		const res = await fetch(`${BASE}${path}`, {
			...init,
			headers: json === undefined ? init.headers : { 'content-type': 'application/json', ...init.headers },
			body: json === undefined ? undefined : JSON.stringify(json)
		});

		if (!res.ok) {
			// Go sends { "error": "..." }. Its 4xx messages are meant for users; 5xx ones are not.
			const body = await res.json().catch(() => null);
			return { success: false, message: failureMessage(res.status, body?.error), status: res.status };
		}

		const data = res.status === 204 ? undefined : await res.json();
		return { success: true, message: '', data: data as T };
	} catch {
		return { success: false, message: UNREACHABLE };
	}
}

const UNREACHABLE = 'Could not reach the API. Is the Go backend running?';

function failureMessage(status: number, apiMessage?: string): string {
	if (status < 500) return apiMessage ?? `Request failed (${status})`;
	// The Vite proxy and nginx answer 502/503/504 themselves when Go is down.
	if (status >= 502 && status <= 504) return UNREACHABLE;
	return 'The server had a problem. Try again.';
}

/** Builds "?a=1&b=2", skipping empty values. */
export function query(params: Record<string, string | number | undefined | null>): string {
	const search = new URLSearchParams();
	for (const [key, value] of Object.entries(params)) {
		if (value !== undefined && value !== null && value !== '') search.set(key, String(value));
	}
	const s = search.toString();
	return s ? `?${s}` : '';
}

/** Adds a success message to a delete result, like MyApp's deleteProduct. */
export function withMessage<T>(ret: CallResult<T>, message: string): CallResult<T> {
	return ret.success ? { ...ret, message } : ret;
}
