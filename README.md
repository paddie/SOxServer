SOx Server
=======================================

1. Install and setup MongoDB
----------------------------
1. Install Xcode and update to latest version (important!). Make sure that the Xcode command-line tools are installed
2. Install HomeBrew (package manager):
    1. Copy-Paste the following into Terminal.app:
        ```$ ruby <(curl -fsSk https://raw.github.com/mxcl/homebrew/go)```
        This installs the `brew` command-line tool to `usr/local/bin/brew` (which should be in the path)
    2. Call the doctor to fix permissions etc.:
                $ brew doctor
3. Use `brew` to install MongoDB
    1. Run the following in Terminal.app:
        ```$ brew install mongodb```
    2. To make sure that MongoDB launches after reboot, we need to register it with `launchctl`. This is done using a `.plist`-file which is located in homebrew's Cellar:
        ```$ cp /usr/local/Cellar/mongodb/<version>-x86_64/homebrew.mxcl.mongodb.plist /Library/LaunchAgents
        $ launchctl load -w ~/Library/LaunchAgents/homebrew.mxcl.mongodb.plist```
        Make sure that the `.plist` was registered correctly with `launchctl` by running the following:
        ```$ launchctl list | grep mongo```
        The result should look something like this:
        ```$ launchctl list | grep mongo
        141    -    homebrew.mxcl.mongo```
        This means the database will run whenever the machine is turned on.
4. Install the WebServer:
    1. Make sure that Git is installed by checking which version you have:
        ```$ git --version
        git version 1.7.4.2```
        If it turns out you don't have Git, install it using brew:
            ```$ brew install git```
    2. Next, we need to clone the reposiory from github into a local folder and launch it:
            ```$ mkdir ~/Desktop/webserver
            $ cd ~/Desktop/webserver
            $ git clone https://github.com/paddie/SOxServer.git
            $ cd SOxServer/soxify
            $ ./applications```
        The webserver should now be running and is accessible via the address:
            ```http://localhost:6060```