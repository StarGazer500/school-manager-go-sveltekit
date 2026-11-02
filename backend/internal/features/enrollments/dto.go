package enrollments

import "time"

type EnrollInput struct {
	StudentID uint `json:"studentId" binding:"required"`
	CourseID  uint `json:"courseId"  binding:"required"`
}

type EnrollmentResponse struct {
	ID          uint      `json:"id"`
	StudentID   uint      `json:"studentId"`
	StudentName string    `json:"studentName"`
	CourseID    uint      `json:"courseId"`
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
