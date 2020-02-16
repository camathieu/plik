package middleware

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
)

// Paginate parse pagination requests
func Paginate(ctx *context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {

		defaultLimit := 20
		pagingQuery := &common.PagingQuery{Limit: &defaultLimit}

		header := req.Header.Get("X-PlikPaging")
		if header != "" {
			err := json.Unmarshal([]byte(header), &pagingQuery)
			if err != nil {
				ctx.InvalidParameter("paging header")
				return
			}
		} else {
			limitStr := req.URL.Query().Get("limit")
			if limitStr != "" {
				limit, err := strconv.Atoi(limitStr)
				if err != nil {
					ctx.InvalidParameter("limit", err)
				}
				pagingQuery.Limit = &limit
			}

			order := req.URL.Query().Get("order")
			switch order {
			case "":
			case "desc", "asc":
				pagingQuery.Order = &order
			default:
				ctx.InvalidParameter("order. Expected 'asc' or 'desc'")
				return
			}

			before := req.URL.Query().Get("before")
			if before != "" {
				pagingQuery.Before = &before
			}

			after := req.URL.Query().Get("after")
			if after != "" {
				pagingQuery.After = &after
			}
		}

		if pagingQuery.Limit == nil || *pagingQuery.Limit <= 0 {
			ctx.InvalidParameter("paging limit")
			return
		}

		if pagingQuery.After != nil && pagingQuery.Before != nil {
			ctx.BadRequest("bidirectional paging cursor")
			return
		}

		ctx.SetPagingQuery(pagingQuery)

		next.ServeHTTP(resp, req)
	})
}
