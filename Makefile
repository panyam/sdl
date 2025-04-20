
NUM_LINKED_GOMODS=`cat go.mod | grep -v "^\/\/" | grep replace | wc -l | sed -e "s/ *//g"`

run: build
	cp .env /tmp 
	cp .env.dev /tmp
	LEETCOACH_ENV=dev LEETCOACH_WEB_PORT=:8080 air

checklinks:
	@if [ x"${NUM_LINKED_GOMODS}" != "x0" ]; then	\
		echo "You are trying to deploying with symlinks.  Remove them first and make sure versions exist" && false ;	\
	fi

deploy: checklinks build
	gcloud app deploy --project leetcoach --verbosity=info

build: webbuild resymlink

webbuild:
	cd web ; npm run build

resymlink:
	mkdir -p locallinks
	rm -Rf locallinks/*
	cd locallinks && ln -s ../../templar
	cd locallinks && ln -s ../../s3gen
	cd locallinks && ln -s ../../goutils
	cd locallinks && ln -s ../../oneauth
