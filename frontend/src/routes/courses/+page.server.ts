import { failFromApi } from '#lib/server/api.ts';
import { coursesApi } from '#lib/features/courses/api.server.ts';
import { teachersApi } from '#lib/features/teachers/api.server.ts';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ url }) => {
	const q = url.searchParams.get('q') ?? '';
	// Teachers are loaded too so the table can show names, not just IDs.
	const [courses, teachers] = await Promise.all([coursesApi.list({ q }), teachersApi.list()]);
	return { q, courses, teachers };
};

export const actions: Actions = {
	delete: async ({ request }) => {
		const data = await request.formData();
		try {
			await coursesApi.remove(String(data.get('id')));
		} catch (err) {
			return failFromApi(err);
		}
	}
};
