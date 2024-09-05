.PHONY: all
all::
	go build .

.PHONY: install
install: all
	sudo id
