module github.com/YHQZ1/esx/services/clearing-house

go 1.22

require (
	github.com/YHQZ1/esx/packages/kafka v0.0.0
	github.com/YHQZ1/esx/packages/logger v0.0.0
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
)

require (
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/rs/zerolog v1.33.0 // indirect
	github.com/segmentio/kafka-go v0.4.47 // indirect
	golang.org/x/sys v0.13.0 // indirect
)

replace (
	github.com/YHQZ1/esx/packages/kafka => ../../packages/kafka
	github.com/YHQZ1/esx/packages/logger => ../../packages/logger
)
