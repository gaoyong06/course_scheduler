package cron

// import (
// 	"errors"
// 	"flag"
// 	"fmt"
// 	"os"
// 	"path/filepath"
// 	"time"

// 	"buffcloud/tracer/internal/util"

// 	"com.xycloud.halo/logger"

// 	haloRedis "com.xycloud.halo/databases/redis"
// 	nested "github.com/antonfisher/nested-logrus-formatter"
// 	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
// 	"github.com/robfig/cron"
// 	log "github.com/sirupsen/logrus"
// 	"github.com/spf13/viper"
// )

// // var (
// // 	MODE    string
// // 	VERSION string

// // 	configFile        string
// // 	defaultConfigFile = "conf/config.yaml"

// // 	callAction string
// // )

// func main() {
// 	if MODE == "" {
// 		MODE = "debug"
// 	}

// 	err := Init()
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	//直接调用命令，不进入cron流程
// 	if callAction != "" {
// 		err = HandlerCallAction()
// 		if err != nil {
// 			fmt.Printf("HandlerCallAction error: %s\n", err.Error())
// 		} else {
// 			fmt.Printf("HandlerCallAction : %s done\n", callAction)
// 		}
// 		return
// 	}

// 	//根据配置确认是否开启cron
// 	enable := viper.GetBool("cron.enable")
// 	if !enable {
// 		logger.Log().Infof("[Cron]Cron disabled, begin noop loop")
// 		select {}
// 		return
// 	}

// 	c := cron.New()

// 	//按天发送
// 	c.AddFunc("0 10 0 * * *", DailyNameCodeFailStat)
// 	c.AddFunc("0 */15 * * * *", DependencyMonitor)
// 	c.AddFunc("0 */1 * * * *", CleanTracerSpans)
// 	c.Start()

// 	select {}

// }

// func HandlerCallAction() error {
// 	if callAction == "DailyNameCodeFailStat" {
// 		DailyNameCodeFailStat()
// 		return nil
// 	} else if callAction == "DependencyMonitor" {
// 		DependencyMonitor()
// 		return nil
// 	} else if callAction == "CleanTracerSpans" {
// 		CleanTracerSpans()
// 		return nil
// 	}

// 	return errors.New("Unkonw action:" + callAction)
// }

// func Init() error {
// 	err := initFlag()
// 	if err != nil {
// 		return err
// 	}

// 	//init config
// 	err = initConfig(configFile)
// 	if err != nil {
// 		return errors.New("config init failed, err:" + err.Error())
// 	}

// 	err = initLog()
// 	if err != nil {
// 		return err
// 	}

// 	err = initRedis()
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// //init flag
// func initFlag() error {
// 	//init params
// 	h := flag.Bool("h", false, "application help")

// 	c := flag.String("c", "", "config file, ex: /data/config.yaml")

// 	a := flag.String("a", "", "action, ex: DailyNameCodeFailStat")

// 	flag.Parse()

// 	if *h {
// 		flag.PrintDefaults()
// 		os.Exit(0)
// 		return nil
// 	}

// 	if *c == "" {
// 		//use default config file
// 		path, _ := os.Executable()
// 		binPath := filepath.Dir(path)
// 		*c = binPath + "/../" + defaultConfigFile
// 	}

// 	if *a != "" {
// 		callAction = *a
// 	}

// 	configFile = *c

// 	return nil
// }

// func initConfig(configFile string) error {
// 	path, _ := os.Executable()
// 	RootPath := filepath.Dir(path)

// 	viper.Set("path.root", RootPath)

// 	configPath := filepath.Dir(configFile)
// 	fileName := filepath.Base(configFile)

// 	//设置读取的配置文件
// 	viper.SetConfigName(fileName)
// 	//添加读取的配置文件路径
// 	viper.AddConfigPath(configPath)
// 	//设置配置文件类型
// 	viper.SetConfigType("yaml")

// 	if err := viper.ReadInConfig(); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func initLog() error {
// 	logPathPrefix := fmt.Sprintf("%s/%s", viper.GetString("path.root"), viper.GetString("log.prefix"))
// 	logName := viper.GetString("log.name")
// 	logLevel := viper.GetInt("log.level")

// 	path := fmt.Sprintf("%s%s", logPathPrefix, logName)
// 	/* 日志轮转相关函数
// 	`WithLinkName` 为最新的日志建立软连接
// 	`WithRotationTime` 设置日志分割的时间，隔多久分割一次
// 	`WithMaxAge 和 WithRotationCount二者只能设置一个
// 	`WithMaxAge` 设置文件清理前的最长保存时间
// 	`WithRotationCount` 设置文件清理前最多保存的个数
// 	*/
// 	// 下面配置日志每隔 1 分钟轮转一个新文件，保留最近 3 分钟的日志文件，多余的自动清理掉。
// 	writer, err := rotatelogs.New(
// 		path+".%Y%m%d",
// 		rotatelogs.WithLinkName(path),
// 		rotatelogs.WithRotationCount(31),
// 		rotatelogs.WithRotationTime(time.Duration(86400)*time.Second),
// 	)
// 	if err != nil {
// 		return errors.New("initLog error:" + err.Error())
// 	}
// 	log.SetReportCaller(true)
// 	log.SetFormatter(&nested.Formatter{
// 		HideKeys:         true,
// 		NoUppercaseLevel: true,
// 		ShowFullLevel:    true,
// 		TimestampFormat:  "2006/01/02 15:04:05",
// 		NoFieldsSpace:    true,
// 		NoFieldsColors:   true,
// 		NoColors:         true,
// 		CallerFirst:      true,
// 		TrimMessages:     true,
// 		// CustomCallerFormatter: func(f *runtime.Frame) string {
// 		// 	s := strings.Split(f.Function, ".")
// 		// 	funcName := s[len(s)-1]
// 		// 	return fmt.Sprintf(" [%s:%d][%s()]", filepath.Base(f.File), f.Line, funcName)
// 		// },
// 		FieldsOrder: []string{"component", "category"},
// 	})

// 	// log.SetFormatter(&new_log.NewTextFormatter{})
// 	log.SetLevel(log.Level(logLevel))
// 	log.SetOutput(writer)
// 	return nil
// }

// func initRedis() error {
// 	defaultRedisConf := haloRedis.RedisConf{}
// 	defaultRedisConf.Host = viper.GetString("redis.default.host")
// 	defaultRedisConf.Port = viper.GetInt("redis.default.port")
// 	defaultRedisConf.Pwd = viper.GetString("redis.default.pwd")
// 	defaultRedisConf.Timeout = viper.GetInt("redis.default.timeout")

// 	util.NewRedisInstance("default", &defaultRedisConf)
// 	rdb, _ := util.GetRedis("default")
// 	_, err := rdb.Ping().Result()
// 	if err != nil {
// 		return errors.New("default redis instance init failed: " + err.Error())
// 	}
// 	logger.Log().Infof("Redis Server Init  on [%s:%d]", defaultRedisConf.Host, defaultRedisConf.Port)

// 	return nil
// }
