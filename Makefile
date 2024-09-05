.PHONY: all
all::
	go build .

.PHONY: install
install:
	cat hemlock-sendmsg.service.template \
		| sed -e "s,{{install_path}},$$PWD," \
		> /etc/systemd/system/hemlock-sendmsg.service
	systemctl daemon-reload
	systemctl enable hemlock-sendmsg
