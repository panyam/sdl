
run: build
	cp .env /tmp 
	cp .env.dev /tmp
	LEETCOACH_ENV=dev LEETCOACH_WEB_PORT=:8080 air

deploy: build
	gcloud app deploy --project leetcoach --verbosity=info

build: webbuild resymlink

webbuild:
	cd web ; npm run build

resymlink:
	mkdir -p locallinks
	rm -Rf locallinks/*
	cd locallinks && ln -s ../../templar
	cd locallinks && ln -s ../../s3gen
