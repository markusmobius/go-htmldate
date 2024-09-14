generate:
	@for name in internal/re2go/*.re; do \
		RE_IN=$$name; \
		RE_OUT=$$(echo $$name | sed 's/\.re/.go/'); \
		re2go -W -F --input-encoding utf8 --utf8 --no-generation-date -i $$RE_IN -o $$RE_OUT; \
		gofmt -w $$RE_OUT; \
	done

test-all:
	@echo "Test normal regex"
	@echo
	go test -timeout 30s ./...
	@echo
	@echo "Test re2 WASM regex"
	go test -tags re2_wasm -timeout 30s ./...
	@echo
	@echo "Test re2 cgo regex"
	go test -tags re2_cgo -timeout 30s ./...