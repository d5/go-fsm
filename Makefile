dep:
	dep ensure

vet:
	go vet ./...

lint:
	golint -set_exit_status ./...

test: dep vet lint
	go test -race -cover ./...

fmt:
	go fmt ./...
