#!/bin/bash

src_dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
plist=io.dms3.dms3fs-daemon.plist
dest_dir="$HOME/Library/LaunchAgents"
DMS3FS_PATH="${DMS3FS_PATH:-$HOME/.dms3-fs}"
escaped_dms3fs_path=$(echo $DMS3FS_PATH|sed 's/\//\\\//g')

DMS3FS_BIN=$(which dms3fs || echo dms3fs)
escaped_dms3fs_bin=$(echo $DMS3FS_BIN|sed 's/\//\\\//g')

mkdir -p "$dest_dir"

sed -e 's/{{DMS3FS_PATH}}/'"$escaped_dms3fs_path"'/g' \
  -e 's/{{DMS3FS_BIN}}/'"$escaped_dms3fs_bin"'/g' \
  "$src_dir/$plist" \
  > "$dest_dir/$plist"

launchctl list | grep dms3fs-daemon >/dev/null
if [ $? ]; then
  echo Unloading existing dms3fs-daemon
  launchctl unload "$dest_dir/$plist"
fi

echo Loading dms3fs-daemon
if (( `sw_vers -productVersion | cut -d'.' -f2` > 9 )); then
  sudo chown root "$dest_dir/$plist"
  sudo launchctl bootstrap system "$dest_dir/$plist"
else
  launchctl load "$dest_dir/$plist"
fi
launchctl list | grep dms3fs-daemon
