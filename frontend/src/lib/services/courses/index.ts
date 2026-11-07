import { query, request, withMessage } from '../shared.ts';
import type { Course, CourseInput } from './types.ts';

export async function readCourses(filter: { search?: string; teacherId?: number } = {}) {
	return request<Course[]>(`/courses${query({ q: filter.search, teacherId: filter.teacherId })}`);
}

export async function readCourse(id: number) {
	return request<Course>(`/courses/${id}`);
}

export async function createCourse(input: CourseInput) {
	return request<Course>('/courses', { method: 'POST', json: input });
}

export async function updateCourse(id: number, input: CourseInput) {
	return request<Course>(`/courses/${id}`, { method: 'PUT', json: input });
}

export async function deleteCourse(id: number) {
	return withMessage(await request<void>(`/courses/${id}`, { method: 'DELETE' }), 'Course deleted');
}
