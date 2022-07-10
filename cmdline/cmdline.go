package cmdline

import (
	"strings"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func Setup(envPrefix string) {
	log.SetFormatter(&log.TextFormatter{TimestampFormat: "2006-01-02 15:04:05"})

	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	flag.CommandLine.SortFlags = false
	flag.Parse()

	// replace - with _ in flags so we can use the snake_case version when accessing through viper
	normalizeFunc := flag.CommandLine.GetNormalizeFunc()
	flag.CommandLine.SetNormalizeFunc(func(fs *flag.FlagSet, name string) flag.NormalizedName {
		result := normalizeFunc(fs, name)
		name = strings.ReplaceAll(string(result), "-", "_")
		return flag.NormalizedName(name)
	})

	viper.BindPFlags(flag.CommandLine)

	if viper.GetBool("debug") {
		log.SetLevel(log.DebugLevel)
	} else if viper.GetBool("verbose") {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}

}
