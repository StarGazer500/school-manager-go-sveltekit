import { redirect } from '@sveltejs/kit';
import { failFromApi } from '#lib/server/api.ts';
import { teachersApi } from '#lib/features/teachers/api.server.ts';
import { parseTeacherForm } from '#lib/features/teachers/form.ts';
import type { Actions } from './$types';

export const actions: Actions = {
	default: async ({ request }) => {
		const input = parseTeacherForm(await request.formData());

		let id: number;
		try {
			id = (await teachersApi.create(input)).id;
		} catch (err) {
			return failFromApi(err, input);
		}
		redirect(303, `/teachers/${id}`);
	}
};
