import type { TeacherInput } from '#lib/services/teachers/types.ts';

export function parseTeacherForm(data: FormData): TeacherInput {
	return {
		firstName: String(data.get('firstName') ?? ''),
		lastName: String(data.get('lastName') ?? ''),
		email: String(data.get('email') ?? ''),
		subject: String(data.get('subject') ?? '')
	};
}
