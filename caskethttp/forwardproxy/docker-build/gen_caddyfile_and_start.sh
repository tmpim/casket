#!/usr/bin/env bash

CASKETFILE="${CASKETFILE:-/etc/casket/Casketfile}"
ROOTDIR="${ROOTDIR:-/srv/index}"
SITE_ADDRESS="${SITE_ADDRESS:-localhost}"

generate_casketfile() {
    mkdir -p "$(dirname "${CASKETFILE}")"

    echo "${SITE_ADDRESS} {" > ${CASKETFILE}
    echo "  root $ROOTDIR" >> ${CASKETFILE}

    echo "  forwardproxy {" >> ${CASKETFILE}
    if [[ ! -z ${PROXY_USERNAME} ]]; then
        echo "    basicauth ${PROXY_USERNAME} ${PROXY_PASSWORD}" >> ${CASKETFILE}
    fi
    if [[ "${PROBE_RESISTANT}" = true ]]; then
        echo "    probe_resistance ${SECRET_LINK}" >> ${CASKETFILE}
    fi
    echo "  }" >> ${CASKETFILE}

    echo "}" >> ${CASKETFILE}
}

if [ -f "${CASKETFILE}" ]; then
    echo "Using provided Casketfile"
else
    echo "Casketfile is not provided: generating new one"
    generate_casketfile
fi

casket ${CASKET_OPTS} -conf ${CASKETFILE}
