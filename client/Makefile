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
	@sudo ${CP} update ${WORK_D}/Library/AdPeople/

pack-adpeople: prepare-files l_Library_LaunchDaemons
	# fix permissions on /Library/AdPeople
	@sudo chown -R root:wheel ${WORK_D}/Library/AdPeople
	@sudo chmod -R 755 ${WORK_D}/Library/AdPeople
	@sudo chmod a+x ${WORK_D}/Library/AdPeople/update

	# install daemon, fix permissions and load
	@sudo ${CP} com.adpeople.sox.plist ${WORK_D}/Library/LaunchDaemons
	@sudo chown root:wheel ${WORK_D}/Library/LaunchDaemons/com.adpeople.sox.plist
	@sudo chmod 755 ${WORK_D}/Library/LaunchDaemons/com.adpeople.sox.plist