package db_resolver

import (
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

type Paginator struct {
	Page        int32 `json:"page"`
	Count       int32 `json:"count"`
	HasNextPage bool  `json:"hasNextPage"`
}

type PaginatorResolver interface {
	Page() int32
	Count() int32
	Entries() []interface{}
	HasNextPage() int32
}

// Releases
func (r *ReleaseListResolver) Entries() []*ReleaseResolver {
	var filteredRows []model.Release
	var results []*ReleaseResolver

	cursorRowIdx := 0

	// filter on things after cursor_id
	for idx, row := range r.ReleaseList {
		if r.PaginatorInput.Cursor != nil && row.Model.ID.String() == *r.PaginatorInput.Cursor {
			cursorRowIdx = idx
			break
		}
	}

	i := cursorRowIdx + 1
	for {
		if len(filteredRows) == int(r.PaginatorInput.Limit) ||
			len(r.ReleaseList) == i {
			break
		}
		filteredRows = append(filteredRows, r.ReleaseList[i])
		i++
	}

	for _, row := range filteredRows {
		results = append(results, &ReleaseResolver{
			DB:      r.DB,
			Release: row,
		})
	}
	return results
}

func (r *ReleaseListResolver) Page() int32 {
	// get page # from count / itemsPerPage
	return r.getPage()
}

func (r *ReleaseListResolver) Count() int32 {
	return int32(len(r.ReleaseList))
}

func (r *ReleaseListResolver) NextCursor() string {
	cursorRowIdx := 0

	// filter on things after cursor_id
	for idx, row := range r.ReleaseList {
		if r.PaginatorInput.Cursor != nil && row.Model.ID.String() == *r.PaginatorInput.Cursor {
			cursorRowIdx = idx
			break
		}
	}

	nextCursorIdx := cursorRowIdx + int(r.PaginatorInput.Limit) + 1
	if len(r.ReleaseList) >= nextCursorIdx {
		return r.ReleaseList[nextCursorIdx].Model.ID.String()
	} else {
		return ""
	}
}

func (r *ReleaseListResolver) getPage() int32 {
	for idx, row := range r.ReleaseList {
		if row.Model.ID.String() == *r.PaginatorInput.Cursor {
			return int32(idx)/r.PaginatorInput.Limit + int32(1)
		}
	}

	return 1
}

// SECRETS

// SecretResolver resolver for Release
type SecretListResolver struct {
	SecretList     []model.Secret
	PaginatorInput *model.PaginatorInput
	DB             *gorm.DB
}

// Secrets
func (r *SecretListResolver) Entries() []*SecretResolver {
	var filteredRows []model.Secret
	var results []*SecretResolver

	cursorRowIdx := 0

	// filter on things after cursor_id
	for idx, row := range r.SecretList {
		if r.PaginatorInput.Cursor != nil && row.Model.ID.String() == *r.PaginatorInput.Cursor {
			cursorRowIdx = idx
			break
		}
	}

	i := cursorRowIdx + 1
	for {
		if len(filteredRows) == int(r.PaginatorInput.Limit) ||
			len(r.SecretList) == i {
			break
		}
		filteredRows = append(filteredRows, r.SecretList[i])
		i++
	}

	for _, row := range filteredRows {
		var secretValue model.SecretValue

		r.DB.Where("secret_id = ?", row.Model.ID).Order("created_at desc").First(&secretValue)

		results = append(results, &SecretResolver{
			DB:          r.DB,
			Secret:      row,
			SecretValue: secretValue,
		})
	}
	return results
}

func (r *SecretListResolver) Page() int32 {
	// get page # from count / itemsPerPage
	return r.getPage()
}

func (r *SecretListResolver) Count() int32 {
	return int32(len(r.SecretList))
}

func (r *SecretListResolver) NextCursor() string {
	cursorRowIdx := 0

	// filter on things after cursor_id
	for idx, row := range r.SecretList {
		if r.PaginatorInput.Cursor != nil && row.Model.ID.String() == *r.PaginatorInput.Cursor {
			cursorRowIdx = idx
			break
		}
	}

	nextCursorIdx := cursorRowIdx + int(r.PaginatorInput.Limit) + 1
	if len(r.SecretList) >= nextCursorIdx {
		return r.SecretList[nextCursorIdx].Model.ID.String()
	} else {
		return ""
	}
}

func (r *SecretListResolver) getPage() int32 {
	for idx, row := range r.SecretList {
		if row.Model.ID.String() == *r.PaginatorInput.Cursor {
			return int32(idx)/r.PaginatorInput.Limit + int32(1)
		}
	}

	return 1
}
