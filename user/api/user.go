// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"WMSS/user/api/internal/middleware"
	"flag"
	"fmt"

	"WMSS/user/api/internal/config"
	"WMSS/user/api/internal/handler"
	"WMSS/user/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/user-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	server.Use(middleware.SimpleRecovery())

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
