// Mirrors the Go DTOs in backend/internal/features/courses/dto.go.

export interface Course {
	id: number;
	code: string;
	name: string;
	description: string;
	capacity: number;
	teacherId: number | null;
	createdAt: string;
	updatedAt: string;
}

export interface CourseInput {
	code: string;
	name: string;
	description: string;
	capacity: number;
	teacherId: number | null;
}
