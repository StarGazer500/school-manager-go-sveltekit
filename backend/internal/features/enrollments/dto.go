package enrollments

import "time"

type EnrollInput struct {
	StudentID int64 `json:"studentId" binding:"required,min=1"`
	CourseID  int64 `json:"courseId"  binding:"required,min=1"`
}

type EnrollmentResponse struct {
	ID          int64     `json:"id"`
	StudentID   int64     `json:"studentId"`
	StudentName string    `json:"studentName"`
	CourseID    int64     `json:"courseId"`
	CourseCode  string    `json:"courseCode"`
	CourseName  string    `json:"courseName"`
	EnrolledAt  time.Time `json:"enrolledAt"`
}

func toResponse(d Detail) EnrollmentResponse {
	return EnrollmentResponse(d)
}

func toResponses(list []Detail) []EnrollmentResponse {
	out := make([]EnrollmentResponse, len(list))
	for i, d := range list {
		out[i] = toResponse(d)
	}
	return out
}
