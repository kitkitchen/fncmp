.PHONY: templ_files

make:
	go run .

minify:
	./es-build

script:
	tsc -watch

tailwind:
	./tailwindcss build -o static/assets/stylesheets/tailwind.min.css

sass:
	sass --watch static/assets/sass:static/assets/stylesheets

publish:
	git tag -s v0.0.22 -m "fncmp v0.0.22" && \
	GOPROXY=proxy.golang.org go list -m github.com/kitkitchen/fncmp@v0.0.22

lookup:
	curl https://sum.golang.org/lookup/github.com/kitkitchen/fncmp@v0.0.22