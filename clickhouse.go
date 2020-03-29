package clickhouse

import (
	"errors"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/outputs"
)


var logger = logp.NewLogger("clickhouse")

func init() {
	outputs.RegisterType("clickhouse", makeClickHouse)
}

func makeClickHouse(
	beat beat.Info,
	observer outputs.Observer,
	cfg *common.Config,
) (outputs.Group, error) {

	config := defaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return outputs.Fail(err)
	}

	if len(config.Table) == 0 {
		return outputs.Fail(errors.New("the table name must be set"))
	}

	if len(config.Columns) == 0 {
		return outputs.Fail(errors.New("the table columns must be set"))
	}

	client := newClient(observer, config.Url, config.Table, config.Columns)

	return outputs.Success(config.BulkMaxSize, config.MaxRetries, client)
}
