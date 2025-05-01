.PHONY: test

test:
	@echo "started running test"
	go test -cover ./...
	@echo "finished running test"
