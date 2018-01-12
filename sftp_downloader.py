#!/usr/bin/env python3

"""
Script to download a file from an SFTP site in Perú.
This script should be set up to be run automatically each day via
crontab.
"""

import datetime
import json
import os
import sys

import pysftp

def main():
    "do the work"
    config_dir = os.path.dirname(os.path.realpath(sys.argv[0]))
    config_file = os.path.join(config_dir, "config.json")
    with open(config_file) as filehandle:
        config = json.load(filehandle)
    now = datetime.datetime.now()
    yest = now - datetime.timedelta(days=1)
    # Will filename always have 223030 in it???? FIXME if not....
    file_to_download = \
      "Reportes_diarios_acumulados-{:02d}-{:02d}-{}-223030.rar".format(yest.day,
                                                                       yest.month,
                                                                       yest.year)

    destfile = os.path.join(config['local_download_folder'], file_to_download)
    cnopts = pysftp.CnOpts()
    cnopts.hostkeys = None
    with pysftp.Connection(config['host'], username=config['user'],
                           port=config['port'], cnopts=cnopts,
                           password=config['password']) as sftp:
        # files = sftp.listdir()
        # print(files)
        if sftp.exists(file_to_download):
            print("Downloading {} to {} ...".format(file_to_download,
                                                    config['local_download_folder']))
            sftp.get(file_to_download, destfile)
            print("Done.")
        else:
            print("File {} does not exist on server!".format(file_to_download))


if __name__ == "__main__":
    main()
