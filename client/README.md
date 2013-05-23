CLIENT Installer
================

1. **Preinstall**: Removes all existing files created by earlier installations
2. **Makefile**: Creates required directories (/Library/AdPeople) and copies scripts
3. **Postinstall**: Fixes permissions, ownership and registers scripts with launchctl

# Preinstall
```bash
# unload script
/bin/launchctl unload /Library/LaunchDaemons/com.adpeople.sox.plist
@sudo /bin/launchctl unload /Library/LaunchDaemons/com.adpeople.sox.plist
/bin/launchctl unload /Library/LaunchAgents/com.adpeople.sox.plist
@sudo /bin/launchctl unload /Library/LaunchAgents/com.adpeople.sox.plist

# delete old plist-file
if [ "/Library/LaunchDaemons/com.adpeople.sox.plist" ]; then 
  rm -rf "/Library/LaunchDaemons/com.adpeople.sox.plist"
fi

if [ "/Library/LaunchAgents/com.adpeople.sox.plist" ]; then 
	rm -rf "/Library/LaunchAgents/com.adpeople.sox.plist"
fi

# delete old adpeople script folder
if [ -d "/Library/AdPeople" ]; then
	rm -rf "/Library/AdPeople"
fi
´´´

# Makefile
```Bash
include luggage.make

PACKAGE_VERSION=1.0

TITLE=AdPeople_SOX
REVERSE_DOMAIN=com.adpeople.sox
PAYLOAD=\
  pack-script-preinstall\
	pack-script-postinstall\
	pack-adpeople\

prepare-files: l_Library
	# create /Library/AdPeople directory
	@sudo mkdir -p "${WORK_D}/Library/AdPeople"
	# copying sox_adpeople.py
	@sudo ${CP} sox_sophos.py ${WORK_D}/Library/AdPeople/

pack-adpeople: prepare-files l_Library_LaunchDaemons
	# fix permissions on /Library/AdPeople
	@sudo chown -R root:wheel ${WORK_D}/Library/AdPeople
	@sudo chmod -R 755 ${WORK_D}/Library/AdPeople
	@sudo chmod a+x ${WORK_D}/Library/AdPeople/sox_sophos.py

	# install daemon, fix permissions and load
	@sudo ${CP} com.adpeople.sox.plist ${WORK_D}/Library/LaunchDaemons
	@sudo chown root:wheel ${WORK_D}/Library/LaunchDaemons/com.adpeople.sox.plist
	@sudo chmod 755 ${WORK_D}/Library/LaunchDaemons/com.adpeople.sox.plist
```
# Postinstall
```bash
/bin/launchctl load -w /Library/LaunchDaemons/com.adpeople.sox.plist
```
