
test:
	@go test --cover --timeout 5s ./...

install:
	@go install \
		-ldflags="-X github.com/airplanedev/cli/pkg/analytics.segmentWriteKey=${SEGMENT_WRITE_KEY} -X github.com/airplanedev/cli/pkg/analytics.sentryDSN=${SENTRY_DSN}" \
		./cmd/airplane
