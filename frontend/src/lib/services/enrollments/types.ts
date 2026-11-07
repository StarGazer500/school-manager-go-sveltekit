// Mirrors the Go DTOs in backend/internal/features/enrollments/dto.go.

export interface Enrollment {
	id: number;
	studentId: number;
	studentName: string;
	courseId: number;
	courseCode: string;
	courseName: string;
	enrolledAt: string;
}

export interface EnrollInput {
	studentId: number;
	courseId: number;
}
