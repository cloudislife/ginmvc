package main

import (
	"fmt"
	"github.com/mritd/ginmvc/cache"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/mritd/ginmvc/auth"
	"github.com/mritd/ginmvc/ginengine"

	"github.com/mritd/ginmvc/models"

	"github.com/mritd/ginmvc/db"
	"github.com/mritd/ginmvc/middleware"
	"github.com/sirupsen/logrus"

	"github.com/mritd/ginmvc/routers"

	"github.com/mritd/ginmvc/utils"

	"github.com/mritd/ginmvc/conf"
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "ginmvc",
	Short: "Gin mvc template",
	Long: `
Gin mvc template.`,
	Run: func(cmd *cobra.Command, args []string) {

		// init framework log
		initLog()
		// init memory cache(you can also choose to use redis)
		// if you modify the cache implementation
		cache.InitMemCache()
		// init mysql(gorm)
		db.InitMySQL()
		// migrate db schema
		models.AutoMigrate()
		// load casbin
		auth.InitCasbin()
		// init gin router engine
		ginengine.Init()
		// load middleware
		middleware.Setup()
		// add gin router
		routers.Setup()

		// run gin http server
		addr := fmt.Sprint(conf.Basic.Addr, ":", conf.Basic.Port)
		logrus.Infof("server listen at %s", addr)
		utils.CheckAndExit(ginengine.Engine.Run(addr))

	},
}

func init() {
	// load config file
	cobra.OnInitialize(initConfig)
	// cmd config flag
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "ginmvc.yaml", "config file (default is ./ginmvc.yaml)")
}

func initConfig() {

	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		_, err := os.Create(cfgFile)
		utils.CheckAndExit(err)
		conf.Basic = conf.ExampleConfig()
		utils.CheckAndExit(conf.Basic.WriteTo(cfgFile))
	} else if err != nil {
		utils.CheckAndExit(err)
	}
	utils.CheckAndExit(conf.Basic.LoadFrom(cfgFile))
}

// init log config
func initLog() {
	if conf.Basic.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	var logFile io.Writer
	var err error
	if strings.ToLower(conf.Basic.LogFile) != "" && strings.ToLower(conf.Basic.LogFile) != "stdout" {
		logFile, err = os.OpenFile(conf.Basic.LogFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
		utils.CheckAndExit(err)
	} else {
		logFile = os.Stdout
	}

	logrus.SetOutput(logFile)
	logrus.Infof("GOMAXPROCS: %d", runtime.NumCPU())
}

func main() {
	cores := runtime.NumCPU()
	runtime.GOMAXPROCS(cores)
	utils.CheckAndExit(rootCmd.Execute())
}
