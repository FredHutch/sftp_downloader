#!/bin/bash

# Environment Modules
export PATH=/app/bin:$PATH
source /app/Lmod/lmod/lmod/init/bash
export LMOD_PACKAGE_PATH=/app/Lmod
module use /app/easybuild/modules/all

ml Python/3.6.4-foss-2016b-fh1

python sftp_downloader.py
