all: test vet fmt lint build

test:
	go test cmd/dataprocessing/main.go

vet:
	go vet cmd/dataprocessing/main.go

fmt:
	go list -f '{{.Dir}}' ./... | grep -v /vendor/ | xargs -L1 gofmt -l
	test -z $$(go list -f '{{.Dir}}' ./... | grep -v /vendor/ | xargs -L1 gofmt -l)

lint:
	go list ./... | grep -v /vendor/ | xargs -L1 golint -set_exit_status

build:
	go build -ldflags="-w -s" -o bin/dataprocessing ./cmd/dataprocessing
	go build -ldflags="-w -s" -o bin/api ./cmd/api

collect-pricing-data:
	git clone https://github.com/Cyclenerd/google-cloud-pricing-cost-calculator || true
	cd google-cloud-pricing-cost-calculator; ../scripts/extract_git_history.sh pricing.yml || true

run: build collect-pricing-data
	./bin/dataprocessing -data /tmp/pricing-data -dbpath /tmp/history.sqlite3
	