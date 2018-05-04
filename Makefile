.PHONY : build install test test-prepare

install:
	go get github.com/golang/mock/gomock
	go get github.com/golang/mock/mockgen
	dep install

test-prepare:
	@echo "\nPreparing test........................................"
	mockgen --source client.go --destination mocks/mock_client.go
	mockgen --source decipher.go --destination mocks/mock_decipher.go

test: | test-prepare
	@echo "\nRun test........................................"
	go test

build:
	go build -o gotube .
