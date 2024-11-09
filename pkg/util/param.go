package util

import (
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type CommonQuery struct {
	Q         string
	Type      string
	Status    string
	UserId    primitive.ObjectID
	ProjectId primitive.ObjectID
	Start     time.Time
	End       time.Time
}

func NewCommonQuery(c echo.Context) *CommonQuery {
	qParam := c.QueryParam("q")
	statusParam := c.QueryParam("status")
	typeParam := c.QueryParam("type")
	userIdParam := c.QueryParam("userId")
	projectIdParam := c.QueryParam("projectId")
	startParam := c.QueryParam("start")
	endParam := c.QueryParam("end")

	RefineString(&statusParam)
	RefineString(&typeParam)

	cq := CommonQuery{
		Q:      qParam,
		Status: statusParam,
		Type:   typeParam,
		UserId: primitive.NilObjectID,
		Start:  time.UnixMilli(0),
		End:    time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 23, 59, 59, 0, time.Local),
	}

	if len(userIdParam) > 0 {
		userId, err := primitive.ObjectIDFromHex(userIdParam)
		if err == nil {
			cq.UserId = userId
		}
	}

	if len(projectIdParam) > 0 {
		projectId, err := primitive.ObjectIDFromHex(projectIdParam)
		if err == nil {
			cq.ProjectId = projectId
		}
	}

	if len(startParam) > 0 {
		start, err := time.Parse(time.RFC3339, startParam)
		if err == nil {
			cq.Start = start
		}
	}

	if len(endParam) > 0 {
		end, err := time.Parse(time.RFC3339, endParam)
		if err == nil {
			cq.End = end
		}
	}
	return &cq
}

func (dr *CommonQuery) PreviousPeriod() *CommonQuery {
	if dr.Start.Unix() != 0 {
		interval := dr.End.Unix() - dr.Start.Unix()
		dr.Start = time.Unix(dr.Start.Unix()-interval, 0)
		dr.End = time.Unix(dr.Start.Unix()-1, 0)
	}
	return dr
}
