import tailwindcss from '@tailwindcss/vite';
import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig, loadEnv } from 'vite';

export default defineConfig(({ mode }) => {
	// Where the Go API runs. The browser calls /api/v1 on the same origin and
	// Vite forwards it here, so there is no CORS. In production nginx does the
	// same job (see nginx.conf). Override with API_TARGET=http://host:port.
	const apiTarget = loadEnv(mode, '.', '').API_TARGET || 'http://localhost:8080';
	const proxy = { '/api': apiTarget };

	return {
		server: { proxy },
		preview: { proxy },
		plugins: [
			tailwindcss(),
			sveltekit({
				compilerOptions: {
					// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
					runes: ({ filename }) =>
						filename.split(/[/\\]/).includes('node_modules') ? undefined : true
				},

				// adapter-static writes plain HTML, JS and CSS to build/, ready for nginx or a CDN.
				// The app is a single page app (ssr = false in src/routes/+layout.ts), so every
				// URL is served the same 200.html shell and the router takes it from there.
				adapter: adapter({ fallback: '200.html' })
			})
		]
	};
});
