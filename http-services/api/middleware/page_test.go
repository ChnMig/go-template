package middleware

import (
	"net/http/httptest"
	"testing"

	"http-services/config"

	"github.com/gin-gonic/gin"
)

func TestParsePageQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		url  string
		want PageQuery
	}{
		{
			name: "默认分页",
			url:  "/test",
			want: PageQuery{Page: config.DefaultPage, PageSize: config.DefaultPageSize, Offset: 0, Limit: config.DefaultPageSize},
		},
		{
			name: "正常分页",
			url:  "/test?page=3&page_size=15",
			want: PageQuery{Page: 3, PageSize: 15, Offset: 30, Limit: 15},
		},
		{
			name: "非法页码回落默认值",
			url:  "/test?page=0&page_size=abc",
			want: PageQuery{Page: config.DefaultPage, PageSize: config.DefaultPageSize, Offset: 0, Limit: config.DefaultPageSize},
		},
		{
			name: "page 取消分页",
			url:  "/test?page=-1&page_size=20",
			want: PageQuery{Page: config.CancelPage, PageSize: config.CancelPageSize, Offset: 0, Limit: 0, Disabled: true},
		},
		{
			name: "page_size 取消分页",
			url:  "/test?page=1&page_size=-1",
			want: PageQuery{Page: config.CancelPage, PageSize: config.CancelPageSize, Offset: 0, Limit: 0, Disabled: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", tt.url, nil)

			got := ParsePageQuery(c)
			if got != tt.want {
				t.Fatalf("ParsePageQuery() = %+v, want %+v", got, tt.want)
			}
			if got.IsDisabled() != tt.want.Disabled {
				t.Fatalf("IsDisabled() = %v, want %v", got.IsDisabled(), tt.want.Disabled)
			}
			if GetPage(c) != tt.want.Page {
				t.Fatalf("GetPage() = %d, want %d", GetPage(c), tt.want.Page)
			}
			if GetPageSize(c) != tt.want.PageSize {
				t.Fatalf("GetPageSize() = %d, want %d", GetPageSize(c), tt.want.PageSize)
			}
		})
	}
}
