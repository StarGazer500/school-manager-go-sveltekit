import { redirect } from '@sveltejs/kit';
import { failFromApi, orNotFound } from '#lib/server/api.ts';
import { coursesApi } from '#lib/features/courses/api.server.ts';
import { enrollAction, enrollmentsApi, withdrawAction } from '#lib/features/enrollments/api.server.ts';
import { studentsApi } from '#lib/features/students/api.server.ts';
import { parseStudentForm } from '#lib/features/students/form.ts';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params }) => {
	const studentId = Number(params.id);
	const [student, enrollments, courses] = await Promise.all([
		orNotFound(studentsApi.get(studentId)),
		enrollmentsApi.list({ studentId }),
		coursesApi.list()
	]);

	// Only offer courses the student isn't already in.
	const enrolled = new Set(enrollments.map((e) => e.courseId));
	return { student, enrollments, availableCourses: courses.filter((c) => !enrolled.has(c.id)) };
};

export const actions: Actions = {
	update: async ({ params, request }) => {
		const input = parseStudentForm(await request.formData());
		try {
			await studentsApi.update(params.id, input);
		} catch (err) {
			return failFromApi(err, input);
		}
		return { saved: true };
	},

	delete: async ({ params }) => {
		try {
			await studentsApi.remove(params.id);
		} catch (err) {
			return failFromApi(err);
		}
		redirect(303, '/students');
	},

	enroll: async ({ request }) => enrollAction(await request.formData()),
	withdraw: async ({ request }) => withdrawAction(await request.formData())
};
