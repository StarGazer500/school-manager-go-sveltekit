import { api, query } from '#lib/server/api.ts';
import type { Student, StudentInput } from './types.ts';

export const studentsApi = {
	list: (params: { q?: string; grade?: number } = {}) => api<Student[]>(`/students${query(params)}`),
	get: (id: number | string) => api<Student>(`/students/${id}`),
	create: (input: StudentInput) => api<Student>('/students', { method: 'POST', json: input }),
	update: (id: number | string, input: StudentInput) =>
		api<Student>(`/students/${id}`, { method: 'PUT', json: input }),
	remove: (id: number | string) => api<void>(`/students/${id}`, { method: 'DELETE' })
};
