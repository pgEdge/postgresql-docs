# Upgrade Notes

## 0.6.0

### Postgres 12

If upgrading with Postgres 12, remove this line from `sql/vector--0.5.1--0.6.0.sql`:

```sql
ALTER TYPE vector SET (STORAGE = external);
```

Then run `make install` and `ALTER EXTENSION vector UPDATE;`.

### Docker

The Docker image is now published in the `pgvector` org, and there are tags for each supported version of Postgres (rather than a `latest` tag).

```sh
docker pull pgvector/pgvector:pg16
# or
docker pull pgvector/pgvector:0.6.0-pg16
```

Also, if you’ve increased `maintenance_work_mem`, make sure `--shm-size` is at least that size to avoid an error with parallel HNSW index builds.

```sh
docker run --shm-size=1g ...
```

