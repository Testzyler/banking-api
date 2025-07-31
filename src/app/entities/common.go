package entities

import (
	"github.com/Testzyler/banking-api/app/validators"
	"github.com/Testzyler/banking-api/server/response"
)

type PaginationParams struct {
	PerPage int    `json:"perPage" query:"perPage" validate:"required,min=1,max=100"`
	Page    int    `json:"page" query:"page" validate:"required,min=1"`
	Search  string `json:"search" query:"search" validate:"max=255"`
}

func (p *PaginationParams) Validate() error {
	return validators.ValidateStruct(p)
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
