package http

import (
	"strconv"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

// ParametrosPaginacao — page (padrão 1), page_size (padrão 20, máximo 100),
// sort e order validados contra a allowlist da entidade (design.md §6.2).
type ParametrosPaginacao struct {
	Page     int
	PageSize int
	Sort     string
	Order    string
}

// ParsePaginacao nunca clampa nem ignora silenciosamente um valor fora do
// permitido — sempre 400 ErrParametro (G-08, G-10). sort/order nunca
// chegam à query SQL fora da allowlist recebida por parâmetro.
func ParsePaginacao(c *gin.Context, allowlistSort map[string]bool, sortPadrao, orderPadrao string) (ParametrosPaginacao, error) {
	page := 1
	if v := c.Query("page"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return ParametrosPaginacao{}, &domain.ErrParametro{Nome: "page"}
		}
		page = n
	}

	pageSize := 20
	if v := c.Query("page_size"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 100 {
			return ParametrosPaginacao{}, &domain.ErrParametro{Nome: "page_size"}
		}
		pageSize = n
	}

	sort := c.Query("sort")
	if sort == "" {
		sort = sortPadrao
	} else if !allowlistSort[sort] {
		return ParametrosPaginacao{}, &domain.ErrParametro{Nome: "sort"}
	}

	order := c.DefaultQuery("order", orderPadrao)
	if order != "asc" && order != "desc" {
		return ParametrosPaginacao{}, &domain.ErrParametro{Nome: "order"}
	}

	return ParametrosPaginacao{Page: page, PageSize: pageSize, Sort: sort, Order: order}, nil
}
