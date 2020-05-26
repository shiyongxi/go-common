package server

import (
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/shiyongxi/go-common/logger"
	"net/http"
	"strconv"
)

type (
	Controller struct {
		localizer *i18n.Localizer
	}

	Response struct {
		Status    int         `json:"status"`
		Msg       string      `json:"msg"`
		Data      interface{} `json:"data"`
		Timestamp float64     `json:"timestamp"`
	}
)

func NewController(ctl *Controller) *Controller {
	return ctl
}

func (ctl *Controller) ResponseList(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, data)
}

func (ctl *Controller) Response(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, Response{
		Status: 200,
		Msg:    ctl.getLocalize(ctx).i18nLocalize(200),
		Data:   data,
	})
}

func (ctl *Controller) ParamsException(ctx *gin.Context, err error) {
	logger.Warn(err, "ParamsException", ctx.Request.RequestURI)

	ctx.JSON(http.StatusOK, Response{
		Status: 900,
		Msg:    ctl.getLocalize(ctx).i18nLocalize(900),
		Data:   nil,
	})
}

func (ctl *Controller) ServiceException(ctx *gin.Context, err error) {
	logger.Error(err, "ServiceException", ctx.Request.RequestURI)

	ctx.JSON(http.StatusOK, Response{
		Status: http.StatusInternalServerError,
		Msg:    ctl.getLocalize(ctx).i18nLocalize(http.StatusInternalServerError),
		Data:   nil,
	})
}

func (ctl *Controller) ServiceCodeException(ctx *gin.Context, status int, err error) {
	logger.Error(err, "ServiceCodeException", ctx.Request.RequestURI)

	ctx.JSON(http.StatusOK, Response{
		Status: status,
		Msg:    ctl.getLocalize(ctx).i18nLocalize(status),
		Data:   nil,
	})
}

func (ctl *Controller) UnauthorizedException(ctx *gin.Context) {
	ctx.JSON(http.StatusUnauthorized, Response{
		Status: http.StatusUnauthorized,
		Msg:    http.StatusText(http.StatusUnauthorized),
		Data:   nil,
	})
}

func (ctl *Controller) Health(ctx *gin.Context) {
	ctl.Response(ctx, nil)
}

func (ctl *Controller) i18nLocalize(status int) string {

	return ctl.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: strconv.Itoa(status),
		DefaultMessage: &i18n.Message{
			ID:    strconv.Itoa(status),
			Other: "Internal error in the service",
		},
	})
}

func (ctl *Controller) getLocalize(ctx *gin.Context) *Controller {
	localizer, ok := ctx.Get("Localizer")
	if ok && localizer != nil {
		ctl.localizer = localizer.(*i18n.Localizer)
	}

	return ctl
}
