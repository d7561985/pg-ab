# Postgres

schema:
```sql
CREATE TABLE IF NOT EXISTS "balance"
(
    "accountId"      INT8   NOT NULL PRIMARY KEY,
    "balance"        float4 NOT NULL,
    "depositAllSum"  float4 NOT NULL CHECK ( "depositAllSum" >= 0),
    "depositCount"   INT    NOT NULL,
    "pincoinBalance" float4 NOT NULL CHECK ( "pincoinBalance" >= 0 ),
    "pincoinAllSum"  float4 NOT NULL CHECK ( "pincoinAllSum" >= 0 )
);

CREATE TABLE IF NOT EXISTS "journal"
(
    "id"              UUID     NOT NULL,
    "id2"             bytea       NOT NULL,
    "accountId"       INT8        NOT NULL,
    "created_at"      TIMESTAMP   NOT NULL,
    "balance"         FLOAT8   DEFAULT NULL,
    "depositAllSum"   FLOAT8   DEFAULT NULL,
    "depositCount"    INT      DEFAULT NULL,
    "pincoinBalance"  FLOAT8   DEFAULT NULL,
    "pincoinAllSum"   FLOAT8   DEFAULT NULL,
    "change"          FLOAT4   DEFAULT NULL,
    "pincoinChange"   FLOAT4   DEFAULT NULL,
    "currency"        SMALLINT DEFAULT NULL,
    "project"         VARCHAR(64) NOT NULL,
    "revert"          BOOLEAN  DEFAULT NULL,
    "transactionId"   INT8        NOT NULL,
    "transactionBson" bytea       NOT NULL,
    "transactionType" VARCHAR(36) NOT NULL
) PARTITION BY RANGE (created_at);
```

config:
```
max_connections = 600
shared_buffers = 12GB
temp_buffers = 256MB
wal_level = replica
checkpoint_timeout = 15min # range 30s-1d
max_wal_size = 100GB
min_wal_size = 1GB
checkpoint_completion_target = 0.9
wal_keep_segments = 0
seq_page_cost = 1.0 # measured on an arbitrary scale
random_page_cost = 1.3 # we use io1, NVME
effective_cache_size = 36GB
default_statistics_target = 200
```

Recommended values if PostgreSQL is on a separate server (http://help.collab.net/index.jsp?topic=/teamforge174/faq/pgsqldefaults.html)

| site-options.conf tokens	                | 8GB RAM	                   | 16GB RAM	            | 32GB RAM	       | 64GB RAM	  | 128GB RAM |
|------------------------------------------|----------------------------|----------------------|-----------------|------------|-----------|
| PGSQL_EFFECTIVE_CACHE_SIZE<br/>effective_cache_size               | 	6GB                       | 	12GB                | 	24GB           | 	48GB      | 	96GB     |
| PGSQL_SHARED_BUFFERS<br/>shared_buffers                     | 	2GB                       | 	4GB                 | 	8GB            | 	8GB       | 	8GB      |
| PGSQL_WORK_MEM<br/>work_mem| 	64 MB| 	64MB| 	64MB| 	64MB| 	64MB     |
|PGSQL_WAL_BUFFERS<br/>wal_buffers|	16 MB|	32MB|	32MB|	32MB|	32MB|
|PGSQL_MAINTENANCE_WORK_MEM<br/>maintenance_work_mem|	256MB|	615MB|	615MB|	615MB|	615MB|

NOTE: wal_buffers = -1 is OK? 

usage via sql:
```sql
WITH RECURSIVE pg_inherit(inhrelid, inhparent) AS
                   (select inhrelid, inhparent
                    FROM pg_inherits
                    UNION
                    SELECT child.inhrelid, parent.inhparent
                    FROM pg_inherit child, pg_inherits parent
                    WHERE child.inhparent = parent.inhrelid),
               pg_inherit_short AS (SELECT * FROM pg_inherit WHERE inhparent NOT IN (SELECT inhrelid FROM pg_inherit))
SELECT table_schema
     , TABLE_NAME
     , row_estimate
     , pg_size_pretty(total_bytes) AS total
     , pg_size_pretty(index_bytes) AS INDEX
     , pg_size_pretty(toast_bytes) AS toast
     , pg_size_pretty(table_bytes) AS TABLE
FROM (
         SELECT *, total_bytes-index_bytes-COALESCE(toast_bytes,0) AS table_bytes
         FROM (
                  SELECT c.oid
                       , nspname AS table_schema
                       , relname AS TABLE_NAME
                       , SUM(c.reltuples) OVER (partition BY parent) AS row_estimate
                       , SUM(pg_total_relation_size(c.oid)) OVER (partition BY parent) AS total_bytes
                       , SUM(pg_indexes_size(c.oid)) OVER (partition BY parent) AS index_bytes
                       , SUM(pg_total_relation_size(reltoastrelid)) OVER (partition BY parent) AS toast_bytes
                       , parent
                  FROM (
                           SELECT pg_class.oid
                                , reltuples
                                , relname
                                , relnamespace
                                , pg_class.reltoastrelid
                                , COALESCE(inhparent, pg_class.oid) parent
                           FROM pg_class
                                    LEFT JOIN pg_inherit_short ON inhrelid = oid
                           WHERE relkind IN ('r', 'p')
                       ) c
                           LEFT JOIN pg_namespace n ON n.oid = c.relnamespace
              ) a
         WHERE oid = parent
     ) a
ORDER BY total_bytes DESC;
```

```json
[
  {
    "oid": 16398,
    "table_schema": "public",
    "table_name": "balance",
    "row_estimate": 100000,
    "total_bytes": 9969664,
    "index_bytes": 2768896,
    "toast_bytes": null,
    "table_bytes": 7200768,
    "total": "9736 kB",
    "index": "2704 kB",
    "toast": null,
    "table": "7032 kB"
  },
  {
    "oid": 16403,
    "table_schema": "public",
    "table_name": "journal",
    "row_estimate": 33506400,
    "total_bytes": 8184045568,
    "index_bytes": 1336508416,
    "toast_bytes": 8192,
    "table_bytes": 6847528960,
    "total": "7805 MB",
    "index": "1275 MB",
    "toast": "8192 bytes",
    "table": "6530 MB"
  }
]
```

SO: 8184045568 / 33506400 = `244.253204403` average

## TEST1:
PG14 instance: r5d.2xlarge x1 with xfs nvme 300SSD
Client:  c6g.xlarge	

Disc usage:
```bash
93G/data/pg_wal
187G/data/base
279G/data
```

usage sql:
```json

[
  {
    "table_schema": "public",
    "table_name": "journal",
    "row_estimate": 686334530,
    "total": "187 GB",
    "index": "48 GB",
    "toast": "488 kB",
    "table": "138 GB"
  },
  {
    "table_schema": "public",
    "table_name": "balance",
    "row_estimate": 100012,
    "total": "11 MB",
    "index": "2800 kB",
    "toast": null,
    "table": "8304 kB"
  }
]
```

```bash
comb/sec: 19874.121340583017 duration: 30389.452879444 603963674
comb/sec: 19873.701633273387 duration: 30390.452978765 603970795
comb/sec: 19873.654040522124 duration: 30391.453618367 603989235
2022/02/07 02:56:40 worker fn ERROR: could not extend file "base/16385/16409.28": No space left on device (SQLSTATE 53100)
github.com/d7561985/pb-ab/pkg/store/postgres.(*Repo).Insert
/Users/dzmitryharupa/Documents/git/d7561985/pg-ab/pkg/store/postgres/postgres.go:103
github.com/d7561985/pb-ab/cmd/postgres.(*postgresCommand).Action.func1
/Users/dzmitryharupa/Documents/git/d7561985/pg-ab/cmd/postgres/postgres.go:81
github.com/d7561985/mongo-ab/pkg/worker.(*services).work
/Users/dzmitryharupa/go/pkg/mod/github.com/d7561985/mongo-ab@v0.0.0-20220206110900-3a9d12c987d7/pkg/worker/worker.go:84
runtime.goexit
```

## TEST2
PG14 instance: r5d.2xlarge x1 with xfs nvme 300SSD
Client:  c6g.xlarge

Test wal file 2GB only
``
max_wal_size = 2GB
``
Disc usage:
```bash
[ec2-user@ip-172-31-18-67 ~]$ sudo du -h /data
2.1G	/data/pg_wal
252G	/data/base
254G	/data
```
SQL:
```json
[
  {
    "table_schema": "public",
    "table_name": "journal",
    "row_estimate": 975796480,
    "total": "251 GB",
    "index": "65 GB",
    "toast": "488 kB",
    "table": "187 GB"
  },
  {
    "table_schema": "public",
    "table_name": "balance",
    "row_estimate": 162545,
    "total": "144 MB",
    "index": "18 MB",
    "toast": null,
    "table": "125 MB"
  }
]
```

Client last output:
```bash
comb/sec: 7682.789228416651 duration: 47269.850076942 363164295
comb/sec: 7682.626697358746 duration: 47270.850101939 363164295
comb/sec: 7682.464172506588 duration: 47271.850131064 363164295
2022/02/08 22:00:08 worker fn PANIC: could not write to file "pg_wal/xlogtemp.27512": No space left on device (SQLSTATE 53100)
github.com/d7561985/pb-ab/pkg/store/postgres.(*Repo).Insert
/Users/dzmitryharupa/Documents/git/d7561985/pg-ab/pkg/store/postgres/postgres.go:103
github.com/d7561985/pb-ab/cmd/postgres.(*postgresCommand).Action.func1
/Users/dzmitryharupa/Documents/git/d7561985/pg-ab/cmd/postgres/postgres.go:81
```

## TEST3
PG14 instance: r5d.2xlarge x1 with xfs nvme 300SSD
Client:  c6g.xlarge

Test config optimize for 64GB ram
```
effective_cache_size = 36GB
shared_buffers = 12GB
#work_mem = 4MB
#wal_buffers = -1
#maintenance_work_mem = 64MB
```
=>
```
effective_cache_size = 24GB
shared_buffers = 8GB
work_mem = 64MB
wal_buffers = 32MB
maintenance_work_mem = 64GB
```

Disc usage:
```bash
[ec2-user@ip-172-31-25-165 ~]$ sudo du -h /data
89G	/data/pg_wal
187G	/data/base
275G	/data
```

SQL:
```json
[
  {
    "table_schema": "public",
    "table_name": "journal",
    "row_estimate": 750483010,
    "total": "186 GB",
    "index": "48 GB",
    "toast": "488 kB",
    "table": "138 GB"
  },
  {
    "table_schema": "pg_catalog",
    "table_name": "pg_statistic",
    "row_estimate": 1447,
    "total": "4536 kB",
    "index": "80 kB",
    "toast": "3048 kB",
    "table": "1408 kB"
  }
]
```

Client last output:
```bash
comb/sec: 19221.93291823457 duration: 39046.752644111 750554060
comb/sec: 19221.440621132 duration: 39047.752704594 750554060
comb/sec: 19220.948355159577 duration: 39048.752753062 750554060
2022/02/09 18:54:23 worker fn ERROR: could not extend file "base/16385/16410.19": No space left on device (SQLSTATE 53100)
```

## TEST4
PG14 instance: r5d.2xlarge x1 with xfs nvme 300SSD
Client:  c6g.2xlarge
Simplify SQL schema (use 2 enum types + bytea ) + rid of obsolete fields:

```sql
CREATE TABLE IF NOT EXISTS "journal"
(
    "id"              UUID     NOT NULL,
    "id2"             bytea       NOT NULL,
    "accountId"       INT8        NOT NULL,
    "created_at"      TIMESTAMP   NOT NULL,
    "balance"         FLOAT8   DEFAULT NULL,
    "depositAllSum"   FLOAT8   DEFAULT NULL,
    "depositCount"    INT      DEFAULT NULL,
    "pincoinBalance"  FLOAT8   DEFAULT NULL,
    "pincoinAllSum"   FLOAT8   DEFAULT NULL,
    "change"          FLOAT4   DEFAULT NULL,
    "pincoinChange"   FLOAT4   DEFAULT NULL,
    "currency"        SMALLINT DEFAULT NULL,
    "project"         VARCHAR(64) NOT NULL,
    "revert"          BOOLEAN  DEFAULT NULL,
    "transactionId"   INT8        NOT NULL,
    "transactionBson" bytea       NOT NULL,
    "transactionType" VARCHAR(36) NOT NULL
) PARTITION BY RANGE (created_at);
```
=>
```sql
CREATE TABLE IF NOT EXISTS "journal"
(
    "id"              bytea       NOT NULL,
    "transactionId"   bytea       NOT NULL,
    "accountId"       INT8        NOT NULL,
    "created_at"      TIMESTAMP WITH TIME ZONE NOT NULL,
    "balance"         FLOAT8   DEFAULT NULL,
    "pincoinBalance"  FLOAT8   DEFAULT NULL,
    "change"          FLOAT4   DEFAULT NULL,
    "pincoinChange"   FLOAT4   DEFAULT NULL,
    "currency"        SMALLINT DEFAULT NULL,
    "project"         PROJECT NOT NULL,
    "type"            JOURNAL_OPERATION NOT NULL,
    "transactionType" VARCHAR(36) NOT NULL,
    "revert"          BOOLEAN  DEFAULT NULL
) PARTITION BY RANGE (created_at);
```


## Overall
| Test | Insert per sec | Element Count | actual DU <br/>(+wal file) | Size<br/>SQL Script |
|------|-------------|-----------------------|----------------------------|---------------------|
| #1   |  19873      | 686334530        |  279G                      | 187G                |
| #2   |  7682           | 975796480     |  254GB                      | 251GB               |
| #3 | 19221 | 750483010 | 275G| 186 GB|

