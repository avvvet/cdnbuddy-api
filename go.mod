module github.com/avvvet/cdnbuddy-api

go 1.23.2

require (
	github.com/cachefly/cachefly-go-sdk v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	github.com/nats-io/nats.go v1.43.0
)

require github.com/sirupsen/logrus v1.9.3

require (
	github.com/go-chi/chi/v5 v5.2.2
	github.com/go-chi/cors v1.2.1
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
)

replace github.com/cachefly/cachefly-go-sdk => /var/repo/clients/cachefly/cachefly-go-sdk
