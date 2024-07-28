package util

import (
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	_const "proman-backend/pkg/const"
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

type ScheduleQuery struct {
	Q           string
	Type        string
	Contributor []primitive.ObjectID
	Start       time.Time
	End         time.Time
}

func NewScheduleQuery(c echo.Context) (*ScheduleQuery, error) {
	var cq = ScheduleQuery{}

	cq.Q = c.QueryParam("q")
	cq.Type = c.QueryParam("type")
	if len(cq.Type) > 0 && !_const.IsValidScheduleType(cq.Type) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid type")
	}

	contributorStr := c.QueryParam("contributor")
	if len(contributorStr) > 0 {
		contributorArr := strings.Split(contributorStr, ",")
		for _, v := range contributorArr {
			contributorID, err := primitive.ObjectIDFromHex(v)
			if err != nil {
				return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid contributor ID")
			}
			cq.Contributor = append(cq.Contributor, contributorID)
		}
	}

	startStr := c.QueryParam("start")
	if len(startStr) > 0 {
		start, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid start date")
		}
		cq.Start = start
	}

	endStr := c.QueryParam("end")
	if len(endStr) > 0 {
		end, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid end date")
		}
		cq.End = end
	}
	return &cq, nil
}
