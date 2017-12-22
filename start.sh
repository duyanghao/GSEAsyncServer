#!/bin/bash

./GSEAsyncServer -log_dir=./logs -logtostderr=true -v 6 -config-file-path="./Configuration/config.yml" -work-channel-len=10 -process-max-timeout=30 1>>./GSEAsyncServer.log 2>&1 &
