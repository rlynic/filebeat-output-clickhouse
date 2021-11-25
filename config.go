package clickhouse_20200328

import (
	"github.com/elastic/beats/v7/libbeat/outputs/codec"
)

type clickHouseConfig struct {
	Url                   string       `config:"url"`
	Table                 string       `config:"table"`
	Columns               []string     `config:"columns"`
	Codec                 codec.Config `config:"codec"`
	BulkMaxSize           int          `config:"bulk_max_size"`
	MaxRetries            int          `config:"max_retries"`
	RetryInterval         int          `config:"retry_interval"`
	SkipUnexpectedTypeRow bool         `config:"skip_unexpected_type_row"`
}

var (
	defaultConfig = clickHouseConfig{
		Url:                   "tcp://127.0.0.1:9000",
		BulkMaxSize:           1000,
		MaxRetries:            3,
		RetryInterval:         60,
		SkipUnexpectedTypeRow: false,
	}
)
