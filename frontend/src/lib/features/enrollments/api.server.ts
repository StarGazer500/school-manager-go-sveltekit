import { api, failFromApi, query } from '#lib/server/api.ts';
import type { EnrollInput, Enrollment } from './types.ts';

export const enrollmentsApi = {
	list: (params: { studentId?: number; courseId?: number } = {}) =>
		api<Enrollment[]>(`/enrollments${query(params)}`),
	create: (input: EnrollInput) => api<Enrollment>('/enrollments', { method: 'POST', json: input }),
	remove: (id: number | string) => api<void>(`/enrollments/${id}`, { method: 'DELETE' })
};

// Shared form action bodies. Both the student page and the course page let
// you enroll/withdraw, so the logic lives here once and each route's
// `actions` just delegates to these.

export async function enrollAction(data: FormData) {
	try {
		await enrollmentsApi.create({
			studentId: Number(data.get('studentId')),
			courseId: Number(data.get('courseId'))
		});
	} catch (err) {
		return failFromApi(err);
	}
}

export async function withdrawAction(data: FormData) {
	try {
		await enrollmentsApi.remove(String(data.get('enrollmentId')));
	} catch (err) {
		return failFromApi(err);
	}
}
