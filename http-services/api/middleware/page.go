package middleware

import (
	"strconv"

	"http-services/config"

	"github.com/gin-gonic/gin"
)

type PageQuery struct {
	Page     int
	PageSize int
	Offset   int
	Limit    int
	Disabled bool
}

func (p PageQuery) IsDisabled() bool {
	return p.Disabled
}

func ParsePageQuery(c *gin.Context) PageQuery {
	page := parsePositiveInt(c.Query("page"), config.DefaultPage)
	pageSize := parsePositiveInt(c.Query("page_size"), config.DefaultPageSize)
	if c.Query("page") == strconv.Itoa(config.CancelPage) || c.Query("page_size") == strconv.Itoa(config.CancelPageSize) {
		return PageQuery{
			Page:     config.CancelPage,
			PageSize: config.CancelPageSize,
			Offset:   0,
			Limit:    0,
			Disabled: true,
		}
	}

	return PageQuery{
		Page:     page,
		PageSize: pageSize,
		Offset:   (page - 1) * pageSize,
		Limit:    pageSize,
		Disabled: false,
	}
}

func GetPage(c *gin.Context) int {
	return ParsePageQuery(c).Page
}

func GetPageSize(c *gin.Context) int {
	return ParsePageQuery(c).PageSize
}

func parsePositiveInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	if parsed <= 0 {
		return fallback
	}
	return parsed
}
