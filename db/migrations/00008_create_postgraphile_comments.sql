-- +goose Up
COMMENT ON TABLE public.nodes IS E'@name NodeInfo';
COMMENT ON TABLE btc.header_cids IS E'@name BtcHeaderCids';
COMMENT ON TABLE btc.transaction_cids IS E'@name BtcTransactionCids';
COMMENT ON COLUMN public.nodes.node_id IS E'@name ChainNodeID';
COMMENT ON COLUMN btc.header_cids.node_id IS E'@name BtcNodeID';