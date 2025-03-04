
run: build
	cp .env /tmp 
	cp .env.dev /tmp
	LEETCOACH_ENV=dev LEETCOACH_WEB_PORT=:8080 air

deploy: build
	gcloud app deploy --project leetcoach --verbosity=info

build:
	cd web ; npm run build
