package util

import (
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/v2/bson"
	"math"
	"strconv"
	"strings"
	"time"
)

type CommonQuery struct {
	Q         string
	Type      string
	Status    string
	UserId    bson.ObjectID
	ProjectId bson.ObjectID
	Start     time.Time
	End       time.Time

	Sort  int8
	Page  int64
	Limit int64
}

func NewCommonQuery(c echo.Context) *CommonQuery {
	qParam := c.QueryParam("q")
	statusParam := strings.ToLower(strings.TrimSpace(c.QueryParam("status")))
	typeParam := strings.ToLower(strings.TrimSpace(c.QueryParam("type")))
	userIdParam := strings.TrimSpace(c.QueryParam("userId"))
	projectIdParam := strings.TrimSpace(c.QueryParam("projectId"))
	startParam := strings.TrimSpace(c.QueryParam("start"))
	endParam := strings.TrimSpace(c.QueryParam("end"))
	sortParam := strings.TrimSpace(c.QueryParam("sort"))
	pageParam := strings.TrimSpace(c.QueryParam("page"))
	limitParam := strings.TrimSpace(c.QueryParam("limit"))

	cq := CommonQuery{
		Q:      qParam,
		Status: statusParam,
		Type:   typeParam,
		UserId: bson.NilObjectID,
		Start:  time.UnixMilli(0),
		End:    time.UnixMilli(math.MaxInt64),

		Sort:  1,
		Page:  1,
		Limit: math.MaxInt64,
	}

	if len(userIdParam) > 0 {
		userId, err := bson.ObjectIDFromHex(userIdParam)
		if err == nil {
			cq.UserId = userId
		}
	}

	if len(projectIdParam) > 0 {
		projectId, err := bson.ObjectIDFromHex(projectIdParam)
		if err == nil {
			cq.ProjectId = projectId
		}
	}

	if len(startParam) > 0 {
		startUnixMilli, err := strconv.ParseInt(startParam, 10, 64)
		if err == nil {
			cq.Start = time.UnixMilli(startUnixMilli)
		}
	}

	if len(endParam) > 0 {
		endUnixMilli, err := strconv.ParseInt(endParam, 10, 64)
		if err == nil {
			cq.End = time.UnixMilli(endUnixMilli)
		}
	}

	if len(sortParam) > 0 {
		switch sortParam {
		case "desc":
			cq.Sort = -1
		case "asc":
			cq.Sort = 1
		}
	}

	if len(pageParam) > 0 {
		page, err := strconv.ParseInt(pageParam, 10, 64)
		if err == nil || page > 0 {
			cq.Page = page
		}
	}

	if len(limitParam) > 0 {
		limit, err := strconv.ParseInt(limitParam, 10, 64)
		if err == nil || limit > 0 {
			cq.Limit = limit
		}
	}
	return &cq
}

func NilCommonQuery() *CommonQuery {
	dr := &CommonQuery{}
	dr.Q = ""
	dr.Status = ""
	dr.Type = ""
	dr.UserId = bson.NilObjectID
	dr.ProjectId = bson.NilObjectID
	dr.Start = time.UnixMilli(0)
	dr.End = time.UnixMilli(math.MaxInt64)
	dr.Sort = 1
	dr.Page = 1
	dr.Limit = math.MaxInt64
	return dr
}

func (dr *CommonQuery) PreviousPeriod() *CommonQuery {
	if dr.Start.Unix() != 0 {
		interval := dr.End.Unix() - dr.Start.Unix()
		dr.Start = time.Unix(dr.Start.Unix()-interval, 0)
		dr.End = time.Unix(dr.Start.Unix()-1, 0)
	}
	return dr
}

func (dr *CommonQuery) ResetAll() *CommonQuery {
	dr.Q = ""
	dr.Status = ""
	dr.Type = ""
	dr.UserId = bson.NilObjectID
	dr.ProjectId = bson.NilObjectID
	dr.Start = time.UnixMilli(0)
	dr.End = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 23, 59, 59, 0, time.Local)
	dr.Sort = 1
	dr.Page = 1
	dr.Limit = math.MaxInt64
	return dr
}

func (dr *CommonQuery) ResetDate() *CommonQuery {
	dr.Start = time.UnixMilli(0)
	dr.End = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 23, 59, 59, 0, time.Local)
	return dr
}

func (dr *CommonQuery) ResetPagination() *CommonQuery {
	dr.Page = 1
	dr.Limit = math.MaxInt64
	return dr
}
