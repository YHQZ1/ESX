module github.com/YHQZ1/esx/services/matching-engine

go 1.22

require (
	github.com/YHQZ1/esx/packages/kafka v0.0.0
	github.com/YHQZ1/esx/packages/logger v0.0.0
	github.com/YHQZ1/esx/packages/proto v0.0.0
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
	github.com/redis/go-redis/v9 v9.5.1
	google.golang.org/grpc v1.64.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/rs/zerolog v1.33.0 // indirect
	github.com/segmentio/kafka-go v0.4.47 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
)

replace (
	github.com/YHQZ1/esx/packages/kafka => ../../packages/kafka
	github.com/YHQZ1/esx/packages/logger => ../../packages/logger
	github.com/YHQZ1/esx/packages/proto => ../../packages/proto
)
