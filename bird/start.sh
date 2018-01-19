!/bin/bash

docker run -it --name bird -p 8179:179 -v ~/go/src/github.com/transitorykris/kbgp/bird/bird.conf:/etc/bird.conf alectolytic/bird -f -D /dev/stdout
