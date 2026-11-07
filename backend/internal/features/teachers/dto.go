package teachers

import "time"

// TeacherInput is the JSON body for create and update.
type TeacherInput struct {
	FirstName string `json:"firstName" binding:"required,max=100"`
	LastName  string `json:"lastName"  binding:"required,max=100"`
	Email     string `json:"email"     binding:"required,email"`
	Subject   string `json:"subject"   binding:"required,max=100"`
}

type TeacherResponse struct {
	ID        int64     `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	Subject   string    `json:"subject"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func toResponse(t Teacher) TeacherResponse {
	return TeacherResponse{
		ID:        t.ID,
		FirstName: t.FirstName,
		LastName:  t.LastName,
		Email:     t.Email,
		Subject:   t.Subject,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func toResponses(list []Teacher) []TeacherResponse {
	out := make([]TeacherResponse, len(list))
	for i, t := range list {
		out[i] = toResponse(t)
	}
	return out
}
