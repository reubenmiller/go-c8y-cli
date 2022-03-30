.PHONY: build_gh_pages

build_gh_pages:		## build github pages (used by netifly)
	cd docs/go-c8y-cli \
		&& npm install \
		&& npm run write-translations \
		&& npm run build
