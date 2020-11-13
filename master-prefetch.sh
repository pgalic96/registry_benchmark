/usr/bin/virtualenv-2.7 mypython
source mypython/bin/activate
pip install python-dxf
pip install requests
pip install bottle
pip install pyyaml
pip install hash_ring
pip install python-decouple
python thesis/docker-performance/master.py -i thesis/docker-performance/config.yaml -c warmup