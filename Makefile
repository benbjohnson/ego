TEST=.
BENCH=.
COVERPROFILE=/tmp/c.out

# http://cloc.sourceforge.net/
cloc:
	@cloc --not-match-f='Makefile|_test.go' .

cover: fmt
	go test -coverprofile=$(COVERPROFILE) -test.run=$(TEST) .
	go tool cover -html=$(COVERPROFILE)
	rm $(COVERPROFILE)

fmt:
	@go fmt ./...

test: fmt
	@go test -v -cover -test.run=$(TEST)

.PHONY: cloc cover fmt test
