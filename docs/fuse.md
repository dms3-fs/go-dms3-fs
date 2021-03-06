# FUSE

`go-dms3fs` makes it possible to mount `/dms3fs` and `/dms3ns` namespaces in your OS,
allowing arbitrary apps access to DMS3FS.

## Install FUSE

You will need to install and configure fuse before you can mount DMS3FS

#### Linux

Note: while this guide should work for most distributions, you may need to refer
to your distribution manual to get things working.

Install `fuse` with your favorite package manager:
```
sudo apt-get install fuse
```

Add the user which will be running DMS3FS daemon to the `fuse` group:
```sh
sudo usermod -a -G fuse <username>
```

Restart user session, if active, for the change to apply, either by restarting
ssh connection or by re-logging to the system.

#### Mac OSX -- OSXFUSE

It has been discovered that versions of `osxfuse` prior to `2.7.0` will cause a
kernel panic. For everyone's sake, please upgrade (latest at time of writing is
`2.7.4`). The installer can be found at https://osxfuse.github.io/. There is
also a homebrew formula (`brew install osxfuse`) but users report best results
installing from the official OSXFUSE installer package.

Note that `dms3fs` attempts an automatic version check on `osxfuse` to prevent you
from shooting yourself in the foot if you have pre `2.7.0`. Since checking the
OSXFUSE version [is more complicated than it should be], running `dms3fs mount`
may require you to install another binary:

```sh
go get github.com/jbenet/go-fuse-version/fuse-version
```

If you run into any problems installing FUSE or mounting DMS3FS, hop on IRC and
speak with us, or if you figure something new out, please add to this document!

## Prepare mountpoints

By default dms3fs uses `/dms3fs` and `/dms3ns` directories for mounting, this can be
changed in config. You will have to create the `/dms3fs` and `/dms3ns` directories
explicitly. Note that modifying root requires sudo permissions.

```sh
# make the directories
sudo mkdir /dms3fs
sudo mkdir /dms3ns

# chown them so dms3fs can use them without root permissions
sudo chown <username> /dms3fs
sudo chown <username> /dms3ns
```

Depending on whether you are using OSX or Linux, follow the proceeding instructions.

## Mounting DMS3FS

```sh
dms3fs daemon --mount
```

If you wish to allow other users to use the mount points, edit `/etc/fuse.conf`
to enable non-root users, i.e.:
```sh
# /etc/fuse.conf - Configuration file for Filesystem in Userspace (FUSE)

# Set the maximum number of FUSE mounts allowed to non-root users.
# The default is 1000.
#mount_max = 1000

# Allow non-root users to specify the allow_other or allow_root mount options.
user_allow_other
```

Next set `Mounts.FuseAllowOther` config option to `true`:
```sh
dms3fs config --json Mounts.FuseAllowOther true
dms3fs daemon --mount
```

## Troubleshooting

#### `Permission denied` or `fusermount: user has no write access to mountpoint` error in Linux

Verify that the config file can be read by your user:
```sh
sudo ls -l /etc/fuse.conf
-rw-r----- 1 root fuse 216 Jan  2  2013 /etc/fuse.conf
```
In most distributions, the group named `fuse` will be created during fuse
installation. You can check this with:

```sh
sudo grep -q fuse /etc/group && echo fuse_group_present || echo fuse_group_missing
```

If the group is present, just add your regular user to the `fuse` group:
```sh
sudo usermod -G fuse -a <username>
```

If the group didn't exist, create `fuse` group (add your regular user to it) and
set necessary permissions, for example:
```sh
sudo chgrp fuse /etc/fuse.conf
sudo chmod g+r  /etc/fuse.conf
```
<!--
TODO: udev rules for /dev/fuse?
-->

Note that the use of `fuse` group is optional and may depend on your operating
system. It is okay to use a different group as long as proper permissions are
set for user running `dms3fs mount` command.

#### Mount command crashes and mountpoint gets stuck

```
sudo umount /dms3fs
sudo umount /dms3ns
```

If you manage to mount on other systems (or followed an alternative path to one
above), please contribute to these docs :D
