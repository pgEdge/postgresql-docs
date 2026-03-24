# Installing pg_cron

Install on Red Hat, CentOS, Fedora, Amazon Linux with PostgreSQL 16 using [PGDG](https://yum.postgresql.org/repopackages/):

```bash
# Install the pg_cron extension
sudo yum install -y pg_cron_16
```

Install on Debian, Ubuntu with PostgreSQL 16 using [apt.postgresql.org](https://wiki.postgresql.org/wiki/Apt):

```bash
# Install the pg_cron extension
sudo apt-get -y install postgresql-16-cron
```

You can also install pg_cron by building it from source:

```bash
git clone https://github.com/citusdata/pg_cron.git
cd pg_cron
# Ensure pg_config is in your path, e.g.
export PATH=/usr/pgsql-16/bin:$PATH
make && sudo PATH=$PATH make install
```

