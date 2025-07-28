package entities

import "github.com/Testzyler/banking-api/server/response"

type PaginationParams struct {
	PerPage int    `json:"per_page" query:"per_page" default:"10" validate:"min=1,max=100"`
	Page    int    `json:"page" query:"page" default:"1" validate:"min=1"`
	Search  string `json:"search" query:"search" default:"" validate:"max=255"`
}

func (p *PaginationParams) Validate() error {
	errHttpStatus := 400
	ErrCode := response.ErrCodeValidationFailed

	if p.PerPage < 1 || p.PerPage > 100 {
		return &response.ErrorResponse{
			HttpStatusCode: errHttpStatus,
			Code:           ErrCode,
			Message:        "Invalid per_page value",
			Details:        "per_page must be between 1 and 100",
		}
	}

	if p.Page < 1 {
		return &response.ErrorResponse{
			HttpStatusCode: errHttpStatus,
			Code:           ErrCode,
			Message:        "Invalid page value",
			Details:        "page must be at least 1",
		}
	}

	if len(p.Search) > 255 {
		return &response.ErrorResponse{
			HttpStatusCode: errHttpStatus,
			Code:           ErrCode,
			Message:        "Search term too long",
			Details:        "search must not exceed 255 characters",
		}
	}
	return nil
}

type PaginatedResponse struct {
	response.SuccessResponse
	Meta PaginationMeta `json:"meta"`
}

type PaginationMeta struct {
	Page        int  `json:"page"`
	PerPage     int  `json:"perPage"`
	Total       int  `json:"total"`
	TotalPages  int  `json:"totalPages"`
	HasNext     bool `json:"hasNext"`
	HasPrevious bool `json:"hasPrevious"`
}
