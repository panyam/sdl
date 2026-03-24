
buf: ensureenv clean
	buf dep update
	buf generate
	goimports -w `find ../gen ../web/gen | grep "\.go"`

clean:
	rm -Rf ../gen
	rm -Rf ../web/gen
	rm -f buf.lock

cleanall: clean remove-proto-symlinks
	rm -f buf.yaml
	rm -f buf.gen.yaml

setupdev: cleanall symlink-protos
	ln -s buf.gen.yaml.dev buf.gen.yaml
	ln -s buf.yaml.dev buf.yaml

setupprod: cleanall remove-proto-symlinks
	ln -s buf.gen.yaml.prod buf.gen.yaml
	ln -s buf.yaml.prod buf.yaml

ensureenv:
	@test -f buf.yaml && test -f buf.gen.yaml || echo "buf.yaml does not exist.  Run 'make setupdev' or 'make setupprod' to setup your environment..."

# Create symlink to wasmjs annotations for development
symlink-protos: remove-proto-symlinks
	echo "Creating symlink for development..."
	# ln -s ../../../engine/protos/turnengine turnengine

# Remove symlink (for switching back to production mode)
remove-proto-symlinks:
	echo "Removing proto symlink..."
	# rm -Rf wasmjs turnengine
