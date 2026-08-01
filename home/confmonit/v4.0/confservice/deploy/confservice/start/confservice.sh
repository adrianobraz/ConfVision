#!/bin/bash

cd /home/confmonit/v4.0/confservice || exit 1

chmod +x ./confservice 2>/dev/null

./confservice
