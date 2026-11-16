// Grafana Faro: observability for the browser side of the app.
//
// What it sends (to /collect, which Vite in dev and nginx in production
// forward to the OpenTelemetry Collector, and from there to Loki and Tempo):
//   - errors: uncaught exceptions, rejected promises, console.error, and the
//     ones we report ourselves (see reportError)
//   - Web Vitals: how fast pages load and respond, as users experience it
//   - page views and the session, so errors can be grouped per visit
//   - traces: each fetch to /api gets a span, and its `traceparent` header
//     makes the Go API continue the same trace. One trace then runs from the
//     browser to the SQL query.
//
// Faro batches what it sends and never blocks the page; if the collector is
// down, telemetry is dropped and the app works as before.
import { faro, getWebInstrumentations, initializeFaro } from '@grafana/faro-web-sdk';
import { TracingInstrumentation } from '@grafana/faro-web-tracing';

let started = false;

/** Starts Faro once, in the browser (called from hooks.client.ts). */
export function initFaro(): void {
	if (started || typeof window === 'undefined') return;
	started = true;

	initializeFaro({
		// Same origin as the app, so no CORS. Absolute, so Faro recognises its
		// own requests and doesn't trace them.
		url: `${window.location.origin}/collect`,
		app: {
			name: 'school-web', // the service name in Grafana, next to school-api
			version: __APP_VERSION__, // the git commit, like the API's (vite.config.ts)
			environment: import.meta.env.MODE
		},
		instrumentations: [
			...getWebInstrumentations({ captureConsole: true }),
			new TracingInstrumentation()
		]
	});
}

/**
 * Reports an error we handled ourselves (so it never reached the browser's
 * error handlers), with extra context such as the API's trace ID. Does
 * nothing if Faro isn't running.
 */
export function reportError(error: Error, context: Record<string, string>): void {
	faro.api?.pushError(error, { context });
}
