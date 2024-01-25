make:
	go run .

minify:
	./es-build

script:
	tsc -watch

sass:
	sass --watch static/assets/sass:static/assets/stylesheets

publish:
	git tag -s v0.0.1 -m "fncmp v0.0.1" && \
	GOPROXY=proxy.golang.org go list -m github.com/kitkitchen/fcmp@v0.0.1

lookup:
	curl https://sum.golang.org/lookup/github.com/kitkitchen/fncmp@v0.0.2