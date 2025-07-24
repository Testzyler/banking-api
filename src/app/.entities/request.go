package entities

type PaginationParams struct {
	PerPage int    `json:"per_page" query:"per_page" default:"10"`
	Page    int    `json:"page" query:"page" default:"1"`
	Search  string `json:"search" query:"search" default:""`
}
