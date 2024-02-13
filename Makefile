.PHONY: templ

make:
	go run .

templ:
	/Users/seanburman/go/bin/templ generate

minify:
	./es-build

compile: templ
	./es-build
	tsc -p "static/assets/"
	./tailwindcss -i static/assets/stylesheets/tailwind.css -o static/assets/stylesheets/tailwind.min.css --minify
	sass static/assets/sass:static/assets/stylesheets

tsc:
	tsc -p "static/assets/" --watch

tailwind:
	./tailwindcss -i static/assets/stylesheets/tailwind.css -o static/assets/stylesheets/tailwind.min.css --watch --minify

sass:
	sass --watch static/assets/sass:static/assets/stylesheets

publish:
	git tag -s v0.1.0 -m "fncmp v0.1.0" && \
	GOPROXY=proxy.golang.org go list -m github.com/kitkitchen/fncmp@v0.1.0

lookup:
	curl https://sum.golang.org/lookup/github.com/kitkitchen/fncmp@v0.1.0