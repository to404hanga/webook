//go:build !k8s

package config

// var Config = config{
// 	DB: DBConfig{
// 		DSN: "root:root@tcp(localhost:13316)/webook",
// 	},
// 	Redis: RedisConfig{
// 		Addr: "localhost:16379",
// 	},
// }

var Config = config{
	DB: DBConfig{
		DSN: "root:123456@tcp(localhost:3306)/webook",
	},
	Redis: RedisConfig{
		Addr: "localhost:6379",
	},
}
