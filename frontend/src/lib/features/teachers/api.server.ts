import { api, query } from '#lib/server/api.ts';
import type { Teacher, TeacherInput } from './types.ts';

export const teachersApi = {
	list: (params: { q?: string } = {}) => api<Teacher[]>(`/teachers${query(params)}`),
	get: (id: number | string) => api<Teacher>(`/teachers/${id}`),
	create: (input: TeacherInput) => api<Teacher>('/teachers', { method: 'POST', json: input }),
	update: (id: number | string, input: TeacherInput) =>
		api<Teacher>(`/teachers/${id}`, { method: 'PUT', json: input }),
	remove: (id: number | string) => api<void>(`/teachers/${id}`, { method: 'DELETE' })
};
