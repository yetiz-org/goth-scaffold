-- maintenance_logs: stores operational maintenance records keyed by type and key.
-- Partition key: type / Clustering key: key (ASC)

CREATE TABLE IF NOT EXISTS maintenance_logs
(
    type       text,
    key        text,
    metadata   text,
    value      text,
    created_at bigint,
    updated_at bigint,
    PRIMARY KEY (type, key)
) WITH CLUSTERING ORDER BY (key ASC)
   AND comment = 'Maintenance log records';
