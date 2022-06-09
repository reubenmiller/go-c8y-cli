.PHONY: build_gh_pages

build_gh_pages:		## build github pages (used by netifly)
	cd docs/goc8ycli \
		&& npm install \
		&& npm run write-translations \
		&& npm run build
