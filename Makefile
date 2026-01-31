.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test: 
	go test ./... -test.shuffle=on -test.fullpath

.PHONY: test-cover
test-cover: 
	go test -cover ./... -test.shuffle=on -test.fullpath

.PHONY: test-cover
test-cover-open: 
	go test -coverprofile=cover.out ./... -test.shuffle=on -test.fullpath
	go tool cover -html=cover.out

clean: 
	rm -f cover.out
