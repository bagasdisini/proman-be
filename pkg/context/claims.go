package context

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"proman-backend/api/repository"
	"proman-backend/internal/config"
	"proman-backend/internal/database"
	_const "proman-backend/pkg/const"
	"proman-backend/pkg/log"
	"strings"
	"sync"
	"time"
)

var onceUserRepo sync.Once
var userRepo *repository.UserCollRepository

type UserClaims struct {
	jwt.StandardClaims
	ID                 string             `json:"id"`
	IDAsObjectID       primitive.ObjectID `json:"-"`
	Role               string             `json:"role"`
	ExpiredDateInMilis int64              `json:"expiredDateInMilis"`
}

func (u *UserClaims) IsAdmin() bool {
	return u.Role == _const.RoleAdmin
}

func (u *UserClaims) IsMaintainer() bool {
	return u.Role == _const.RoleMaintainer
}

func (u *UserClaims) IsAdminOrMaintainer() bool {
	return u.IsMaintainer() || u.IsAdmin()
}

func (u *UserClaims) IsDeveloper() bool {
	return u.Role == _const.RoleDeveloper
}

type Context struct {
	echo.Context
	Claims       *UserClaims
	loggedInUser *repository.User
}

func (c *Context) LoggedInUser() *repository.User {
	if c.loggedInUser == nil {
		onceUserRepo.Do(func() {
			userRepo = repository.NewUserRepository(database.ConnectMongo())
		})
		u, err := userRepo.FindOneByID(c.Claims.IDAsObjectID)
		if err != nil {
			log.Panicc(c, err)
		}
		c.loggedInUser = u
		c.Claims.Role = u.Role
	}
	return c.loggedInUser
}

func NewUserClaimsFromString(s string) (*UserClaims, error) {
	cred := &UserClaims{}
	token, err := jwt.ParseWithClaims(s, cred, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.JWT.Key), nil
	})
	if err != nil {
		return nil, echo.ErrUnauthorized
	}
	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		//Tidak bisa akses jika jwt sudah kadaluarsa.
		if claims.ExpiredDateInMilis < time.Now().UnixNano()/1000000 {
			return nil, echo.ErrUnauthorized
		}

		IDAsObjectID, err := primitive.ObjectIDFromHex(claims.ID)
		if err != nil {
			return nil, err
		}
		claims.IDAsObjectID = IDAsObjectID
		return claims, nil
	}
	return nil, echo.ErrUnauthorized
}

func NewUserClaims(c echo.Context) (*UserClaims, error) {
	// get identity
	header := c.Request().Header.Get("Authorization")
	bearer := strings.Split(header, " ")
	if len(bearer) != 2 {
		return nil, echo.ErrUnauthorized
	}

	if bearer[0] != "Bearer" {
		return nil, echo.ErrUnauthorized
	}

	return NewUserClaimsFromString(bearer[1])
}

func MakeContext(c echo.Context) (*Context, error) {
	claims, err := NewUserClaims(c)
	if err != nil {
		return nil, err
	}
	return &Context{c, claims, nil}, nil
}

func ContextHandler(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		nc, err := MakeContext(c)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		}
		if nc.LoggedInUser().IsDeleted {
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		}
		return next(nc)
	}
}

func AdminOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		nc, ok := c.(*Context)
		if !ok {
			return echo.ErrUnauthorized
		}
		if nc.Claims.IsAdmin() {
			return next(c)
		}
		return echo.ErrUnauthorized
	}
}

func AdminOrMaintainerOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		nc, ok := c.(*Context)
		if !ok {
			return echo.ErrUnauthorized
		}
		if nc.Claims.IsAdminOrMaintainer() {
			return next(c)
		}
		return echo.ErrUnauthorized
	}
}

func MakeToken(u *repository.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = u.ID.Hex()
	claims["role"] = u.Role
	claims["expiredDateInMilis"] = time.Now().AddDate(0, 0, config.JWT.Expire).Unix() * 1000

	accessToken, err := token.SignedString([]byte(config.JWT.Key))
	if err != nil {
		return "", echo.NewHTTPError(http.StatusInternalServerError, "Internal server exception: "+err.Error()).SetInternal(err)
	}
	return accessToken, nil
}
