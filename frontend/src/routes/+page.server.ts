import { coursesApi } from '#lib/features/courses/api.server.ts';
import { enrollmentsApi } from '#lib/features/enrollments/api.server.ts';
import { studentsApi } from '#lib/features/students/api.server.ts';
import { teachersApi } from '#lib/features/teachers/api.server.ts';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	// Fire all four requests at once rather than one after another.
	const [students, teachers, courses, enrollments] = await Promise.all([
		studentsApi.list(),
		teachersApi.list(),
		coursesApi.list(),
		enrollmentsApi.list()
	]);

	return {
		counts: {
			students: students.length,
			teachers: teachers.length,
			courses: courses.length,
			enrollments: enrollments.length
		},
		recentEnrollments: enrollments.slice(0, 5)
	};
};
