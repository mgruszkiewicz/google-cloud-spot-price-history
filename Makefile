all: test vet fmt lint build

test:
	go test cmd/app/main.go

vet:
	go vet cmd/app/main.go

fmt:
	go list -f '{{.Dir}}' ./... | grep -v /vendor/ | xargs -L1 gofmt -l
	test -z $$(go list -f '{{.Dir}}' ./... | grep -v /vendor/ | xargs -L1 gofmt -l)

lint:
	go list ./... | grep -v /vendor/ | xargs -L1 golint -set_exit_status

build:
	go build -o bin/app ./cmd/app

run: build
	git clone https://github.com/Cyclenerd/google-cloud-pricing-cost-calculator || true
	cd google-cloud-pricing-cost-calculator; ../scripts/extract_git_history.sh pricing.yml || true
	./bin/app -data /tmp/pricing-data -dbpath /tmp/history.sqlite3
	