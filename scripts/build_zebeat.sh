#!/bin/bash

# This script builds mwsd container image

set -x
set -e

PROG=${0##*/}

usage() {
    echo "Usage: $PROG [-r <build id>] [-g <go path>] [-b <branch> ] [-B <baseos_vers>] [-o] [-u] [-P] [TAG]" 1>&2
    exit 1
}

# SET TAG
function set_tag {
        if [ "${TAG}" == "" ]; then
                git checkout ${BRANCH}
        else
                git checkout tags/${TAG}
        fi
}

main() {
    local DO_PUSH=true
    local DO_UPDATE=false
    local SUDO=${SUDO:-sudo}
    export BUILDID=`date '+%Y%m%d%H%M%S'`
    export BRANCH=master
    while getopts "b:B:g:oPr:ua" opt; do
          case $opt in
              b) BRANCH=$OPTARG ;;
              B) BASEOS_VERS=$OPTARG ;;
              g) GOPATH=$OPTARG ;;
              P) DO_PUSH=false ;;
              r) export BUILDID=$OPTARG ;;
              u) DO_UPDATE=true ;;
              *) usage ;;
          esac
    done

    shift $((OPTIND -1))
    export TAG=$1

    local BASEOS_VERS="${BASEOS_VERS:-$BUILDID}"
    export GIT_REPO_BASE_URL=${GIT_REPO_BASE_URL:-git@bitbucket.org:zebrium}

    # PROJECT DIR
    #
    export PROJECT=zebrium

    export GOPATH=${GOPATH:-${HOME}/go}
    if ! $DO_UPDATE; then
        ${SUDO} rm -rf ${GOPATH}
        mkdir -p ${GOPATH}/src/${PROJECT}
    fi

    export PATH=${GOPATH}/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

    cd ${GOPATH}/src/${PROJECT}
    $DO_UPDATE || mkdir beats
    cd beats
    $DO_UPDATE || git init
    $DO_UPDATE || git remote add origin https://github.com/elastic/beats.git
    $DO_UPDATE || git fetch origin 8a60ca922f57a3ebb75e93d686a98edf83c766da
    $DO_UPDATE || git clone ${GIT_REPO_BASE_URL}/zebeat.git
    $DO_UPDATE || mv zebeat module
    go install
    exit

    # MWSD
    #
    cd ${GOPATH}/src/${PROJECT}
    cd mwsd
    git pull
    set_tag
    export MWSD_COMMIT_ID=`/auto/share/bin/git_commit_id.sh`
    local IMAGE_BUILD_ID=`/auto/share/bin/build_id.sh`
    cat > link/software_version.go <<EOF
package link
const CSoftwareCommits = "${GO_RAML_COMMIT_ID} ${MWSD_COMMIT_ID}"
const CSoftwareRelease = "${BUILDID}"
EOF
    mo-raml server -l go --dir mwsd --no-apidocs --import-path ${PROJECT}/mwsd/mwsd --link-path ${PROJECT}/mwsd/link --ramlfile mwsd.raml
    (cd mwsd; export GO111MODULE=off;go install)
    (cd mwinit; export GO111MODULE=off;go install)

    # DOCKER IMAGE
    #
    export REGISTRY=872295030327.dkr.ecr.us-west-2.amazonaws.com/zebrium
    cd ${GOPATH}/src/${PROJECT}/mwsd/docker
    cp ${GOPATH}/bin/mwsd .
    cp ${GOPATH}/bin/mwinit .

    # BUILDID is actual release name.
    docker build -t ${REGISTRY}/mwsd:${BUILDID} \
                 --label "com.zebrium.build.id=$IMAGE_BUILD_ID"                      \
                 --label "com.zebrium.build.release=$BUILDID"                        \
                 --label "com.zebrium.software.mo-raml.revision=$GO_RAML_COMMIT_ID"  \
                 --label "com.zebrium.software.mwsd.revision=$MWSD_COMMIT_ID"        \
                 --build-arg "RELEASE=$BASEOS_VERS" .

    if $DO_PUSH; then
        docker push ${REGISTRY}/mwsd:${BUILDID}
    fi
    echo export BUILDID=${BUILDID}
}

main "$@"
