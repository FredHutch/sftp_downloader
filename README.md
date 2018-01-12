# SFTP Downloader Script

This script will download a file (if present) from the
SFTP server in Perú.

It is designed to be run daily via [crontab](http://www.adminschoice.com/crontab-quick-reference).

## Installation

You only have to run these steps once.

Run the following steps in your home directory
on one of the `rhino` machines (`ssh rhino`).

```bash
git clone https://github.com/FredHutch/sftp_downloader.git
cd sftp_downloader
ml Python/3.6.4-foss-2016b-fh1
pipenv install

# create a configuration file from a template:
cp config.json.example config.json

# Now edit config.json to add the username, password, etc., of
# the SFTP server.

# Make a note of the path to Python and
# the current directory:

which python
pwd

```

Using the `crontab -e` command (and your favorite
text-based editor), add a line like the following
to your crontab, replacing `MYUSERNAME` with your
HutchNet ID.

15 13 * * * /app/easybuild/software/Python/3.6.4-foss-2016b-fh1/bin/python /home/MYUSERNAME/sftp_downloader/sftp_downloader.py >> /home/MYUSERNAME/sftp_downloader/sftp_downloader.log 2>&1

The script will now run every day at 1:15PM (1315 hours).

It will append to the log file
`/home/MYUSERNAME/sftp_downloader/sftp_downloader.log`.
If the files do not show up as expected, check this file
for error information.

## Problems

Contact Dan or `scicomp@fredhutch.org`.
