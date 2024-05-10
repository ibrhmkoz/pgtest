.PHONY: run

MAKEFLAGS += --silent

DOCPORT := 6060
LOCALHOST := localhost

MODNAME := github.com/ibrhmkoz/pgtest

run:
	pkgsite -http ${LOCALHOST}:${DOCPORT} & \
	while ! nc -z ${LOCALHOST} ${DOCPORT}; do sleep 0.1; done; \
	open http://${LOCALHOST}:${DOCPORT}/${MODNAME}; \
	trap 'pkill -f "pkgsite -http ${LOCALHOST}:${DOCPORT}"; exit' INT; \
	wait