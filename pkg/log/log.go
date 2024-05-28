package log

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"net/http"
	"runtime"
)

var e *echo.Echo
var recoverConfig middleware.RecoverConfig

func loggerSkipper(c echo.Context) bool {
	return c.Response().Status == http.StatusOK ||
		c.Response().Status == http.StatusBadRequest ||
		c.Response().Status == http.StatusAccepted ||
		c.Response().Status == http.StatusCreated ||
		c.Response().Status == http.StatusForbidden ||
		c.Response().Status == http.StatusUnauthorized ||
		c.Response().Status == http.StatusNotFound
}

func SetLogger(ech *echo.Echo) {
	ech.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: loggerSkipper,
		Format:  "${time_rfc3339} ${method} ${uri} ${status} ${error}\n",
		Output:  middleware.DefaultLoggerConfig.Output,
	}))

	recoverConfig = middleware.RecoverConfig{
		Skipper:           middleware.DefaultSkipper,
		StackSize:         middleware.DefaultRecoverConfig.StackSize,
		DisableStackAll:   middleware.DefaultRecoverConfig.DisableStackAll,
		DisablePrintStack: middleware.DefaultRecoverConfig.DisablePrintStack,
		LogLevel:          log.INFO,
	}
	ech.Use(middleware.RecoverWithConfig(recoverConfig))

	ech.Logger.SetLevel(log.INFO)
	ech.Logger.SetHeader("${time_rfc3339} ${level} ${short_file}:${line} -")

	e = ech
}

func Print(i ...interface{}) {
	e.Logger.Print(i...)
}

func Printf(format string, args ...interface{}) {
	e.Logger.Printf(format, args...)
}

func Printc(c echo.Context, i ...interface{}) {
	newI := append([]interface{}{c.Request().RequestURI + " - "}, i...)
	e.Logger.Print(newI...)
}

func Printcf(c echo.Context, format string, args ...interface{}) {
	newFormat := c.Request().RequestURI + " - " + format
	e.Logger.Printf(newFormat, args...)
}

func Debug(i ...interface{}) {
	e.Logger.Debug(i...)
}

func Debugf(format string, args ...interface{}) {
	e.Logger.Debugf(format, args...)
}

func Debugc(c echo.Context, i ...interface{}) {
	newI := append([]interface{}{c.Request().RequestURI + " - "}, i...)
	e.Logger.Debug(newI...)
}

func Debugcf(c echo.Context, format string, args ...interface{}) {
	newFormat := c.Request().RequestURI + " - " + format
	e.Logger.Debugf(newFormat, args...)
}

func Info(i ...interface{}) {
	e.Logger.Info(i...)
}

func Infof(format string, args ...interface{}) {
	e.Logger.Infof(format, args...)
}

func Infoc(c echo.Context, i ...interface{}) {
	newI := append([]interface{}{c.Request().RequestURI + " - "}, i...)
	e.Logger.Info(newI...)
}

func Infocf(c echo.Context, format string, args ...interface{}) {
	newFormat := c.Request().RequestURI + " - " + format
	e.Logger.Infof(newFormat, args...)
}

func Warn(i ...interface{}) {
	e.Logger.Warn(i...)
}

func Warnf(format string, args ...interface{}) {
	e.Logger.Warnf(format, args...)
}

func Warnc(c echo.Context, i ...interface{}) {
	newI := append([]interface{}{c.Request().RequestURI + " - "}, i...)
	e.Logger.Warn(newI...)
}

func Warncf(c echo.Context, format string, args ...interface{}) {
	newFormat := c.Request().RequestURI + " - " + format
	e.Logger.Warnf(newFormat, args...)
}

func Error(i ...interface{}) {
	e.Logger.Error(i...)
}

func Errorf(format string, args ...interface{}) {
	e.Logger.Errorf(format, args...)
}

func Errorc(c echo.Context, i ...interface{}) {
	newI := append([]interface{}{c.Request().RequestURI + " - "}, i...)
	e.Logger.Error(newI...)
}

func Errorcf(c echo.Context, format string, args ...interface{}) {
	newFormat := c.Request().RequestURI + " - " + format
	e.Logger.Errorf(newFormat, args...)
}

func Fatal(i ...interface{}) {
	e.Logger.Fatal(i...)
}

func Fatalf(format string, args ...interface{}) {
	e.Logger.Fatalf(format, args...)
}

func Fatalc(c echo.Context, i ...interface{}) {
	newI := append([]interface{}{c.Request().RequestURI + " - "}, i...)
	e.Logger.Fatal(newI...)
}

func Fatalcf(c echo.Context, format string, args ...interface{}) {
	newFormat := c.Request().RequestURI + " - " + format
	e.Logger.Fatalf(newFormat, args...)
}

func Panic(i ...interface{}) {
	e.Logger.Panic(i...)
}

func Panicf(format string, args ...interface{}) {
	e.Logger.Panicf(format, args...)
}

func Panicc(c echo.Context, i ...interface{}) {
	newI := append([]interface{}{c.Request().RequestURI + " - "}, i...)
	e.Logger.Panic(newI...)
}

func Paniccf(c echo.Context, format string, args ...interface{}) {
	newFormat := c.Request().RequestURI + " - " + format
	e.Logger.Panicf(newFormat, args...)
}

func RecoverWithTrace() {
	if r := recover(); r != nil {
		err, ok := r.(error)
		if !ok {
			err = fmt.Errorf("%v", r)
		}
		stack := make([]byte, recoverConfig.StackSize)
		length := runtime.Stack(stack, !recoverConfig.DisableStackAll)
		if !recoverConfig.DisablePrintStack {
			msg := fmt.Sprintf("[PANIC RECOVER] %v %s\n", err, stack[:length])
			switch recoverConfig.LogLevel {
			case log.DEBUG:
				e.Logger.Debug(msg)
			case log.INFO:
				e.Logger.Info(msg)
			case log.WARN:
				e.Logger.Warn(msg)
			case log.ERROR:
				e.Logger.Error(msg)
			case log.OFF:
				//None.
			default:
				e.Logger.Print(msg)
			}
		}
	}
}
