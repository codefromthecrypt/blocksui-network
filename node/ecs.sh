#!/bin/bash

bui init

balance=`bui balance --stake -v`

if [ $balance == 0 ]; then
  bui register || exit 1
fi

bui node
