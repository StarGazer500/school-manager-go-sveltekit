import type { TeacherInput } from './types.ts';

export function parseTeacherForm(data: FormData): TeacherInput {
	return {
		firstName: String(data.get('firstName') ?? ''),
		lastName: String(data.get('lastName') ?? ''),
		email: String(data.get('email') ?? ''),
		subject: String(data.get('subject') ?? '')
	};
}
