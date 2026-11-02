import { api, query } from '#lib/server/api.ts';
import type { Course, CourseInput } from './types.ts';

export const coursesApi = {
	list: (params: { q?: string; teacherId?: number } = {}) => api<Course[]>(`/courses${query(params)}`),
	get: (id: number | string) => api<Course>(`/courses/${id}`),
	create: (input: CourseInput) => api<Course>('/courses', { method: 'POST', json: input }),
	update: (id: number | string, input: CourseInput) =>
		api<Course>(`/courses/${id}`, { method: 'PUT', json: input }),
	remove: (id: number | string) => api<void>(`/courses/${id}`, { method: 'DELETE' })
};
