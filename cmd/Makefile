
.PHONY: build server agent

build: server agent

server:
	cd server && statik -src=ui && gox -osarch linux/amd64 -output "syshealth-{{.Dir}}_{{.OS}}_{{.Arch}}"

agent:
	cd agent && gox -osarch linux/amd64 -output "syshealth-{{.Dir}}_{{.OS}}_{{.Arch}}"