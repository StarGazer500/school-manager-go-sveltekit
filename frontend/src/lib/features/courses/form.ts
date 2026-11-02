import type { CourseInput } from './types.ts';

export function parseCourseForm(data: FormData): CourseInput {
	const teacherId = Number(data.get('teacherId'));
	return {
		code: String(data.get('code') ?? ''),
		name: String(data.get('name') ?? ''),
		description: String(data.get('description') ?? ''),
		capacity: Number(data.get('capacity')),
		teacherId: teacherId > 0 ? teacherId : null
	};
}
