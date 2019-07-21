#!/usr/bin/env zsh

file=$1
duration=$2

expire check $file
check_code=$?
if [ $check_code = 2 ]; then
    expire new --reset-on-touch --duration $duration $file
    if [ -e $file ]; then
        rm $file
        touch $file
    fi
    exit 0
elif [ $check_code = 1 ]; then
    expire renew $file
    if [ -e $file ]; then
        rm $file
        touch $file
    fi
    exit 0
else # 0
    exit 1
fi
