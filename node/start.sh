#!/bin/bash

if [ $ENV == "development" ]; then
  go build -o /usr/bin/bui
fi

bui init

balance=`bui balance --stake -v`

if [ $? != 0 ]; then
  echo $balance
  exit $?
fi

if [ $balance == 0 ]; then
  bui register || exit 1
fi

modd -f $1
