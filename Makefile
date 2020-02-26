BUILD_VERSION   := $(shell cat version)
BUILD_DATE      := $(shell date "+%F %T")
COMMIT_SHA1     := $(shell git rev-parse HEAD)

all: clean
	gox -osarch="darwin/amd64 linux/386 linux/amd64 linux/arm" \
		-output="dist/{{.Dir}}_{{.OS}}_{{.Arch}}" \
		-ldflags	"-X 'github.com/gozap/csi-nfs/cmd.Version=${BUILD_VERSION}' \
					-X 'github.com/gozap/csi-nfs/cmd.BuildDate=${BUILD_DATE}' \
					-X 'github.com/gozap/csi-nfs/cmd.CommitID=${COMMIT_SHA1}'"

clean:
	rm -rf dist

release: all
	ghr -u mritd -t ${GITHUB_TOKEN} -replace -recreate --debug ${BUILD_VERSION} dist

install:
	go install -ldflags	"-X 'github.com/gozap/csi-nfs/cmd.Version=${BUILD_VERSION}' \
               			-X 'github.com/gozap/csi-nfs/cmd.BuildDate=${BUILD_DATE}' \
               			-X 'github.com/gozap/csi-nfs/cmd.CommitID=${COMMIT_SHA1}'"

docker:
	cat Dockerfile | docker build -t gozap/csi-nfs:${BUILD_VERSION} -f - .

docker-debug:
	cat Dockerfile.debug | docker build -t gozap/csi-nfs:debug -f - .

.PHONY: all release clean install docker docker-debug

.EXPORT_ALL_VARIABLES:

GO111MODULE = on
GOPROXY = https://goproxy.io
GOSUMDB = sum.golang.google.cn
