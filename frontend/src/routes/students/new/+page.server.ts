import { redirect } from '@sveltejs/kit';
import { failFromApi } from '#lib/server/api.ts';
import { studentsApi } from '#lib/features/students/api.server.ts';
import { parseStudentForm } from '#lib/features/students/form.ts';
import type { Actions } from './$types';

export const actions: Actions = {
	default: async ({ request }) => {
		const input = parseStudentForm(await request.formData());

		let id: number;
		try {
			id = (await studentsApi.create(input)).id;
		} catch (err) {
			// Send the values back so the form isn't wiped on error.
			return failFromApi(err, input);
		}

		// redirect() works by throwing, so keep it outside the try/catch.
		redirect(303, `/students/${id}`);
	}
};
