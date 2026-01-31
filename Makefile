rtfm:
	cat Makefile

test:
	go test ./...

show_latest_tag:
	@echo Latest tag is $(shell git tag --sort=-v:refname | head -n 1); \

release: show_latest_tag
	@while true; do \
		read -p "Enter new tag (vX.Y.Z): " TAG; \
		if echo "$$TAG" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$$'; then \
			break; \
		else \
			echo "Invalid tag, try again (vX.Y.Z)"; \
		fi; \
	done; \
	rm -rf dist; \
	sed -i.bak "s/exips:.*/exips:$$TAG/g" deploy/deployment.yaml; \
	rm -f deploy/deployment.yaml.bak; \
	git add deploy/deployment.yaml; \
	git commit -m "deployment uses tag $$TAG"; \
	git tag $$TAG; \
	goreleaser release;
