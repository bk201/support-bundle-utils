#!/bin/bash -x

HOST_PATH=${HARVESTER_HOST_PATH:-/}
OUTPUT_DIR=${HARVESTER_CACHE_PATH:-/tmp/harvester-support-bundle}

[ ! -e ${OUTPUT_DIR} ] && mkdir -p $OUTPUT_DIR

NODENAME=${HARVESTER_NODENAME:-$(cat ${HOST_PATH}/etc/hostname)}
BUNDLE_DIR="${OUTPUT_DIR}/bundle-${NODENAME}"

mkdir -p ${BUNDLE_DIR}
cd ${BUNDLE_DIR}
# get some host information
cp ${HOST_PATH}/etc/hostname .

# collect logs
mkdir -p logs
cd logs
dmesg &> dmesg.log
cp ${HOST_PATH}/var/log/k3s* .
cp ${HOST_PATH}/var/log/qemu-ga.log* .
cp ${HOST_PATH}/var/log/messages* .
cp ${HOST_PATH}/var/log/console.log .

cd ${OUTPUT_DIR}
zip -r bundle.zip $(basename ${BUNDLE_DIR})
rm -rf bundle
