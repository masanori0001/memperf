#! /bin/bash

declare -a HOSTS
for h in $@;do
    HOSTS=("${HOSTS[@]}" $h)
done

THREAD=${THREAD:-50}
SERVERS=${SERVERS:-4}

c=0
for h in ${HOSTS[@]}; do
    c=`expr ${c} + 1`
    ssh ${h} /home/masanori.akabane/memperf/memperf -thread ${THREAD} &> ${h}-result.txt &
    if test ${c} -ge ${SERVERS}; then
       exit 0
    fi
done

