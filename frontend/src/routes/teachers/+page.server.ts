import { failFromApi } from '#lib/server/api.ts';
import { teachersApi } from '#lib/features/teachers/api.server.ts';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ url }) => {
	const q = url.searchParams.get('q') ?? '';
	return { q, teachers: await teachersApi.list({ q }) };
};

export const actions: Actions = {
	delete: async ({ request }) => {
		const data = await request.formData();
		try {
			await teachersApi.remove(String(data.get('id')));
		} catch (err) {
			return failFromApi(err);
		}
	}
};
