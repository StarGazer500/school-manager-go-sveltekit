// Mirrors the Go DTOs in backend/internal/features/teachers/dto.go.

export interface Teacher {
	id: number;
	firstName: string;
	lastName: string;
	email: string;
	subject: string;
	createdAt: string;
	updatedAt: string;
}

export interface TeacherInput {
	firstName: string;
	lastName: string;
	email: string;
	subject: string;
}

export const fullName = (t: Pick<Teacher, 'firstName' | 'lastName'>) => `${t.firstName} ${t.lastName}`;
