package common

import (
	paginator "github.com/pilagod/gorm-cursor-paginator"
	"github.com/root-gg/utils"
)

type PagingQuery struct {
	After  *string `json:"after"`
	Before *string `json:"before"`
	Limit  *int    `json:"limit"`
	Order  *string `json:"order"`
}

func (q *PagingQuery) Paginator() *paginator.Paginator {
	p := paginator.New()

	if q.After != nil {
		p.SetAfterCursor(*q.After) // [default: nil]
	}

	if q.Before != nil {
		p.SetBeforeCursor(*q.Before) // [default: nil]
	}

	if q.Limit != nil {
		p.SetLimit(*q.Limit) // [default: 10]
	}

	if q.Order != nil && *q.Order == "asc" {
		p.SetOrder(paginator.ASC) // [default: paginator.DESC]
	}

	return p
}

type PagingResponse struct {
	After   *string       `json:"after"`
	Before  *string       `json:"before"`
	Size    int           `json:"size"`
	Results []interface{} `json:"results"`
}

func NewPagingResponse(results interface{}, cursor *paginator.Cursor) (pr *PagingResponse) {
	pr = &PagingResponse{}
	pr.Results = utils.ToInterfaceArray(results)
	pr.Before = cursor.Before
	pr.After = cursor.After
	pr.Size = len(pr.Results)
	return pr
}
