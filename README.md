[![ci badge](https://circleci.com/gh/FredHutch/sftp_downloader.png?style=shield)](https://circleci.com/gh/FredHutch/sftp_downloader)

# SFTP Downloader Script

This script will download a file (if present) from the
SFTP server in Perú.

It is designed to be run daily via [crontab](http://www.adminschoice.com/crontab-quick-reference).

## Installation

On the system where you want this script to run, download the tool.

The tool is a single executable, called `sftp_downloader`, built to run on Linux.

To download it, change to the directory where you want the executable to live and run this command:

```
curl -O https://s3-us-west-2.amazonaws.com/fredhutch-scicomp-tools/sftp_downloader/sftp_downloader
```

FIXME - replace the rest of this....


You only have to run these steps once.

Run the following steps in your home directory
on `rhino1` (`ssh rhino1`).

```bash
git clone https://github.com/FredHutch/sftp_downloader.git
cd sftp_downloader
ml Python/3.6.4-foss-2016b-fh1
pipenv install
# create a configuration file from a template:
cp config.json.example config.json

# Now edit config.json to add the username, password, etc., of
# the SFTP server.

# Make a note of the current directory:

pwd

```

Using the `crontab -e` command (and your favorite
text-based editor), add a line like the following
to your crontab, replacing `MYUSERNAME` with your
HutchNet ID.

```
SHELL=/bin/bash
15 13 * * * /home/MYUSERNAME/sftp_downloader/sftp_downloader.sh >> /home/MYUSERNAME/sftp_downloader/sftp_downloader.log 2>&1
```

The script will now run every day at 1:15PM (1315 hours).

It will append to the log file
`/home/MYUSERNAME/sftp_downloader/sftp_downloader.log`.
If the files do not show up as expected, check this file
for error information.

## Problems

Contact Dan or `scicomp@fredhutch.org`.
