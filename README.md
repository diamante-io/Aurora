# Aurora setup

To build aurora you must have go and dep in your ubuntu enivironment. To install go you need to perform  below steps and go version must be >=1.9


# Install go


- sudo curl -O https://storage.googleapis.com/golang/go1.9.1.linux-amd64.tar.gz
- sudo tar -xvf go1.9.1.linux-amd64.tar.gz
- Open vi .profile file and add following lines:
```ssh
PATH="$HOME/bin:$HOME/.local/bin:$PATH"
export GOPATH=$HOME/go
export PATH=${GOPATH}/bin:${PATH}
and save the file
```
- Open vi .bashrc and add following lines
```ssh
eval "$(direnv hook bash)"
export GOPATH=$HOME/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
```
- To refresh the files execute following commands:
```ssh
source ~/.profile
source  ~/.bashrc
```
- Check go version
```ssh
$ go version
```

# Install dep
- sudo -s
- curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
- Check dep version 
```ssh
$ dep version

```

# Install aurora

- mkdir github.com/diamantenet in /home/ubuntu/go/src
- go to /home/ubuntu/go/src/github.com/diamantenet and execute following command:
- git clone https://github.com/diamante-block/go.git
- go to /home/ubuntu/go/src/github.com/diamantenet/go and execute following command:
```
$ sudo apt-get install mercurial
```

```ssh
$ dep ensure –v
```
- go to /home/ubuntu/go/src  and execute following command:
```ssh
$ go install -ldflags "-X github.com/diamantenet/go/support/app.version=aurora-0.16.0" github.com/diamantenet/go/services/aurora/
```
- after running above command you check aurora build in <Your_dir>/go/bin folder and you can check aurora version by  following command :
```ssh
$ ./aurora version
```

# Aurora database setup
- Create a user for Diamante Net aurora database.
```
$ sudo -s
$ su – postgres
$ createuser <username> --pwprompt
$ Enter password for new role: <Enter password>
$ Enter it again: <Enter the pwd again>
```
- You need to add Aurora user. Exit from postgres and login as root user and execute following command.
```
$ exit
$ adduser <username>;
```
- To verify if user is created, execute following commands
```
$ su - postgres
$ psql
$ \du
```
- Create a blank database using following command.
```
 $ CREATE DATABASE <DB_NAME> OWNER <user created>;
```
 # Initialize aurora
 - Initialize aurora with database login as root user and Go   “go/bin” using following command
```
 $ export DATABASE_URL="postgresql://define aurora db username:define aurora db user password@localhost/define aurora database name"
```
- After that Go “go/bin” and execute following command
```
$ ./aurora db init
```
# Aurora up command
- Go “go/bin” execute following command
```
$ sudo nohup ./aurora --db-url="postgresql://define aurora db username:define aurora db user password @localhost/define aurora database name" --hcnet-core-db-url="postgresql://define Node1 database username:define Node1 database user password@localhost/define Node1 database name" --hcnet-core-url="http://localhost:11626" --network-passphrase="define Network password" --ingest="true" --per-hour-rate-limit=540000000 &
```
