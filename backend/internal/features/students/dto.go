package students

import "time"

// StudentInput is the JSON body for create and update. The binding tags are
// validated by Gin's ShouldBindJSON before the service is called.
type StudentInput struct {
	FirstName  string `json:"firstName"  binding:"required,max=100"`
	LastName   string `json:"lastName"   binding:"required,max=100"`
	Email      string `json:"email"      binding:"required,email"`
	GradeLevel int    `json:"gradeLevel" binding:"required,min=1,max=12"`
}

// StudentResponse is what clients see. Keeping it separate from the entity
// means the database shape can change without breaking the API.
type StudentResponse struct {
	ID         int64     `json:"id"`
	FirstName  string    `json:"firstName"`
	LastName   string    `json:"lastName"`
	Email      string    `json:"email"`
	GradeLevel int       `json:"gradeLevel"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func toResponse(s Student) StudentResponse {
	return StudentResponse{
		ID:         s.ID,
		FirstName:  s.FirstName,
		LastName:   s.LastName,
		Email:      s.Email,
		GradeLevel: s.GradeLevel,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}
}

func toResponses(list []Student) []StudentResponse {
	out := make([]StudentResponse, len(list))
	for i, s := range list {
		out[i] = toResponse(s)
	}
	return out
}
