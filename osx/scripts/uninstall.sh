#!/bin/bash

set -e

SERVICE="com.ton31337.nerf.app.launchd"
SERVICE_FILE="/Library/LaunchDaemons/$SERVICE"

osascript -e 'quit app "Nerf"'

while true;
do
    if pgrep nerf-api; then
        sudo pkill -9 nerf-api
        sudo launchctl stop $SERVICE
        sudo launchctl unload -w $SERVICE_FILE
        sleep 1
    else
        break
    fi
done

sudo rm -f /Library/LaunchDaemons/$SERVICE.plist
sudo rm -rf /Library/Services/Nerf
sudo rm -rf /Applications/Nerf.app

IFS=$'\n'
for service in $(networksetup -listallnetworkservices);
do
    if [[ "$service" == *"disabled"* ]]; then
        continue
    fi
    networksetup -setdnsservers "$service" 8.8.8.8 8.8.4.4
done

echo "Uninstall completed!"

exit 0
