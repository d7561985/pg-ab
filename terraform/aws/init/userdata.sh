sudo amazon-linux-extras install -y epel
sudo tee /etc/yum.repos.d/pgdg.repo<<EOF
[pgdg14]
name=PostgreSQL 14 for RHEL/CentOS 7 - x86_64
baseurl=http://download.postgresql.org/pub/repos/yum/14/redhat/rhel-7-x86_64
enabled=1
gpgcheck=0
EOF
sudo yum makecache
sudo yum install -y postgresql14 postgresql14-server postgresql-contrib


# mount ssd
sudo mkdir /data
sudo file -s /dev/nvme1n1
sudo lsblk -f
sudo mkfs -t xfs /dev/nvme1n1
sudo mount /dev/nvme1n1 /data

# edit
# /var/lib/pgsql/14/data/postgresql.conf
sudo chown postgres:postgres /data
LC_ALL="en_US.UTF-8"
LC_CTYPE="en_US.UTF-8"
sudo -u postgres /usr/pgsql-14/bin/initdb -D /data
# change Environment=PGDATA=/var/lib/pgsql/14/data/
sudo vim /usr/lib/systemd/system/postgresql-14.service
# copy cfg to /var/lib/pgsql/14/data/postgresql.conf
sudo systemctl daemon-reload
sudo systemctl enable --now postgresql-14
# psql
# CREATE ROLE test WITH LOGIN PASSWORD 'test';
# ALTER USER test WITH SUPERUSER;
# CREATE DATABASE db;

