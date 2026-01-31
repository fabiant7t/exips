rtfm:
	cat Makefile

test:
	go test ./...

release:
	rm -rf dist/
	goreleaser release
	git checkout -- deploy/deployment.yaml # tags gets changed while releasing
