package courses

import "time"

// CourseInput is the JSON body for create and update.
type CourseInput struct {
	Code        string `json:"code"        binding:"required,max=20"`
	Name        string `json:"name"        binding:"required,max=150"`
	Description string `json:"description" binding:"max=1000"`
	Capacity    int    `json:"capacity"    binding:"required,min=1,max=500"`
	TeacherID   *uint  `json:"teacherId"`
}

type CourseResponse struct {
	ID          uint      `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Capacity    int       `json:"capacity"`
	TeacherID   *uint     `json:"teacherId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func toResponse(c Course) CourseResponse {
	return CourseResponse{
		ID:          c.ID,
		Code:        c.Code,
		Name:        c.Name,
		Description: c.Description,
		Capacity:    c.Capacity,
		TeacherID:   c.TeacherID,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func toResponses(list []Course) []CourseResponse {
	out := make([]CourseResponse, len(list))
	for i, c := range list {
		out[i] = toResponse(c)
	}
	return out
}
