package main

import (
	"context"
	"fmt"
	"go-gin/cmd/srv/controller"
	"go-gin/pkg/mygin"
	"go-gin/pkg/utils"
	"go-gin/service/singleton"

	"github.com/gin-gonic/gin"
	"github.com/ory/graceful"
	flag "github.com/spf13/pflag"
)

type CliParam struct {
	Version    bool   // Show version
	ConfigName string // Config file name
	Port       uint   // Server port
}

var (
	cliParam CliParam
)

func init() {
	flag.CommandLine.ParseErrorsWhitelist.UnknownFlags = true
	flag.BoolVarP(&cliParam.Version, "version", "v", false, "show version")
	flag.StringVarP(&cliParam.ConfigName, "config", "c", "config", "config file name")
	flag.UintVarP(&cliParam.Port, "port", "p", 0, "server port")
	flag.Parse()
	flag.Lookup("config").NoOptDefVal = "config"
	singleton.InitConfig(cliParam.ConfigName)
	singleton.InitLog(singleton.Conf)
	singleton.InitTimezoneAndCache()
	singleton.InitDBFromPath(singleton.Conf.DBPath)
	initService()
}

func main() {
	if cliParam.Version {
		fmt.Println(singleton.Version)
		return
	}

	port := singleton.Conf.Server.Port
	if cliParam.Port != 0 {
		port = cliParam.Port
	}

	srv := controller.ServerWeb(port)

	startOutput := func() {
		fmt.Println()
		fmt.Println("Server is running with config:")
		utils.PrintStructFieldsAndValues(singleton.Conf, "")

		fmt.Println()
		fmt.Println("Server is running at:")
		fmt.Printf(" - %-7s: %s\n", "Local", utils.Colorize(utils.ColorGreen, fmt.Sprintf("http://127.0.0.1:%d", port)))
		ipv4s, err := utils.GetIPv4NetworkIPs()
		if ipv4s != nil && err == nil {
			for _, ip := range ipv4s {
				fmt.Printf(" - %-7s: %s\n", "Network", utils.Colorize(utils.ColorGreen, fmt.Sprintf("http://%s:%d", ip, port)))
			}
		}

		fmt.Println()
		fmt.Println("Server available routes:")
		mygin.PrintRoute(srv.Handler.(*gin.Engine))
		fmt.Println()
	}

	if err := graceful.Graceful(func() error {
		startOutput()
		return srv.ListenAndServe()
	}, func(c context.Context) error {
		fmt.Print(utils.Colorize("Server is shutting down", utils.ColorRed))
		srv.Shutdown(c)
		return nil
	}); err != nil {
		fmt.Println(utils.Colorize("Server is shutting down with error: %s", utils.ColorRed), err)
	}
}

func initService() {
	// Load all services in the singleton package
	singleton.LoadSingleton()

	if _, err := singleton.Cron.AddFunc("0 * * * * *", sayHello); err != nil {
		panic(err)
	}
}

func sayHello() {
	singleton.Log.Info().Msg("Hello world, I am a cron task")
	// singleton.SendNotificationByType("wecom", "Hello world", "I am a cron task")
}
