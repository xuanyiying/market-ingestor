-- Migration: TimescaleDB Hypertables Optimization

-- 1. Convert klines to hypertable
-- We partition by 'time' and additionally by 'symbol' for better query performance
SELECT create_hypertable('klines', 'time', partitioning_column => 'symbol', number_partitions => 4, if_not_exists => TRUE);

-- 2. Convert trades to hypertable (if it exists)
-- SELECT create_hypertable('trades', 'timestamp', partitioning_column => 'symbol', number_partitions => 4, if_not_exists => TRUE);

-- 3. Set chunk time interval (e.g., 1 day for klines if data volume is high)
SELECT set_chunk_time_interval('klines', INTERVAL '1 day');

-- 4. Future-proofing: Compression policy
-- ALTER TABLE klines SET (
--   timescaledb.compress,
--   timescaledb.compress_segmentby = 'symbol',
--   timescaledb.compress_orderby = 'time'
-- );
-- SELECT add_compression_policy('klines', INTERVAL '7 days');
