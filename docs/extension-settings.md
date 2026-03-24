# Extension settings

The pg_cron extension supports the following configuration parameters:

| Setting                          | Default     | Description                                                                              |
| ---------------------------------| ----------- | ---------------------------------------------------------------------------------------- |
| `cron.database_name`             | `postgres`  | Database in which the pg_cron background worker should run.                              |
| `cron.enable_superuser_jobs`     | `on`        | Allow jobs to be scheduled as superusers.                                                |
| `cron.host`                      | `localhost` | Hostname to connect to postgres.                                                         |
| `cron.launch_active_jobs`        | `on`        | When off, disables all active jobs without requiring a server restart                    |
| `cron.log_min_messages`          | `WARNING`   | log_min_messages for the launcher bgworker.                                              |
| `cron.log_run`                   | `on`        | Log all run details in the`cron.job_run_details` table.                                  |
| `cron.log_statement`             | `on`        | Log all cron statements prior to execution.                                              |
| `cron.max_running_jobs`          | `32`        | Maximum number of jobs that can be running at the same time.                             |
| `cron.timezone`                  | `GMT`       | Timezone in which the pg_cron background worker should run.                              |
| `cron.use_background_workers`    | `off`       | Use background workers instead of client connections.                                    |

## Changing settings

To view setting configurations, run:

```sql
SELECT * FROM pg_settings WHERE name LIKE 'cron.%';
```

Setting can be changed in the postgresql.conf file or with the below command:

```sql
ALTER SYSTEM SET cron.<parameter> TO '<value>';
```

`cron.log_min_messages` and `cron.launch_active_jobs` have a [setting context](https://www.postgresql.org/docs/current/view-pg-settings.html#VIEW-PG-SETTINGS) of `sighup`. They can be finalized by executing `SELECT pg_reload_conf();`.

All the other settings have a postmaster context and only take effect after a server restart.

