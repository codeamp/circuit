package codeamp_resolvers

type Paginator struct {
	Page        int32 `json:"page"`
	Count       int32 `json:"count"`
	HasNextPage bool  `json:"hasNextPage"`
}

type PaginatorResolver interface {
	Page() int32
	Count() int32
	HasNextPage() int32
}
