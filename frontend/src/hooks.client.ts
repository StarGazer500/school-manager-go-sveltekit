// Client hooks: SvelteKit calls these in the browser.
import type { ClientInit, HandleClientError } from '@sveltejs/kit/hooks';
import { initFaro, reportError } from '#lib/observability/faro.ts';

// Runs once when the app starts: begin collecting browser telemetry.
export const init: ClientInit = () => {
	initFaro();
};

// Errors thrown while loading or rendering a page. SvelteKit catches them to
// show the error page, so the browser's own error handlers (and Faro's) never
// see them; report them here. Only `unknown` ones are bugs: `app` errors are
// thrown on purpose with error(), and `framework` ones are things like a 404.
// Returning nothing keeps SvelteKit's default status and message.
export const handleError: HandleClientError = ({ kind, error, event }) => {
	if (kind !== 'unknown') return;
	const err = error instanceof Error ? error : new Error(String(error));
	reportError(err, { page: event.url.pathname });
};
