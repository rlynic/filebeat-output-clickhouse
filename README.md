# filebeat-output-clickhouse
The output for filebeat support push events to ClickHouse，You need to recompile filebeat with the ClickHouse Output.

# Compile
## We need clone beats first
```$xslt
git clone git@github.com:elastic/beats.git
```

## then, Install clickhouse Output, under GOPATH directory
```
go get -u github.com/rlynic/filebeat-output-clickhouse
```

## modify beats outputs includes, add clickhouse output
```
cd {your beats directory}/github.com/elastic/beats/libbeat/publisher/includes/includes.go
```
```
import (
	...
	_ "github.com/rlynic/filebeat-output-clickhouse"
)
```
## build package, in filebeat
```
cd {your beats directory}/github.com/elastic/beats/filebeat
make
```

# Configure Output
## clickHouse output configuration
```
#----------------------------- ClickHouse output --------------------------------
output.clickHouse:
  # clickHouse tcp link address
  # https://github.com/ClickHouse/clickhouse-go
  # example tcp://host1:9000?username=user&password=qwerty&database=clicks&read_timeout=10&write_timeout=20&alt_hosts=host2:9000,host3:9000
  url: "tcp://127.0.0.1:9000?debug=true"
  # table name for receive data
  table: ck_test
  # table columns for data filter, match the keys in log file
  columns: ["id", "name", "created_date"]
  # will sleep the retry_interval seconds when unexpected exception, default 60s
  retry_interval: 60
```
