Using UploadX on macOS
=====================================

> Remember: ShareX isn't available on macOS, and it doesn't plan to be. I use Greenshot, but any screenshot program that has the utility to save to a directory will work.

A number of scripts are involved, but once it works it works pretty nicely. I utilize the `fswatch` (`brew install fswatch`) to watch the Greenshot pictures directory (`~/Pictures/Greenshot`) for updates and then pipe them off to a script.

The script it pipes them off to:

```bash
#!/bin/bash

# Config
UPLOAD_KEY="your upload key"
URL_PREFIX="https"
URL="example.com"
# End config

echo "Got new screenshot: $1"

# Attempt to upload it

jsonresult=$(curl -s -F "img=@${1}" -F "key=${UPLOAD_KEY}" ${URL_PREFIX}://${URL}/upload)
result=$(echo "$jsonresult" | jq -r ".Name")

if [ $result = "null" ]; then
	echo "error: $(echo "$jsonresult" | jq -r ".Message")"
	osascript -e 'display notification "Upload failed" with title "screenshot uploader"'
else
	osascript -e 'display notification "Upload success" with title "screenshot uploader"'
	echo "${URL_PREFIX}://${URL}/${result}" | pbcopy
fi

```

I then have a script which I run in the background on boot which just contains the fswatch command -

```bash
fswatch -0 ~/Pictures/Greenshot | while read -d "" event; do bash ~/auto-screenshot-uploader.sh "$event"; done
```

These scripts are pretty expandable, providing $YourScreenshotProgram can save to a directory you can switch directories around in the script and enjoy auto-uploading to your UploadX instance.
