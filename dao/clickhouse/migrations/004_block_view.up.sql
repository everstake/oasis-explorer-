--CREATE VIEW IF NOT EXISTS block_row_view AS
--select * from blocks ANY LEFT JOIN
--              (select blk_lvl, sum(tx_gas_price), sum(tx_fee), count()
--from transactions group by blk_lvl) as s USING blk_lvl ORDER BY blk_lvl DESC
-- HERE IS FULL SCAN OF txs table, cause it ordered by hash, not block height
-- SO we have create Materialized view to store ordered data  

CREATE MATERIALIZED VIEW block_row_view
(
    blk_lvl UInt64,
    blk_created_at DateTime,
    blk_hash FixedString(64),
    blk_proposer_address FixedString(40),
    blk_validator_hash FixedString(64),
    blk_epoch UInt64,
    tx_gas_price UInt64,
    tx_fee UInt64,
    count UInt32
)
ENGINE = ReplacingMergeTree
PARTITION BY intDiv(blk_lvl,500000)
ORDER BY (blk_lvl)
SETTINGS index_granularity = 8192 POPULATE
 AS
(

SELECT *
FROM blocks
ANY LEFT JOIN
(
    SELECT
        blk_lvl,
        sum(tx_gas_price),
        sum(tx_fee),
        count()
    FROM transactions
    GROUP BY blk_lvl
) AS s USING (blk_lvl)
ORDER BY blk_lvl DESC   
);

CREATE VIEW IF NOT EXISTS blocks_sig_count AS select blk_lvl, count() sig_count FROM block_signatures GROUP BY blk_lvl;
