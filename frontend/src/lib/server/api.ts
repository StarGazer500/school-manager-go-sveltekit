// Server-only HTTP client for the Go API. Because it lives under /server/,
// SvelteKit refuses to bundle it into browser code, so API_URL never leaks
// and the browser never talks to the backend directly (no CORS needed).
import { error, fail } from '@sveltejs/kit';
import { API_URL } from '$app/env/private';

export class ApiError extends Error {
	status: number;

	constructor(status: number, message: string) {
		super(message);
		this.status = status;
	}
}

type RequestOptions = Omit<RequestInit, 'body'> & { json?: unknown };

export async function api<T>(path: string, { json, ...init }: RequestOptions = {}): Promise<T> {
	const res = await fetch(`${API_URL}${path}`, {
		...init,
		headers: json === undefined ? init.headers : { 'content-type': 'application/json', ...init.headers },
		body: json === undefined ? undefined : JSON.stringify(json)
	});

	if (!res.ok) {
		const body = await res.json().catch(() => null);
		throw new ApiError(res.status, body?.error ?? res.statusText);
	}
	if (res.status === 204) return undefined as T;
	return res.json();
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

/** For load functions: a 404 from the API becomes SvelteKit's 404 page. */
export async function orNotFound<T>(request: Promise<T>): Promise<T> {
	try {
		return await request;
	} catch (err) {
		if (err instanceof ApiError && err.status === 404) error(404, err.message);
		throw err;
	}
}

/**
 * Turns a 4xx ApiError into a form action failure so the page can show the
 * message and re-fill the form with `values`. Anything else (5xx, backend
 * down) is re-thrown and ends up on the +error page.
 */
export function failFromApi<V>(err: unknown, values?: V) {
	if (err instanceof ApiError && err.status < 500) {
		return fail(err.status, { error: err.message, values });
	}
	throw err;
}
