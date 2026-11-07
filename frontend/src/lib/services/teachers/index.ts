import { query, request, withMessage } from '../shared.ts';
import type { Teacher, TeacherInput } from './types.ts';

export async function readTeachers(search = '') {
	return request<Teacher[]>(`/teachers${query({ q: search })}`);
}

export async function readTeacher(id: number) {
	return request<Teacher>(`/teachers/${id}`);
}

export async function createTeacher(input: TeacherInput) {
	return request<Teacher>('/teachers', { method: 'POST', json: input });
}

export async function updateTeacher(id: number, input: TeacherInput) {
	return request<Teacher>(`/teachers/${id}`, { method: 'PUT', json: input });
}

export async function deleteTeacher(id: number) {
	return withMessage(await request<void>(`/teachers/${id}`, { method: 'DELETE' }), 'Teacher deleted');
}
