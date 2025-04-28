module mykit

go 1.16

require (
	github.com/coreos/etcd v3.3.27+incompatible
	github.com/didi/gendry v1.9.0
	github.com/gin-contrib/gzip v1.2.3
	github.com/gin-gonic/gin v1.10.0
	github.com/go-ini/ini v1.67.0
	github.com/go-playground/locales v0.14.1
	github.com/go-playground/universal-translator v0.18.1
	github.com/go-playground/validator/v10 v10.26.0
	github.com/go-redis/redis/v8 v8.11.5
	github.com/go-redis/redismock/v8 v8.11.5
	github.com/go-sql-driver/mysql v1.9.2
	github.com/go-xorm/xorm v0.7.9
	github.com/golang/protobuf v1.5.4
	github.com/google/uuid v1.6.0
	github.com/influxdata/influxdb v1.12.0
	github.com/jmoiron/sqlx v1.4.0
	github.com/levigross/grequests v0.0.0-20231203190023-9c307ef1f48d
	github.com/micro/go-micro/v2 v2.9.1
	github.com/micro/go-plugins/registry/etcdv3 v0.0.0-00010101000000-000000000000
	github.com/prometheus-community/pro-bing v0.7.0
	github.com/prometheus/client_golang v1.22.0
	github.com/shopspring/decimal v1.4.0
	go.opentelemetry.io/otel v1.35.0
	go.opentelemetry.io/otel/sdk v1.35.0
	go.opentelemetry.io/otel/trace v1.35.0
	go.uber.org/zap v1.27.0
	google.golang.org/grpc v1.64.1
	google.golang.org/protobuf v1.36.6
)

replace (
	github.com/micro/go-plugins/registry/etcdv3 => ./core/go_plugins/registry/etcdv3
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)
