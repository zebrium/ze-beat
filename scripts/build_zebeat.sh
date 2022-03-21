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
    export PROJECT=github.com/elastic

    export GOPATH=${GOPATH:-${HOME}/go}
    if ! $DO_UPDATE; then
        ${SUDO} rm -rf ${GOPATH}
        mkdir -p ${GOPATH}/src/${PROJECT}
    fi

    export PATH=${GOPATH}/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

    cd ${GOPATH}/src/${PROJECT}
    $DO_UPDATE || mkdir beats
    cd beats
    if ! $DO_UPDATE; then
        git init
        git remote add origin https://github.com/elastic/beats.git
        git fetch origin 8a60ca922f57a3ebb75e93d686a98edf83c766da
        git reset --hard FETCH_HEAD
    fi
    cd metricbeat
    if ! $DO_UPDATE; then
        git clone ${GIT_REPO_BASE_URL}/zebeat.git
        (cd ..; patch -p1 < metricbeat/zebeat/metricbeat_patch.diff)
        mv zebeat/zebrium module
        rm -rf modules.d
        mkdir modules.d
        cp module/zebrium/_meta/config.yml modules.d/zebrium.yml
        chmod go-w metricbeat.yml
        chmod go-w module/zebrium/module.yml
        chmod go-w modules.d/zebrium.yml
    fi
    go install

    # DOCKER IMAGE
    #
    export REGISTRY=872295030327.dkr.ecr.us-west-2.amazonaws.com/zebrium
    cd ${GOPATH}/src/${PROJECT}/elastic/beats/metricbeat/zebeat
    export ZEBEAT_COMMIT_ID=`git log -n 1 --pretty=format:"%H"`
    export DATE="$(/bin/date +%Y-%m-%d-%H-%M-%S)"
    export IMGAGE_BUILD_ID="${COMMIT_ID}_${DATE}_$(hostname)_$(whoami)"

    # BUILDID is actual release name.
    cd docker
    cp ${GOPATH}/bin/zebeat .
    docker build -t ${REGISTRY}/zebeat:${BUILDID} \
                 --label "com.zebrium.build.id=$IMAGE_BUILD_ID"                      \
                 --label "com.zebrium.build.release=$BUILDID"                        \
                 --label "com.zebrium.software.zebeat.revision=$ZEBEAT_COMMIT_ID"    \
                 --build-arg "RELEASE=$BASEOS_VERS" .

    if $DO_PUSH; then
        docker push ${REGISTRY}/zebeat:${BUILDID}
    fi
    echo export BUILDID=${BUILDID}
}

main "$@"
