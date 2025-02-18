package main

import (
	"bluebell/controller"
	"bluebell/dao/mysql"
	"bluebell/dao/redis"
	"bluebell/logger"
	"bluebell/logic"
	"bluebell/pkg/snowflake"
	"bluebell/router"
	"bluebell/setting"
	"flag"
	"fmt"

	"go.uber.org/zap"
)

//	@title			bluebell项目接口文档
//	@version		1.0
//	@description	Go web开发进阶项目实战 bluebell

//	@license.name	Apache 2.0
//	@host			127.0.0.1:8084
//	@BasePath		/api/v1

func main() {
	//if len(os.Args) < 2 {
	//	fmt.Println("need config file.eg: bluebell config.yaml")
	//	return
	//}
	// 使用flag库传递 config 文件地址
	configpath := flag.String("config", "./config.yaml", "")
	flag.Parse()

	// 加载配置
	if err := setting.Init(*configpath); err != nil {
		fmt.Printf("load config failed, err:%v\n", err)
		return
	}
	if err := logger.Init(setting.Conf.LogConfig, setting.Conf.Mode); err != nil {
		fmt.Printf("init logger failed, err:%v\n", err)
		return
	}
	if err := mysql.Init(setting.Conf.MySQLConfig); err != nil {
		fmt.Printf("init mysql failed, err:%v\n", err)
		return
	}
	defer mysql.Close() // 程序退出关闭数据库连接

	if err := redis.Init(setting.Conf.RedisConfig); err != nil {
		fmt.Printf("init redis failed, err:%v\n", err)
		return
	}
	defer redis.Close()

	// 创建持久化任务管理器
	err, persistenceManager := logic.NewPersistence(setting.Conf.RedisPersistenceConfig)
	if err != nil {
		fmt.Printf("create persistence manager failed, err:%v\n", err)
	}
	// 启动持久化任务
	if err := persistenceManager.Start(); err != nil {
		fmt.Printf("start persistence cron job failed, err:%v\n", err)
		return
	}
	defer persistenceManager.Stop() // 确保退出时停止任务

	// 初始化雪花算法
	if err := snowflake.Init(setting.Conf.StartTime, setting.Conf.MachineID); err != nil {
		fmt.Printf("init snowflake failed, err:%v\n", err)
		return
	}

	// 初始化gin框架内置的校验器使用的翻译器
	if err := controller.InitTrans("zh"); err != nil {
		zap.L().Fatal("Init validator trans failed, err: ", zap.Error(err))
	}

	// 注册路由
	r := router.SetupRouter(setting.Conf.Mode)
	err = r.Run(fmt.Sprintf(":%d", setting.Conf.Port))
	if err != nil {
		fmt.Printf("run server failed, err:%v\n", err)
		return
	}
}

// Redis 持久化定时任务
