PLUGIN_PATH := .

build_plugin:
	go build -buildmode=plugin -o ${PLUGIN_PATH}/tppcs.so plugin/main.go
	
.PHONY: build_plugin
