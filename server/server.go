package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/shiyongxi/go-common/logger"
	"github.com/soheilhy/cmux"
	"golang.org/x/sync/errgroup"
	"golang.org/x/text/language"
	"net"
	"net/http"
	"reflect"
)

type (
	GinServer struct {
		ServerCfg     *Config
		Engine        *gin.Engine
		RouterGroup   *gin.RouterGroup
		I18nBundle    *i18n.Bundle
		LoopCall      func(structs ...interface{})
		RegisterRoute func()
		httpListener  net.Listener
		//GrpcServer    *GrpcServer
		CorsFunc func(ctx *gin.Context)
		//Tracer        opentracing.Tracer
	}

	Config struct {
		ContextPath string  `json:"contextPath" yaml:"contextPath"`
		Host        string  `json:"host" yaml:"host"`
		Port        int     `json:"port" yaml:"port"`
		Mode        string  `json:"mode" yaml:"mode"`
		Debug       bool    `json:"debug" yaml:"debug"`
		TraceParam  float64 `yaml:"traceParam"`
	}
)

const (
	Localizer = "Localizer"
)

func NewGinServer(cfg *Config) *GinServer {
	gin.SetMode(cfg.Mode)
	binding.Validator = new(DefaultValidator)

	return &GinServer{
		ServerCfg:  cfg,
		Engine:     gin.New(),
		I18nBundle: i18n.NewBundle(language.Chinese),
		//GrpcServer: NewGrpcServer(),
		LoopCall: func(structs ...interface{}) {
			for _, v := range structs {
				classType := reflect.TypeOf(v)
				classValue := reflect.ValueOf(v)

				for i := 0; i < classType.NumMethod(); i++ {
					m := classValue.MethodByName(classType.Method(i).Name)
					if m.IsValid() {
						var params []reflect.Value
						m.Call(params)
					}
				}
			}
		},
		CorsFunc: NewCors().Defualt,
		//Tracer: tracer.NewTracer(&tracer.TraceConfig{
		//	Param:       cfg.TraceParam,
		//	ServiceName: strings.TrimPrefix(cfg.ContextPath, "/"),
		//}),
	}
}

func (svc *GinServer) Run() {
	svc.Engine.Use(gin.Recovery())

	if svc.ServerCfg.Debug {
		svc.Engine.Use(gin.Logger())
	}

	addr := fmt.Sprintf("%s:%d", svc.ServerCfg.Host, svc.ServerCfg.Port)

	fmt.Println("addr:", addr)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatal(err)
	}

	m := cmux.New(listener)
	g := new(errgroup.Group)

	//if svc.GrpcServer.RegisteGrpcServer != nil {
	//	svc.GrpcServer.Listener = m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	//	g.Go(func() error { return svc.GrpcServer.RunGrpcServe() })
	//}

	svc.httpListener = m.Match(cmux.HTTP1Fast())

	svc.LoopCall()

	svc.Engine.Use(svc.CorsFunc)

	g.Go(func() error { return svc.RunHttpServe() })
	g.Go(func() error { return m.Serve() })

	fmt.Println("run server: ", g.Wait())
}

func (svc *GinServer) RunHttpServe() error {
	svc.newRoute().RegisterRoute()

	s := &http.Server{Handler: svc.Engine}
	return s.Serve(svc.httpListener)
}

func (svc *GinServer) newRoute() *GinServer {
	//prom := NewPrometheus("gin")
	//prom.Use(svc.Engine)

	svc.RouterGroup = svc.Engine.Group(
		svc.ServerCfg.ContextPath,
		//tracer.NewTracerServer(svc.Tracer).MiddlewareTracerFunc,
		svc.Localizer)

	svc.RouterGroup.GET("/health", new(Controller).Health)

	return svc
}

func (svc *GinServer) Localizer(ctx *gin.Context) {
	localizer := i18n.NewLocalizer(svc.I18nBundle, ctx.Request.FormValue("lang"), ctx.GetHeader("Accept-Language"), "zh-CN")

	ctx.Set(Localizer, localizer)
	ctx.Next()
}
