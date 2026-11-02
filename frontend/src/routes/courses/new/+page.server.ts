import { redirect } from '@sveltejs/kit';
import { failFromApi } from '#lib/server/api.ts';
import { coursesApi } from '#lib/features/courses/api.server.ts';
import { parseCourseForm } from '#lib/features/courses/form.ts';
import { teachersApi } from '#lib/features/teachers/api.server.ts';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async () => ({ teachers: await teachersApi.list() });

export const actions: Actions = {
	default: async ({ request }) => {
		const input = parseCourseForm(await request.formData());

		let id: number;
		try {
			id = (await coursesApi.create(input)).id;
		} catch (err) {
			return failFromApi(err, input);
		}
		redirect(303, `/courses/${id}`);
	}
};
