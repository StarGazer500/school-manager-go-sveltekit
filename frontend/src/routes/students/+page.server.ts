import { failFromApi } from '#lib/server/api.ts';
import { studentsApi } from '#lib/features/students/api.server.ts';
import type { Actions, PageServerLoad } from './$types';

// Runs on the server for every visit. `url` lets us read ?q= from the search box.
export const load: PageServerLoad = async ({ url }) => {
	const q = url.searchParams.get('q') ?? '';
	return { q, students: await studentsApi.list({ q }) };
};

// Form actions: <form method="POST" action="?/delete"> calls `delete` below.
// After it returns, SvelteKit re-runs `load` so the list is fresh.
export const actions: Actions = {
	delete: async ({ request }) => {
		const data = await request.formData();
		try {
			await studentsApi.remove(String(data.get('id')));
		} catch (err) {
			return failFromApi(err);
		}
	}
};
