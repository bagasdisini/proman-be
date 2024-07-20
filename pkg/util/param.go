package util

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"time"
)

type CommonQuery struct {
	Q     string    `query:"q"`     // Default = ""
	Qs    []string  `query:"qs"`    // Default = ""
	Page  int       `query:"page"`  // Default = 1
	Limit int       `query:"limit"` // Default = 10
	Start time.Time `query:"start"` // Default = 7 days ago
	End   time.Time `query:"end"`   // Default = today
	Sort  string    `query:"sort"`  // Default = desc
}

func NewCommonQuery(c echo.Context) (*CommonQuery, error) {
	var cq = CommonQuery{}
	if err := c.Bind(&cq); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid query", c)
	}

	if cq.Q != "" {
		arr := strings.Split(strings.ReplaceAll(cq.Q, "%2C", ","), ",")
		cq.Qs = arr
	}

	if cq.Page < 1 {
		cq.Page = 1
	}

	if cq.Limit < 1 {
		cq.Limit = 10
	}

	if cq.End.IsZero() {
		cq.End = time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)
	}
	if cq.Start.IsZero() {
		cq.Start = cq.End.AddDate(0, 0, -7)
	}

	if cq.Sort != "asc" && cq.Sort != "desc" {
		cq.Sort = "desc"
	}
	return &cq, nil
}

func (dr *CommonQuery) PreviousPeriod() *CommonQuery {
	if dr.Start.Unix() != 0 {
		interval := dr.End.Unix() - dr.Start.Unix()
		dr.Start = time.Unix(dr.Start.Unix()-interval, 0)
		dr.End = time.Unix(dr.Start.Unix()-1, 0)
	}
	return dr
}
