1. Install git versioning system


        $ sudo apt-get install git

2. Install bazaar versioning system


        $ sudo apt-get install bzr

3. Install mercurial versioning system


        $ sudo apt-get install mercurial

4. Configure APT (Package Management System)


        $ sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv 7F0CEB10
        $ echo 'deb http://downloads-distro.mongodb.org/repo/ubuntu-upstart dist 10gen' | sudo tee /etc/apt/sources.list.d/mongodb.list

5. Install MongoDB

  We need to install the specific 2.4.8 version

      $ sudo apt-get install mongodb-10gen=2.4.8

  Make sure that the mongodb serivec is running:

      $ sudo service mongodb start
      start: Job is already running: mongodb

6. Install Go Version 1.2

  Find the the go1.2-linux-amd64.tar.gz in the package repo [https://code.google.com/p/go/downloads/list](https://code.google.com/p/go/downloads/list) (amd64 is the 64-bit version). We need to unzip the compressed `go` directory to `/usr/local`. Basically follow this [guide](http://www.giantflyingsaucer.com/blog/?p=4649):

      $ cd <directory where go1.2-linux-amd64.tar.gz is located>
      $ sudo tar -C /usr/local -xzf go1.1.2.linux-amd64.tar.gz

  Next we need to make sure that the `go` binary is in the system `$PATH` by appending these two lines to the  ~/.bashrc ($HOME/.bashrc) file:

      export GOPATH="$HOME/golang/"
      export PATH=$PATH:/usr/local/go/bin

  Now reload the .bashrc:

      $ source ~/.bashrc

  Verify that go1.2 is the current version:

      $ go version
      go version go1.2 linux/amd64

  `go get` the soxify application from github.com:

      $ go get github.com/paddie/SOxServer/soxify

  Execute soxify:

      $ cd $GOPATH/github.com/paddie/SOxServer/soxify
      $ go run *.go
      Trying to connect to  localhost
      Connected to MongoDB on 'localhost'

7. Verify that the server is available on `localhost:6060`

  The database is empty, but the server should appear in a scraped form.

8. Restore data to MongoDB
