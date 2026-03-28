package client

import "time"

// CreateFormParams holds parameters for creating a form via the API.
type CreateFormParams struct {
	Title     string  `json:"title"`
	Fields    []Field `json:"fields"`
	Recipient string  `json:"recipient"`
	ExpiresIn string  `json:"expires_in,omitempty"`
	Context   string  `json:"context,omitempty"`
}

// Field describes a single form field.
type Field struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Label    string   `json:"label"`
	Required bool     `json:"required"`
	Options  []string `json:"options,omitempty"`
}

// AskOptions are optional parameters for the Ask shorthand.
type AskOptions struct {
	Context string
	Expires string
}

// ListOptions holds query parameters for listing forms.
type ListOptions struct {
	Status string
	Limit  int
}

// Form is the API representation of a created form.
type Form struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	URL       string    `json:"url"`
	Recipient string    `json:"recipient"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// FormDetail includes full form information returned by GetForm.
type FormDetail struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Status    string     `json:"status"`
	URL       string     `json:"url"`
	Recipient string     `json:"recipient"`
	Fields    []Field    `json:"fields"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Response  *FormResponse `json:"response,omitempty"`
}

// FormResponse contains the data submitted via a form.
type FormResponse struct {
	FormID      string                 `json:"form_id"`
	Data        map[string]interface{} `json:"data"`
	RespondedAt time.Time              `json:"responded_at"`
}

// ListResult is the paginated list of forms.
type ListResult struct {
	Forms      []Form `json:"forms"`
	TotalCount int    `json:"total_count"`
	HasMore    bool   `json:"has_more"`
}

// APIError represents an error response from the API.
type APIError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"error"`
	Detail     string `json:"detail,omitempty"`
}

func (e *APIError) Error() string {
	if e.Detail != "" {
		return e.Message + ": " + e.Detail
	}
	return e.Message
}
