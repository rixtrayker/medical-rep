include make/common.mk
include make/help.mk
include make/build.mk
include make/migrate.mk
include make/docker.mk
include make/dev.mk
include make/test.mk

.DEFAULT_GOAL := help
