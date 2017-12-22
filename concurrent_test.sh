#!/bin/bash

cnt=0
while [ $cnt -lt $1 ]
do
	cnt=$((cnt+1))
	echo "try $cnt ..."
	curl -d '{"msg":"test message '$cnt'"}' http://x.x.x.x/task/v1/message
done
