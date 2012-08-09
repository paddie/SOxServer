SOx Server
=======================================

1. Install and setup MongoDB
----------------------------
1. Install Xcode and update to latest version (important!). Make sure that the Xcode command-line tools are installed
2. Install HomeBrew (package manager):
    1. Copy-Paste the following into Terminal.app:

            $ ruby <(curl -fsSk https://raw.github.com/mxcl/homebrew/go) &&\
            brew doctor

        This installs the `brew` command-line tool to `usr/local/bin/brew` (which should be in the path) and call the `doctor` to fix permissions etc.

3. Use `brew` to install MongoDB
    1. Run the following in Terminal.app:

        $ brew install mongodb

    2. To make sure that MongoDB launches after reboot, we need to register it with `launchctl`. This is done using a `.plist`-file which is located in homebrew's Cellar:
        
            $ cp /usr/local/Cellar/mongodb/<version>-x86_64/homebrew.mxcl.mongodb.plist /Library/LaunchAgents &&\
            launchctl load -w ~/Library/LaunchAgents/homebrew.mxcl.mongodb.plist
            
        Make sure that the `.plist` was registered correctly with `launchctl` by running the following:

            $ launchctl list | grep mongo
            141    -    homebrew.mxcl.mongod

        This means the database will run whenever the machine is turned on.

Install Soxify webserver
------------------------
1. Make sure that Git is installed by checking which version you have:

        $ git --version
        git version 1.7.4.2

    If it turns out you don't have Git, install it using brew:

        $ brew install git

2. Next, we need to clone the reposiory from github into a local folder and launch it:

        $ mkdir ~/Desktop/webserver &&\
        cd ~/Desktop/webserver &&\
        git clone https://github.com/paddie/SOxServer.git &&\
        cd SOxServer/soxify &&\
        ./applications
        Connected to MongoDB on 'localhost'
        09/08/12 11:21:21: Connection from cph41madsenp - ip: 152.146.38.141
        09/08/12 11:21:39: Connection from cph41freelance_creative - ip: 152.146.210.77
        09/08/12 11:27:52: Connection from cph41mollera - ip: 152.146.210.86
        09/08/12 11:32:42: Connection from cph41taylorj - ip: 152.146.210.95
        09/08/12 11:32:56: Connection from cph41braginskym - ip: 152.146.38.117
        09/08/12 11:37:24: Connection from cph41lacoura - ip: 152.146.38.138
        09/08/12 11:53:15: Connection from cph41olsenm - ip: 152.146.210.97
        09/08/12 11:54:59: Connection from cph41jensens - ip: 152.146.210.82
        09/08/12 11:59:32: Connection from cph41valbjornu - ip: 152.146.210.68
        09/08/12 12:00:28: Connection from cph41gronegaardl - ip: 127.0.0.1
        09/08/12 12:04:21: Connection from cph41thomsenf - ip: 152.146.38.140
        09/08/12 12:04:51: Connection from cph41ornom - ip: 152.146.210.96
        09/08/12 12:08:52: Connection from cph41freelance_studio1 - ip: 152.146.210.98
        09/08/12 12:11:39: Connection from cph41borupn - ip: 152.146.38.122
        09/08/12 12:15:10: Connection from cph41poulsenm - ip: 152.146.210.89
        09/08/12 12:21:33: Connection from cph41loftrx  - ip: 152.146.38.153
        09/08/12 12:35:06: Connection from cph41mini - ip: 152.146.38.56
        09/08/12 12:44:25: Connection from cph41bendixa - ip: 152.146.38.130

    As is obvious from above, this is the output from clients connecting and updating their information.

    The website is availabe from any machine on our intranet at ip: [http://152.146.38.56:6060](http://152.146.38.56:6060).

Install Client scripts
======================
To install the client scripts simply use *Apple Remote Desktop* to distribute the installer package in

    SOxServer/client/SOx.json.Client.pkg

