package clickhouse

import (
	"database/sql"
	"fmt"
	"github.com/ClickHouse/clickhouse-go"
	"github.com/elastic/beats/libbeat/outputs"
	"github.com/elastic/beats/libbeat/outputs/codec"
	"github.com/elastic/beats/libbeat/outputs/outil"
	"github.com/elastic/beats/libbeat/publisher"
	"strings"
	"time"
)

type client struct {
	observer outputs.Observer
	url string
	key      outil.Selector
	password string
	table string
	columns []string
	codec    codec.Codec
	connect *sql.DB
}

func newClient(
	observer outputs.Observer,
	url string,
	table string,
	columns []string,
) *client {
	return &client{
		observer: observer,
		url: url,
		table: table,
		columns: columns,
	}
}

func (c *client) Connect() error {
	logger.Debugf("connect")

	connect, err := sql.Open("clickhouse", c.url)
	if err != nil {
		return err
	}
	if err := connect.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("[%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		} else {
			fmt.Println(err)
		}
		return err
	}
	c.connect = connect

	return nil
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
		}else{
			err = e
		}
		defer stmt.Close()
	}

	if err != nil {
		if tx != nil{
			tx.Rollback()
		}
		logger.Info("will sleep for one minute because an error occurs")
		// Sleep for one minute when an error occurs
		time.Sleep(time.Second * 60)
		batch.RetryEvents(events)
	}else{
		if tx != nil{
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

	return fmt.Sprint("insert into ", c.table, " (", columnStr.String(), ") values (", valueStr.String(), ")", )
}

