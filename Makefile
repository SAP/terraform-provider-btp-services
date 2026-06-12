default: build

SETENV=
ifeq ($(OS),Windows_NT)
	SETENV=set
endif

# sed -i syntax differs: BSD (macOS) requires an explicit backup extension argument,
# GNU (Linux) treats a space-separated empty string as a filename.
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
	SED_INPLACE = sed -i ''
else
	SED_INPLACE = sed -i
endif

lefthook:
	@go install github.com/evilmartians/lefthook@latest
	lefthook install

build:
	go build -v ./...

fix:
	go fix -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

generate: docs-subcategory

docs-generate:
	go generate ./...

# docs-subcategory: adds subcategory into every generated doc.
#
# Depends on docs-generate so docs are (re)generated first; this target then
# post-processes them — never edit docs/ manually, they will be overwritten.
#
# Input: docs/subcategories.map — one "prefix:Label" line per service, e.g.:
#   cicd:SAP CICD service
#
# For each prefix, every file named <prefix>_*.md inside the four Terraform
# Registry doc directories gets its "subcategory:"  replaced
# with the corresponding label.  The label appears in the Registry UI as the
# group heading under which that resource or data source is listed.
docs-subcategory: docs-generate
	@while IFS=: read -r prefix label || [ -n "$$prefix" ]; do \
		case "$$prefix" in \#*|"") continue ;; esac; \
		for dir in docs/resources docs/data-sources docs/list-resources docs/ephemeral-resources; do \
			[ -d "$$dir" ] || continue; \
			for f in $$dir/$${prefix}_*.md; do \
				[ -f "$$f" ] || continue; \
				$(SED_INPLACE) "s|^subcategory:.*|subcategory: \"$$label\"|" "$$f"; \
			done; \
		done; \
	done < docs/subcategories.map
	@echo "Subcategories applied."

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -tags=all -timeout=900s -parallel=4 ./...

testacc:
	$(SETENV) TF_ACC=1 && go test -v -cover -tags=all -timeout 120m ./...

.PHONY: build fix install lint generate docs-generate docs-subcategory fmt test testacc