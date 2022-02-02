PROJECT := github.com/juju/utils/v3

.PHONY: check-licence check-go check

check: check-licence check-go
    # TODO - testing this way results in a go.sum dep error
	# go test $(PROJECT)/...
	go test ./...

check-licence:
	@(grep -rFl "Licensed under the LGPLv3" .;\
		grep -rFl "MACHINE GENERATED BY THE COMMAND ABOVE; DO NOT EDIT" .;\
		grep -rFl "license that can be found in the LICENSE.ricochet2200 file" .; \
		find . -name "*.go") | sed -e 's,\./,,' | sort | uniq -u | \
		xargs -I {} echo FAIL: licence missed: {}

check-go:
	$(eval GOFMT := $(strip $(shell gofmt -l .| sed -e "s/^/ /g")))
	@(if [ x$(GOFMT) != x"" ]; then \
		echo go fmt is sad: $(GOFMT); \
		exit 1; \
	fi )
	@(go vet -all -composites=false -copylocks=false .)

# Install packages required to develop in utils and run tests.
install-dependencies: install-snap-dependencies
	@echo Installing dependencies
	@echo Installing bzr
	@sudo apt install bzr --yes
	@echo Installing zip
	@sudo apt install zip --yes

install-snap-dependencies:
## install-snap-dependencies: Install the supported snap dependencies
	@echo Installing go-1.17 snap
	@sudo snap install go --channel=1.17/stable --classic
