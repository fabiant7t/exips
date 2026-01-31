rtfm:
	cat Makefile

test:
	go test ./...

release:
	rm -rf dist/
	goreleaser release
	mv deploy/deployment.yaml.bak deploy/deployment.yaml
