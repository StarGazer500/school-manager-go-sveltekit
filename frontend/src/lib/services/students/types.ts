// Mirrors the Go DTOs in backend/internal/features/students/dto.go.

export interface Student {
	id: number;
	firstName: string;
	lastName: string;
	email: string;
	gradeLevel: number;
	createdAt: string;
	updatedAt: string;
}

export interface StudentInput {
	firstName: string;
	lastName: string;
	email: string;
	gradeLevel: number;
}

export const fullName = (s: Pick<Student, 'firstName' | 'lastName'>) => `${s.firstName} ${s.lastName}`;
