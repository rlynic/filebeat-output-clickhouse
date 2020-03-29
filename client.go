package clickhouse_20200328

import (
	"database/sql"
	"fmt"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/outputs"
	"github.com/elastic/beats/libbeat/outputs/codec"
	"github.com/elastic/beats/libbeat/outputs/outil"
	"github.com/elastic/beats/libbeat/publisher"
	"strings"
	"time"

	_ "github.com/ClickHouse/clickhouse-go"
)

type client struct {
	observer      outputs.Observer
	url           string
	key           outil.Selector
	password      string
	table         string
	columns       []string
	retryInterval int
	codec         codec.Codec
	connect       *sql.DB
}

func newClient(
	observer outputs.Observer,
	url string,
	table string,
	columns []string,
	retryInterval int,
) *client {
	return &client{
		observer:      observer,
		url:           url,
		table:         table,
		columns:       columns,
		retryInterval: retryInterval,
	}
}

func (c *client) Connect() error {
	logger.Debugf("connect")

	var err error
	connect, err := sql.Open("clickhouse", c.url)
	if err == nil {
		if e := connect.Ping(); e != nil {
			err = e
		}
	}
	if err != nil {
		c.sleepBeforeRetry(err)
	}
	c.connect = connect

	return err
}

func (c *client) Close() error {
	logger.Debugf("close connection")
	return c.connect.Close()
}

func (c *client) Publish(batch publisher.Batch) error {
	if c == nil {
		panic("no client")
	}
	if batch == nil {
		panic("no batch")
	}

	events := batch.Events()
	c.observer.NewBatch(len(events))

	rows := c.extractData(events)
	sql := c.generateSql()

	var err error
	tx, err := c.connect.Begin()
	if err == nil {
		stmt, e := tx.Prepare(sql)
		if e == nil {
			for _, row := range rows {
				_, e := stmt.Exec(row...)
				if e != nil {
					err = e
					break
				}
			}
		} else {
			err = e
		}
		if stmt != nil {
			defer stmt.Close()
		}
	}
	if err != nil {
		if tx != nil {
			tx.Rollback()
		}
		c.sleepBeforeRetry(err)
		batch.RetryEvents(events)
	} else {
		if tx != nil {
			tx.Commit()
		}
		batch.ACK()
	}

	return err
}

func (c *client) String() string {
	return "clickhouse(" + c.url + ")"
}

func (c *client) extractData(events []publisher.Event) [][]interface{} {
	cSize := len(c.columns)
	rows := make([][]interface{}, len(events))
	for i, event := range events {
		content := event.Content
		row := make([]interface{}, cSize)
		for i, c := range c.columns {
			if _, e := content.Fields[c]; e {
				row[i], _ = content.GetValue(c)
			}
		}
		rows[i] = row
	}

	return rows
}

func (c *client) generateSql() string {
	cSize := len(c.columns)
	var columnStr, valueStr strings.Builder
	for i, cl := range c.columns {
		columnStr.WriteString(cl)
		valueStr.WriteString("?")
		if i < cSize-1 {
			columnStr.WriteString(",")
			valueStr.WriteString(",")
		}
	}

	return fmt.Sprint("insert into ", c.table, " (", columnStr.String(), ") values (", valueStr.String(), ")")
}

func (c *client) sleepBeforeRetry(err error) {
	logp.Err("will sleep for %v seconds because an error occurs: %s", c.retryInterval, err)
	time.Sleep(time.Second * time.Duration(c.retryInterval))
}
