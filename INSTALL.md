If directory does not yet exist,

    $ sudo mkdir /usr/local/database-backups
    $ cd $_
    $ sudo chown $(whoami): .
    $ git clone https://github.com/jbaikge/database-backups.git ./

Pulling down changes

    $ cd /usr/local/database-backups
    $ git pull

Compiling and installing

    $ cd /usr/local/database-backups
    $ go build -v ./cmd/database-backup
    $ go build -v ./cmd/database-backup-api
    $ sudo mv database-backup database-backup-api /usr/local/bin
    $ sudo chown root: /usr/local/bin/database-backup*
