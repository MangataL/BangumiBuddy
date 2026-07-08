package discovery

const DefaultMikanHost = "mikanani.me"

type Config struct {
	MikanHost string `mapstructure:"mikan_host" json:"mikanHost" default:"mikanani.me"`
}
