package config

import "github.com/spf13/viper"

func InitConfig() {
	viper.SetConfigName("config")    // 配置文件名
	viper.SetConfigType("yaml")      // 配置文件类型
	viper.AddConfigPath("../config") // 配置文件路径
	// viper.SetConfigFile("../confg/config.yml") // 指定配置文件以及路径

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}
