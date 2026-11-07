import { query, request, withMessage } from '../shared.ts';
import type { EnrollInput, Enrollment } from './types.ts';

export async function readEnrollments(filter: { studentId?: number; courseId?: number } = {}) {
	return request<Enrollment[]>(`/enrollments${query(filter)}`);
}

export async function enroll(input: EnrollInput) {
	return request<Enrollment>('/enrollments', { method: 'POST', json: input });
}

export async function withdraw(id: number) {
	return withMessage(await request<void>(`/enrollments/${id}`, { method: 'DELETE' }), 'Withdrawn');
}
