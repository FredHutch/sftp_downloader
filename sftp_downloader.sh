#!/bin/bash

set -e
# set -x


date
source /app/Lmod/lmod/lmod/init/profile
export MODULEPATH=/app/easybuild/modules/all
ml Python/3.6.4-foss-2016b-fh1

# VENV=$(pipenv --venv)
WORKON_HOME=$HOME/.virtualenvs

VENV_BASEDIR=$(ls -1 $WORKON_HOME | grep sftp_downloader)

# VENV=$WORKON_HOME/$VENV_BASEDIR


# FIXME unhardcode this!
VENV=/home/dtenenba/envs/sftp_downloader-o99c5K2r

$VENV/bin/python $HOME/sftp_downloader/sftp_downloader.py
