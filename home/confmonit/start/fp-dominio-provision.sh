#!/bin/bash

cd /home/confmonit/v4.0/fp-dominio-provision || exit 1

chmod +x ./fp-dominio-provision 2>/dev/null

./fp-dominio-provision
