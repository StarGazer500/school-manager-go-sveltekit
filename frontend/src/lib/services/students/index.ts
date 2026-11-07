import { query, request, withMessage } from '../shared.ts';
import type { Student, StudentInput } from './types.ts';

export async function readStudents(search = '') {
	return request<Student[]>(`/students${query({ q: search })}`);
}

export async function readStudent(id: number) {
	return request<Student>(`/students/${id}`);
}

export async function createStudent(input: StudentInput) {
	return request<Student>('/students', { method: 'POST', json: input });
}

export async function updateStudent(id: number, input: StudentInput) {
	return request<Student>(`/students/${id}`, { method: 'PUT', json: input });
}

export async function deleteStudent(id: number) {
	return withMessage(await request<void>(`/students/${id}`, { method: 'DELETE' }), 'Student deleted');
}
