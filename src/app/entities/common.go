package entities

import "github.com/Testzyler/banking-api/server/response"

type PaginationParams struct {
	PerPage int    `json:"per_page" query:"per_page" default:"10" validate:"min=1,max=100"`
	Page    int    `json:"page" query:"page" default:"1" validate:"min=1"`
	Search  string `json:"search" query:"search" default:"" validate:"max=255"`
}

type PaginationMeta struct {
	response.PaginationMeta
}
