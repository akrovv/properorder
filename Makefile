PLUGIN_PATH := .

build_plugin:
	go build -buildmode=plugin -o ${PLUGIN_PATH}/custom_linter.so plugin/main.go

.PHONY: build_plugin install clean

install: dependencies build_plugin clean

dependencies:
	go work init
	go work use .
	git clone https://github.com/golangci/golangci-lint.git && go work use golangci-lint
	cd golangci-lint && git checkout tags/v1.61.0 && go build -o ../golangci ./cmd/golangci-lint && cd ../

clean:
	rm -rf golangci-lint
	rm -f go.work go.work.sum
