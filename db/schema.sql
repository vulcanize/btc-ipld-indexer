--
-- PostgreSQL database dump
--

-- Dumped from database version 12.1
-- Dumped by pg_dump version 12.1

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: btc; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA btc;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: header_cids; Type: TABLE; Schema: btc; Owner: -
--

CREATE TABLE btc.header_cids (
    id integer NOT NULL,
    block_number bigint NOT NULL,
    block_hash character varying(66) NOT NULL,
    parent_hash character varying(66) NOT NULL,
    cid text NOT NULL,
    mh_key text NOT NULL,
    "timestamp" numeric NOT NULL,
    bits bigint NOT NULL,
    node_id integer NOT NULL,
    times_validated integer DEFAULT 1 NOT NULL
);


--
-- Name: TABLE header_cids; Type: COMMENT; Schema: btc; Owner: -
--

COMMENT ON TABLE btc.header_cids IS '@name BtcHeaderCids';


--
-- Name: COLUMN header_cids.node_id; Type: COMMENT; Schema: btc; Owner: -
--

COMMENT ON COLUMN btc.header_cids.node_id IS '@name BtcNodeID';


--
-- Name: header_cids_id_seq; Type: SEQUENCE; Schema: btc; Owner: -
--

CREATE SEQUENCE btc.header_cids_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: header_cids_id_seq; Type: SEQUENCE OWNED BY; Schema: btc; Owner: -
--

ALTER SEQUENCE btc.header_cids_id_seq OWNED BY btc.header_cids.id;


--
-- Name: transaction_cids; Type: TABLE; Schema: btc; Owner: -
--

CREATE TABLE btc.transaction_cids (
    id integer NOT NULL,
    header_id integer NOT NULL,
    index integer NOT NULL,
    tx_hash character varying(66) NOT NULL,
    cid text NOT NULL,
    mh_key text NOT NULL,
    segwit boolean NOT NULL,
    witness_hash character varying(66)
);


--
-- Name: TABLE transaction_cids; Type: COMMENT; Schema: btc; Owner: -
--

COMMENT ON TABLE btc.transaction_cids IS '@name BtcTransactionCids';


--
-- Name: transaction_cids_id_seq; Type: SEQUENCE; Schema: btc; Owner: -
--

CREATE SEQUENCE btc.transaction_cids_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: transaction_cids_id_seq; Type: SEQUENCE OWNED BY; Schema: btc; Owner: -
--

ALTER SEQUENCE btc.transaction_cids_id_seq OWNED BY btc.transaction_cids.id;


--
-- Name: tx_inputs; Type: TABLE; Schema: btc; Owner: -
--

CREATE TABLE btc.tx_inputs (
    id integer NOT NULL,
    tx_id integer NOT NULL,
    index integer NOT NULL,
    witness character varying[],
    sig_script bytea NOT NULL,
    outpoint_tx_hash character varying(66) NOT NULL,
    outpoint_index numeric NOT NULL
);


--
-- Name: tx_inputs_id_seq; Type: SEQUENCE; Schema: btc; Owner: -
--

CREATE SEQUENCE btc.tx_inputs_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tx_inputs_id_seq; Type: SEQUENCE OWNED BY; Schema: btc; Owner: -
--

ALTER SEQUENCE btc.tx_inputs_id_seq OWNED BY btc.tx_inputs.id;


--
-- Name: tx_outputs; Type: TABLE; Schema: btc; Owner: -
--

CREATE TABLE btc.tx_outputs (
    id integer NOT NULL,
    tx_id integer NOT NULL,
    index integer NOT NULL,
    value bigint NOT NULL,
    pk_script bytea NOT NULL,
    script_class integer NOT NULL,
    addresses character varying(66)[],
    required_sigs integer NOT NULL
);


--
-- Name: tx_outputs_id_seq; Type: SEQUENCE; Schema: btc; Owner: -
--

CREATE SEQUENCE btc.tx_outputs_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tx_outputs_id_seq; Type: SEQUENCE OWNED BY; Schema: btc; Owner: -
--

ALTER SEQUENCE btc.tx_outputs_id_seq OWNED BY btc.tx_outputs.id;


--
-- Name: blocks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.blocks (
    key text NOT NULL,
    data bytea NOT NULL
);


--
-- Name: goose_db_version; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.goose_db_version (
    id integer NOT NULL,
    version_id bigint NOT NULL,
    is_applied boolean NOT NULL,
    tstamp timestamp without time zone DEFAULT now()
);


--
-- Name: goose_db_version_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.goose_db_version_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: goose_db_version_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.goose_db_version_id_seq OWNED BY public.goose_db_version.id;


--
-- Name: nodes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nodes (
    id integer NOT NULL,
    client_name character varying,
    genesis_block character varying(66),
    network_id character varying,
    node_id character varying(128)
);


--
-- Name: TABLE nodes; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.nodes IS '@name NodeInfo';


--
-- Name: COLUMN nodes.node_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nodes.node_id IS '@name ChainNodeID';


--
-- Name: nodes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nodes_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nodes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nodes_id_seq OWNED BY public.nodes.id;


--
-- Name: header_cids id; Type: DEFAULT; Schema: btc; Owner: -
--

ALTER TABLE ONLY btc.header_cids ALTER COLUMN id SET DEFAULT nextval('btc.header_cids_id_seq'::regclass);


--
-- Name: transaction_cids id; Type: DEFAULT; Schema: btc; Owner: -
--

ALTER TABLE ONLY btc.transaction_cids ALTER COLUMN id SET DEFAULT nextval('btc.transaction_cids_id_seq'::regclass);


--
-- Name: tx_inputs id; Type: DEFAULT; Schema: btc; Owner: -
--

ALTER TABLE ONLY btc.tx_inputs ALTER COLUMN id SET DEFAULT nextval('btc.tx_inputs_id_seq'::regclass);


--
-- Name: tx_outputs id; Type: DEFAULT; Schema: btc; Owner: -
--

ALTER TABLE ONLY btc.tx_outputs ALTER COLUMN id SET DEFAULT nextval('btc.tx_outputs_id_seq'::regclass);


--
-- Name: goose_db_version id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.goose_db_version ALTER COLUMN id SET DEFAULT nextval('public.goose_db_version_id_seq'::regclass);


--
-- Name: nodes id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nodes ALTER COLUMN id SET DEFAULT nextval('public.nodes_id_seq'::regclass);


--
-- Name: header_cids header_cids_block_number_block_hash_key; Type: CONSTRAINT; Schema: btc; Owner: -
--

ALTER TABLE ONLY btc.header_cids
    ADD CONSTRAINT header_cids_block_number_block_hash_key UNIQUE (block_number, block_hash);


--
-- Name: header_cids header_cids_pkey; Type: CONSTRAINT; Schema: btc; Owner: -
--

ALTER TABLE ONLY btc.header_cids
    ADD CONSTRAINT header_cids_pkey PRIMARY KEY (id);


--
-- Name: transaction_cids transaction_cids_header_id_tx_hash_key; Type: CONSTRAINT; Schema: btc; Owner: -
--

ALTER TABLE ONLY btc.transaction_cids
    ADD CONSTRAINT transaction_cids_header_id_tx_hash_key UNIQUE (header_id, tx_hash);


--
-- Name: transaction_cids transaction_cids_pkey; Type: CONSTRAINT; Schema: btc; Owner: -
--

ALTER TABLE ONLY btc.transaction_cids
    ADD CONSTRAINT transaction_cids_pkey PRIMARY KEY (id);


--
-- Name: tx_inputs tx_inputs_pkey; Type: CONSTRAINT; Schema: btc; Owner: -
--

ALTER TABLE ONLY btc.tx_inputs
    ADD CONSTRAINT tx_inputs_pkey PRIMARY KEY (id);


--
-- Name: tx_inputs tx_inputs_tx_id_index_key; Type: CONSTRAINT; Schema: btc; Owner: -
--

ALTER TABLE ONLY btc.tx_inputs
    ADD CONSTRAINT tx_inputs_tx_id_index_key UNIQUE (tx_id, index);


--
-- Name: tx_outputs tx_outputs_pkey; Type: CONSTRAINT; Schema: btc; Owner: -
--

ALTER TABLE ONLY btc.tx_outputs
    ADD CONSTRAINT tx_outputs_pkey PRIMARY KEY (id);


--
-- Name: tx_outputs tx_outputs_tx_id_index_key; Type: CONSTRAINT; Schema: btc; Owner: -
--

ALTER TABLE ONLY btc.tx_outputs
    ADD CONSTRAINT tx_outputs_tx_id_index_key UNIQUE (tx_id, index);


--
-- Name: blocks blocks_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blocks
    ADD CONSTRAINT blocks_key_key UNIQUE (key);


--
-- Name: goose_db_version goose_db_version_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.goose_db_version
    ADD CONSTRAINT goose_db_version_pkey PRIMARY KEY (id);


--
-- Name: nodes node_uc; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nodes
    ADD CONSTRAINT node_uc UNIQUE (genesis_block, network_id, node_id);


--
-- Name: nodes nodes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nodes
    ADD CONSTRAINT nodes_pkey PRIMARY KEY (id);


--
-- Name: header_cids header_cids_mh_key_fkey; Type: FK CONSTRAINT; Schema: btc; Owner: -
--

ALTER TABLE ONLY btc.header_cids
    ADD CONSTRAINT header_cids_mh_key_fkey FOREIGN KEY (mh_key) REFERENCES public.blocks(key) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: header_cids header_cids_node_id_fkey; Type: FK CONSTRAINT; Schema: btc; Owner: -
--

ALTER TABLE ONLY btc.header_cids
    ADD CONSTRAINT header_cids_node_id_fkey FOREIGN KEY (node_id) REFERENCES public.nodes(id) ON DELETE CASCADE;


--
-- Name: transaction_cids transaction_cids_header_id_fkey; Type: FK CONSTRAINT; Schema: btc; Owner: -
--

ALTER TABLE ONLY btc.transaction_cids
    ADD CONSTRAINT transaction_cids_header_id_fkey FOREIGN KEY (header_id) REFERENCES btc.header_cids(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: transaction_cids transaction_cids_mh_key_fkey; Type: FK CONSTRAINT; Schema: btc; Owner: -
--

ALTER TABLE ONLY btc.transaction_cids
    ADD CONSTRAINT transaction_cids_mh_key_fkey FOREIGN KEY (mh_key) REFERENCES public.blocks(key) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: tx_inputs tx_inputs_tx_id_fkey; Type: FK CONSTRAINT; Schema: btc; Owner: -
--

ALTER TABLE ONLY btc.tx_inputs
    ADD CONSTRAINT tx_inputs_tx_id_fkey FOREIGN KEY (tx_id) REFERENCES btc.transaction_cids(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: tx_outputs tx_outputs_tx_id_fkey; Type: FK CONSTRAINT; Schema: btc; Owner: -
--

ALTER TABLE ONLY btc.tx_outputs
    ADD CONSTRAINT tx_outputs_tx_id_fkey FOREIGN KEY (tx_id) REFERENCES btc.transaction_cids(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- PostgreSQL database dump complete
--

