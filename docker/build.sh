#!/bin/bash

Version=v18.06.14

echo start build Version: $Version

chmod 777 filtetest

docker build -t reg.libratone.com:5000/filewatch:$Version .

#docker push reg.birdytone.com/postoffice:$Version
