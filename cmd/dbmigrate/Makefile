APP_NAME	?= dbmigrate

clean-all: clean
	-rm -rf schemas/

prepare:
	-rm -rf schemas/
	# Embed cannot travel beyond parent directories, hence copy
	# migration files here.
	cp -R ../../sql/schemas .

include ../../Makefile

