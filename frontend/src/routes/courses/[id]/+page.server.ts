import { redirect } from '@sveltejs/kit';
import { failFromApi, orNotFound } from '#lib/server/api.ts';
import { coursesApi } from '#lib/features/courses/api.server.ts';
import { parseCourseForm } from '#lib/features/courses/form.ts';
import { enrollAction, enrollmentsApi, withdrawAction } from '#lib/features/enrollments/api.server.ts';
import { studentsApi } from '#lib/features/students/api.server.ts';
import { teachersApi } from '#lib/features/teachers/api.server.ts';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params }) => {
	const courseId = Number(params.id);
	const [course, enrollments, teachers, students] = await Promise.all([
		orNotFound(coursesApi.get(courseId)),
		enrollmentsApi.list({ courseId }),
		teachersApi.list(),
		studentsApi.list()
	]);

	const enrolled = new Set(enrollments.map((e) => e.studentId));
	return {
		course,
		enrollments,
		teachers,
		availableStudents: students.filter((s) => !enrolled.has(s.id))
	};
};

export const actions: Actions = {
	update: async ({ params, request }) => {
		const input = parseCourseForm(await request.formData());
		try {
			await coursesApi.update(params.id, input);
		} catch (err) {
			return failFromApi(err, input);
		}
		return { saved: true };
	},

	delete: async ({ params }) => {
		try {
			await coursesApi.remove(params.id);
		} catch (err) {
			return failFromApi(err);
		}
		redirect(303, '/courses');
	},

	enroll: async ({ request }) => enrollAction(await request.formData()),
	withdraw: async ({ request }) => withdrawAction(await request.formData())
};
