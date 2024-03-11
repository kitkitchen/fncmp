.PHONY: templ example

make:
	go run .

assets: templ
	tsc -p "static/assets/"
	./es-build
	./tailwindcss -i static/assets/stylesheets/tailwind.css -o static/assets/stylesheets/tailwind.min.css --minify
	sass static/assets/sass:static/assets/stylesheets

minify:
	./es-build

templ:
	/Users/seanburman/go/bin/templ generate

tsc:
	tsc -p "static/assets/" --watch

tailwind:
	./tailwindcss -i static/assets/stylesheets/tailwind.css -o static/assets/stylesheets/tailwind.min.css --watch --minify

sass:
	sass --watch static/assets/sass:static/assets/stylesheets

publish:
	git tag -s v0.3.21 -m "fncmp v0.3.21" && \
	git push --tags && \
	GOPROXY=proxy.golang.org go list -m github.com/kitkitchen/fncmp@v0.3.21 && \
	curl https://sum.golang.org/lookup/github.com/kitkitchen/fncmp@v0.3.21