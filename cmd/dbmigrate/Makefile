APP_NAME	?= dbmigrate

build: prepare _build

.PHONY: prepare
prepare:
	-rm -rf schemas/
	# Embed cannot travel beyond parent directories, hence copy
	# migration files here.
	cp -R ../../sql/schemas .

.PHONY: clean
clean-all: clean
	-rm -rf schemas/

include ../../Makefile

