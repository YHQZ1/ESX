module github.com/YHQZ1/esx/services/risk-engine

go 1.22

require (
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
	github.com/rs/zerolog v1.33.0
	google.golang.org/grpc v1.64.0
)

require google.golang.org/protobuf v1.34.2 // indirect

require (
	github.com/YHQZ1/esx/packages/proto v0.0.0
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240604185151-ef581f913117 // indirect
)

replace github.com/YHQZ1/esx/packages/proto => ../../packages/proto
