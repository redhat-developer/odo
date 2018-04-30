#!/bin/bash

/usr/local/bin/wait-for-it.sh REDIS_HOST:6379

cd /app
python app.py
