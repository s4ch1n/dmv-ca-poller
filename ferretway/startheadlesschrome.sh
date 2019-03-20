#! /bin/bash
docker run -d --rm -p=0.0.0.0:9222:9222 --name=chrome-headless -v `pwd`/chromedata:/data alpeware/chrome-headless-stable
docker ps