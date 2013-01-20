SOx Server
=======================================


1. Setup a static IP on the Mac Mini
------------------------------------
The clients all try and connect to the server on IP

    152.146.38.56

Therefore, the server must have that IP statically defined in its network settings.

2. Install and setup MongoDB
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

    2. To make sure that MongoDB launches after reboot, we need to register it with `launchctl`. This is done using a `.plist`-file which is located in homebrew's Cellar (this assumes MongoDB v. 1.8.0-x86_64 is installed, given a different version, the path would obviously be different):
        
            $ cp /usr/local/Cellar/mongodb/1.8.0-x86_64/homebrew.mxcl.mongodb.plist /Library/LaunchAgents &&\
            launchctl load -w /Library/LaunchAgents/homebrew.mxcl.mongodb.plist
            
        Make sure that the `.plist` was registered correctly with `launchctl` by running the following:

            $ launchctl list | grep mongo
            141    -    homebrew.mxcl.mongod

        This means the database will run whenever the machine is turned on.

3. Install Soxify webserver
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

    This launches the server which tries to connect to the mongodb database on ```localhost``` and starts accepting connections from clients:

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

4. Install Client scripts
-------------------------
To install the client scripts simply use *Apple Remote Desktop* to distribute the installer package in

    SOxServer/client/SOx.json.Client.pkg

5. Edit Client Installer
========================
If you ever need to edit the client scripts or where/how the scripts are installed, you need to create a new installation package. Launch ```PackageManager.app``` and follow the below guide:

1. Fill out the prompt as illustrated below and click OK:

    ![Create new installation package](http://imgur.com/fIX93.png)

2. Choose the ```Configuration```:

    ![Make sure 'Installation Destination' is set to 'System Volume' and that you use 'Easy Install Only'](http://imgur.com/tWb4P.png)

3. Drag the ```SOxServer/client/package_root/Library/AdPeople/sox_sophos.py``` onto the open application

4. Fill out the 'Configuration' as seen below:
    
    ![Make sure the field 'Allow Custom Location' is not checked](http://imgur.com/agzBF.png)

5. And edit the files permissions in 'Contend':
    
    !['Wheel' is the launchd schedular's group](http://i.imgur.com/8xbvR.png)

6. Now drag the the ```SOxServer/client/package_root/Library/LaunchAgents/com.adpeople.sox.plist``` onto the open application and edit the 'Configuration' to match the below picture:
    
    ![Restart action needs to be 'None' for all files.](http://i.imgur.com/E62nt.png)

7. Now edit the 'Contend':
    
    ![Only let the owner/admin edit the file. No one needs permission to execute it.](http://i.imgur.com/WsyP7.png)

8. Lastly, to unload the old scripts from the schedular and register and replace them with the new ones; open the 'Scripts' menu and drag the files ```SOxServer/client/Resources/PostFlight``` and ```SOxServer/client/Resources/PreFlight``` to their respective fields:
    
    !['PreFlight' to preinstall, 'PostFlight' to postinstall](http://i.imgur.com/73T0t.png)

9. Now hit 'Build and Run' and watch for any compilation errors
    * ignore any errors wrt. permissions on the two files (but do go back and make sure you've set the correct ones on each of the files the time around)

10. If the installation was a success, check the output from the server console to see if the machine successfully posted its data to the Soxify server:
    
    ![compare time-stamps to see that the client reached the server](http://i.imgur.com/WQr3Z.png)

11. Go make yourself a cup of coffee, you've earned it!
