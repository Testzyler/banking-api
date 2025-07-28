package entities

import (
	"github.com/Testzyler/banking-api/server/exception"
	"github.com/Testzyler/banking-api/server/response"
)

type PaginationParams struct {
	PerPage int    `json:"per_page" query:"per_page" default:"10" validate:"min=1,max=100"`
	Page    int    `json:"page" query:"page" default:"1" validate:"min=1"`
	Search  string `json:"search" query:"search" default:"" validate:"max=255"`
}

func (p *PaginationParams) Validate() error {
	if p.PerPage < 1 || p.PerPage > 100 {
		return exception.NewValidationError("Invalid perPage value")
	}

	if p.Page < 1 {
		return exception.NewValidationError("Invalid page value")
	}

	if len(p.Search) > 255 {
		return exception.NewValidationError("Search term too long")
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
