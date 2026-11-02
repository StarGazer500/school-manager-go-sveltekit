import type { StudentInput } from './types.ts';

/** Reads a submitted StudentForm. Validation itself happens in the Go API. */
export function parseStudentForm(data: FormData): StudentInput {
	return {
		firstName: String(data.get('firstName') ?? ''),
		lastName: String(data.get('lastName') ?? ''),
		email: String(data.get('email') ?? ''),
		gradeLevel: Number(data.get('gradeLevel'))
	};
}
