#
#   Copyright 2011 Joe Block <jpb@ApesSeekingKnowledge.net>
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#       You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.
#

include /usr/local/share/luggage/luggage.make

TITLE=OSX_SOPHOS_CLIENT_INSTALL
REVERSE_DOMAIN=com.adpeople.sox
PAYLOAD=\
	prepare-files\
	preflight\
	copy-files\
	postflight\

prepare-files: l_Library_LaunchAgents l_Library_LaunchDaemons

preflight:
	# unload script from launchctl
	# @sudo /bin/launchctl stop com.adpeople.sox
	# /bin/launchctl stop com.adpeople.sox
	@sudo /bin/launchctl unload /Library/LaunchAgents/com.adpeople.sox.plist || true
	/bin/launchctl unload /Library/LaunchAgents/com.adpeople.sox.plist || true
	@sudo /bin/launchctl unload /Library/LaunchDaemons/com.adpeople.sox.plist || true
	/bin/launchctl unload /Library/LaunchDaemons/com.adpeople.sox.plist || true
	# cleanup old plist files from launchdaemons and launchagents
	@sudo rm -f /Library/LaunchAgents/com.adpeople.sox.plist
	@sudo rm -f /Library/LaunchDaemons/com.adpeople.sox.plist
	# remove adpeople parent dir
	@sudo rm -rf /Library/AdPeople

copy-files:
	@sudo ${CP} com.adpeople.sox.plist /Library/LaunchDaemons/com.adpeople.sox.plist
	@sudo chown root:wheel /Library/LaunchDaemons/com.adpeople.sox.plist
	@sudo chmod 644 /Library/LaunchDaemons/com.adpeople.sox.plist
	# create adpeople dir
	@sudo mkdir /Library/AdPeople
	@sudo chown root:wheel /Library/AdPeople
	# copy script and make it executable
	@sudo ${CP} sox_sophos.py /Library/AdPeople/sox_sophos.py
	@sudo chown root:wheel /Library/AdPeople
	@sudo chmod a+x /Library/AdPeople/sox_sophos.py


postflight:
	@sudo /bin/launchctl load /Library/LaunchDaemons/com.adpeople.sox.plist
	@sudo /bin/launchctl start com.adpeople.sox