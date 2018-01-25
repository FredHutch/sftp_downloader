[![ci badge](https://circleci.com/gh/FredHutch/sftp_downloader.png?style=shield)](https://circleci.com/gh/FredHutch/sftp_downloader)

# SFTP Downloader Script

This script will download a file (if present) from the
SFTP server in Peru.

It is designed to be run daily via [crontab](http://www.adminschoice.com/crontab-quick-reference).

## Installation / Upgrading

On the system where you want this script to run, download the tool.

The tool is a single executable, called `sftp_downloader`, built to run on Linux.

To download it, change to the directory where you want the executable to live and run this command:

```
curl -O https://s3-us-west-2.amazonaws.com/fredhutch-scicomp-tools/sftp_downloader/sftp_downloader
```


You only have to run these steps once.

Download the example JSON file:

```
curl https://raw.githubusercontent.com/FredHutch/sftp_downloader/master/config.json.example > config.json
chmod 0600 config.json
```

Edit the `config.json` file. Values should be as follows:

* `host` - IP address or hostname of SFTP server
* `port` - port number (not in quotes) of SFTP server
* `user` - username to log into SFTP server
* `password` - password to log into SFTP server
* `rar_decryption_password` - password to decrypt RAR file
* `local_download_folder` - folder in which to download/extract RAR file
* `postprocessing_command` - a command to run after downloading and extraction is complete
  (command will be run in the directory where the files have been archived).



Using the `crontab -e` command (and your favorite
text-based editor), add a line like the following
to your crontab, replacing `MYUSERNAME` with your
HutchNet ID.

```
SHELL=/bin/bash
15 01 * * 2-6 /home/MYUSERNAME/sftp_downloader /home/MYUSERNAME/config.json >> /home/MYUSERNAME/sftp_downloader.log 2>&1
```

The script will now run every day at 1:15AM (0115 hours), Tuesday through Saturday
(the day after each weekday).

It will append to the log file
`/home/MYUSERNAME/sftp_downloader.log`.
If the files do not show up as expected, check this file
for error information.

## What the script does

When invoked in a crontab, as above, the script will do the following:

* Connect to the SFTP server
* Download yesterday's file (to download the file from a different day, see the next section).
* Unarchive the RAR file.
* Run a post-processing script (based on a command-line that you supply in the `config.json` file)
  in the directory where the files have been unarchived.


## Running the script manually

You can run the script manually. You may want to do this, for example, if you
have a backlog of RAR files to download from before `sftp_downloader` was available.

To run the script manually, just add a date to the command line. For example, to
download the files from January 5th, 2018, do this:

```
./sftp_downloader config.json 2018-01-05
```

## Problems

Contact Dan or `scicomp@fredhutch.org`, or
[file an issue](https://github.com/FredHutch/sftp_downloader/issues/new).
