package clickhouse

import (
	"github.com/elastic/beats/libbeat/outputs/codec"
)

type redisConfig struct {
	Url			string 				  `config:"url"`
	Table       string                `config:"table"`
	Columns     []string              `config:"columns"`
	Codec       codec.Config          `config:"codec"`
	BulkMaxSize int                   `config:"bulk_max_size"`
	MaxRetries  int                   `config:"max_retries"`
}

var (
	defaultConfig = redisConfig{
		Url:			"tcp://127.0.0.1:9000",
		BulkMaxSize:	1000,
		MaxRetries:		3,
	}
)
