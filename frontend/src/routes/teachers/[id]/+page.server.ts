import { redirect } from '@sveltejs/kit';
import { failFromApi, orNotFound } from '#lib/server/api.ts';
import { coursesApi } from '#lib/features/courses/api.server.ts';
import { teachersApi } from '#lib/features/teachers/api.server.ts';
import { parseTeacherForm } from '#lib/features/teachers/form.ts';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params }) => {
	const teacherId = Number(params.id);
	const [teacher, courses] = await Promise.all([
		orNotFound(teachersApi.get(teacherId)),
		coursesApi.list({ teacherId })
	]);
	return { teacher, courses };
};

export const actions: Actions = {
	update: async ({ params, request }) => {
		const input = parseTeacherForm(await request.formData());
		try {
			await teachersApi.update(params.id, input);
		} catch (err) {
			return failFromApi(err, input);
		}
		return { saved: true };
	},

	delete: async ({ params }) => {
		try {
			await teachersApi.remove(params.id);
		} catch (err) {
			return failFromApi(err);
		}
		redirect(303, '/teachers');
	}
};
