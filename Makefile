VERSION=0.0.3

cmd/version.data.go: ALWAYS
	@echo "package cmd" > $@
	@echo var version = "\`$(VERSION)\`" >> $@
	@echo var commit = "\`" >> $@
	@echo `git rev-parse HEAD` >> $@
	@echo "\`" >> $@
	gofmt -w $@

.PHONY: ALWAYS
