#!/bin/bash

/usr/bin/virtualenv-2.7 mypython
source mypython/bin/activate
pip install python-dxf
pip install requests
pip install bottle
pip install pyyaml
pip install hash_ring
pip install python-decouple
python thesis/docker-performance/client.py -i 0.0.0.0 -p 8084

