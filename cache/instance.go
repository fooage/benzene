package cache

import "github.com/spf13/viper"

var (
	Carbon *UniversalCache
)

func CreateInstance() {
	// create a carbon which means cache node instance
	Carbon = NewUniversalCache(viper.GetInt64("cache.max_bytes"))
}
