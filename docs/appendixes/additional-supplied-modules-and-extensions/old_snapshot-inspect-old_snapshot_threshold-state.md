<a id="oldsnapshot"></a>

## old_snapshot — inspect `old_snapshot_threshold` state


 The `old_snapshot` module allows inspection of the server state that is used to implement [old_snapshot_threshold](../../server-administration/server-configuration/resource-consumption.md#guc-old-snapshot-threshold).
 <a id="oldsnapshot-functions"></a>

### Functions


`pg_old_snapshot_time_mapping(array_offset OUT int4, end_timestamp OUT timestamptz, newest_xmin OUT xid) returns setof record`
:   Returns all of the entries in the server's timestamp to XID mapping. Each entry represents the newest xmin of any snapshot taken in the corresponding minute.
