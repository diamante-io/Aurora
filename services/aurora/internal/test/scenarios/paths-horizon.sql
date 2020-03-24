--
-- PostgreSQL database dump
--

-- Dumped from database version 9.6.1
-- Dumped by pg_dump version 9.6.1

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

SET search_path = public, pg_catalog;

ALTER TABLE IF EXISTS ONLY public.history_trades DROP CONSTRAINT IF EXISTS history_trades_counter_asset_id_fkey;
ALTER TABLE IF EXISTS ONLY public.history_trades DROP CONSTRAINT IF EXISTS history_trades_counter_account_id_fkey;
ALTER TABLE IF EXISTS ONLY public.history_trades DROP CONSTRAINT IF EXISTS history_trades_base_asset_id_fkey;
ALTER TABLE IF EXISTS ONLY public.history_trades DROP CONSTRAINT IF EXISTS history_trades_base_account_id_fkey;
ALTER TABLE IF EXISTS ONLY public.asset_stats DROP CONSTRAINT IF EXISTS asset_stats_id_fkey;
DROP INDEX IF EXISTS public.trade_effects_by_order_book;
DROP INDEX IF EXISTS public.index_history_transactions_on_id;
DROP INDEX IF EXISTS public.index_history_operations_on_type;
DROP INDEX IF EXISTS public.index_history_operations_on_transaction_id;
DROP INDEX IF EXISTS public.index_history_operations_on_id;
DROP INDEX IF EXISTS public.index_history_ledgers_on_sequence;
DROP INDEX IF EXISTS public.index_history_ledgers_on_previous_ledger_hash;
DROP INDEX IF EXISTS public.index_history_ledgers_on_ledger_hash;
DROP INDEX IF EXISTS public.index_history_ledgers_on_importer_version;
DROP INDEX IF EXISTS public.index_history_ledgers_on_id;
DROP INDEX IF EXISTS public.index_history_ledgers_on_closed_at;
DROP INDEX IF EXISTS public.index_history_effects_on_type;
DROP INDEX IF EXISTS public.index_history_accounts_on_id;
DROP INDEX IF EXISTS public.index_history_accounts_on_address;
DROP INDEX IF EXISTS public.htrd_time_lookup;
DROP INDEX IF EXISTS public.htrd_pid;
DROP INDEX IF EXISTS public.htrd_pair_time_lookup;
DROP INDEX IF EXISTS public.htrd_counter_lookup;
DROP INDEX IF EXISTS public.htrd_by_offer;
DROP INDEX IF EXISTS public.htrd_by_counter_offer;
DROP INDEX IF EXISTS public.htrd_by_counter_account;
DROP INDEX IF EXISTS public.htrd_by_base_offer;
DROP INDEX IF EXISTS public.htrd_by_base_account;
DROP INDEX IF EXISTS public.htp_by_htid;
DROP INDEX IF EXISTS public.hs_transaction_by_id;
DROP INDEX IF EXISTS public.hs_ledger_by_id;
DROP INDEX IF EXISTS public.hop_by_hoid;
DROP INDEX IF EXISTS public.hist_tx_p_id;
DROP INDEX IF EXISTS public.hist_op_p_id;
DROP INDEX IF EXISTS public.hist_e_id;
DROP INDEX IF EXISTS public.hist_e_by_order;
DROP INDEX IF EXISTS public.by_ledger;
DROP INDEX IF EXISTS public.by_hash;
DROP INDEX IF EXISTS public.by_account;
DROP INDEX IF EXISTS public.asset_by_issuer;
DROP INDEX IF EXISTS public.asset_by_code;
ALTER TABLE IF EXISTS ONLY public.history_transaction_participants DROP CONSTRAINT IF EXISTS history_transaction_participants_pkey;
ALTER TABLE IF EXISTS ONLY public.history_operation_participants DROP CONSTRAINT IF EXISTS history_operation_participants_pkey;
ALTER TABLE IF EXISTS ONLY public.history_assets DROP CONSTRAINT IF EXISTS history_assets_pkey;
ALTER TABLE IF EXISTS ONLY public.history_assets DROP CONSTRAINT IF EXISTS history_assets_asset_code_asset_type_asset_issuer_key;
ALTER TABLE IF EXISTS ONLY public.gorp_migrations DROP CONSTRAINT IF EXISTS gorp_migrations_pkey;
ALTER TABLE IF EXISTS ONLY public.asset_stats DROP CONSTRAINT IF EXISTS asset_stats_pkey;
ALTER TABLE IF EXISTS public.history_transaction_participants ALTER COLUMN id DROP DEFAULT;
ALTER TABLE IF EXISTS public.history_operation_participants ALTER COLUMN id DROP DEFAULT;
ALTER TABLE IF EXISTS public.history_assets ALTER COLUMN id DROP DEFAULT;
DROP TABLE IF EXISTS public.history_transactions;
DROP SEQUENCE IF EXISTS public.history_transaction_participants_id_seq;
DROP TABLE IF EXISTS public.history_transaction_participants;
DROP TABLE IF EXISTS public.history_trades;
DROP TABLE IF EXISTS public.history_operations;
DROP SEQUENCE IF EXISTS public.history_operation_participants_id_seq;
DROP TABLE IF EXISTS public.history_operation_participants;
DROP TABLE IF EXISTS public.history_ledgers;
DROP TABLE IF EXISTS public.history_effects;
DROP SEQUENCE IF EXISTS public.history_assets_id_seq;
DROP TABLE IF EXISTS public.history_assets;
DROP TABLE IF EXISTS public.history_accounts;
DROP SEQUENCE IF EXISTS public.history_accounts_id_seq;
DROP TABLE IF EXISTS public.gorp_migrations;
DROP TABLE IF EXISTS public.asset_stats;
DROP AGGREGATE IF EXISTS public.min_price(numeric[]);
DROP AGGREGATE IF EXISTS public.max_price(numeric[]);
DROP AGGREGATE IF EXISTS public.last(anyelement);
DROP AGGREGATE IF EXISTS public.first(anyelement);
DROP FUNCTION IF EXISTS public.min_price_agg(numeric[], numeric[]);
DROP FUNCTION IF EXISTS public.max_price_agg(numeric[], numeric[]);
DROP FUNCTION IF EXISTS public.last_agg(anyelement, anyelement);
DROP FUNCTION IF EXISTS public.first_agg(anyelement, anyelement);
DROP EXTENSION IF EXISTS plpgsql;
DROP SCHEMA IF EXISTS public;
--
-- Name: public; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA public;


--
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA public IS 'standard public schema';


--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET search_path = public, pg_catalog;

--
-- Name: first_agg(anyelement, anyelement); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION first_agg(anyelement, anyelement) RETURNS anyelement
    LANGUAGE sql IMMUTABLE STRICT
    AS $_$ SELECT $1 $_$;


--
-- Name: last_agg(anyelement, anyelement); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION last_agg(anyelement, anyelement) RETURNS anyelement
    LANGUAGE sql IMMUTABLE STRICT
    AS $_$ SELECT $2 $_$;


--
-- Name: max_price_agg(numeric[], numeric[]); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION max_price_agg(numeric[], numeric[]) RETURNS numeric[]
    LANGUAGE sql IMMUTABLE STRICT
    AS $_$ SELECT (
  CASE WHEN $1[1]/$1[2]>$2[1]/$2[2] THEN $1 ELSE $2 END) $_$;


--
-- Name: min_price_agg(numeric[], numeric[]); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION min_price_agg(numeric[], numeric[]) RETURNS numeric[]
    LANGUAGE sql IMMUTABLE STRICT
    AS $_$ SELECT (
  CASE WHEN $1[1]/$1[2]<$2[1]/$2[2] THEN $1 ELSE $2 END) $_$;


--
-- Name: first(anyelement); Type: AGGREGATE; Schema: public; Owner: -
--

CREATE AGGREGATE first(anyelement) (
    SFUNC = first_agg,
    STYPE = anyelement
);


--
-- Name: last(anyelement); Type: AGGREGATE; Schema: public; Owner: -
--

CREATE AGGREGATE last(anyelement) (
    SFUNC = last_agg,
    STYPE = anyelement
);


--
-- Name: max_price(numeric[]); Type: AGGREGATE; Schema: public; Owner: -
--

CREATE AGGREGATE max_price(numeric[]) (
    SFUNC = max_price_agg,
    STYPE = numeric[]
);


--
-- Name: min_price(numeric[]); Type: AGGREGATE; Schema: public; Owner: -
--

CREATE AGGREGATE min_price(numeric[]) (
    SFUNC = min_price_agg,
    STYPE = numeric[]
);


SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: asset_stats; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE asset_stats (
    id bigint NOT NULL,
    amount character varying NOT NULL,
    num_accounts integer NOT NULL,
    flags smallint NOT NULL,
    toml character varying(255) NOT NULL
);


--
-- Name: gorp_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE gorp_migrations (
    id text NOT NULL,
    applied_at timestamp with time zone
);


--
-- Name: history_accounts_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE history_accounts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: history_accounts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE history_accounts (
    id bigint DEFAULT nextval('history_accounts_id_seq'::regclass) NOT NULL,
    address character varying(64)
);


--
-- Name: history_assets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE history_assets (
    id integer NOT NULL,
    asset_type character varying(64) NOT NULL,
    asset_code character varying(12) NOT NULL,
    asset_issuer character varying(56) NOT NULL
);


--
-- Name: history_assets_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE history_assets_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: history_assets_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE history_assets_id_seq OWNED BY history_assets.id;


--
-- Name: history_effects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE history_effects (
    history_account_id bigint NOT NULL,
    history_operation_id bigint NOT NULL,
    "order" integer NOT NULL,
    type integer NOT NULL,
    details jsonb
);


--
-- Name: history_ledgers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE history_ledgers (
    sequence integer NOT NULL,
    ledger_hash character varying(64) NOT NULL,
    previous_ledger_hash character varying(64),
    transaction_count integer DEFAULT 0 NOT NULL,
    operation_count integer DEFAULT 0 NOT NULL,
    closed_at timestamp without time zone NOT NULL,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    id bigint,
    importer_version integer DEFAULT 1 NOT NULL,
    total_coins bigint NOT NULL,
    fee_pool bigint NOT NULL,
    base_fee integer NOT NULL,
    base_reserve integer NOT NULL,
    max_tx_set_size integer NOT NULL,
    protocol_version integer DEFAULT 0 NOT NULL,
    ledger_header text,
    successful_transaction_count integer,
    failed_transaction_count integer
);


--
-- Name: history_operation_participants; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE history_operation_participants (
    id integer NOT NULL,
    history_operation_id bigint NOT NULL,
    history_account_id bigint NOT NULL
);


--
-- Name: history_operation_participants_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE history_operation_participants_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: history_operation_participants_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE history_operation_participants_id_seq OWNED BY history_operation_participants.id;


--
-- Name: history_operations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE history_operations (
    id bigint NOT NULL,
    transaction_id bigint NOT NULL,
    application_order integer NOT NULL,
    type integer NOT NULL,
    details jsonb,
    source_account character varying(64) DEFAULT ''::character varying NOT NULL
);


--
-- Name: history_trades; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE history_trades (
    history_operation_id bigint NOT NULL,
    "order" integer NOT NULL,
    ledger_closed_at timestamp without time zone NOT NULL,
    offer_id bigint NOT NULL,
    base_account_id bigint NOT NULL,
    base_asset_id bigint NOT NULL,
    base_amount bigint NOT NULL,
    counter_account_id bigint NOT NULL,
    counter_asset_id bigint NOT NULL,
    counter_amount bigint NOT NULL,
    base_is_seller boolean,
    price_n bigint,
    price_d bigint,
    base_offer_id bigint,
    counter_offer_id bigint,
    CONSTRAINT history_trades_base_amount_check CHECK ((base_amount > 0)),
    CONSTRAINT history_trades_check CHECK ((base_asset_id < counter_asset_id)),
    CONSTRAINT history_trades_counter_amount_check CHECK ((counter_amount > 0))
);


--
-- Name: history_transaction_participants; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE history_transaction_participants (
    id integer NOT NULL,
    history_transaction_id bigint NOT NULL,
    history_account_id bigint NOT NULL
);


--
-- Name: history_transaction_participants_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE history_transaction_participants_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: history_transaction_participants_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE history_transaction_participants_id_seq OWNED BY history_transaction_participants.id;


--
-- Name: history_transactions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE history_transactions (
    transaction_hash character varying(64) NOT NULL,
    ledger_sequence integer NOT NULL,
    application_order integer NOT NULL,
    account character varying(64) NOT NULL,
    account_sequence bigint NOT NULL,
    max_fee integer NOT NULL,
    operation_count integer NOT NULL,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    id bigint,
    tx_envelope text NOT NULL,
    tx_result text NOT NULL,
    tx_meta text NOT NULL,
    tx_fee_meta text NOT NULL,
    signatures character varying(96)[] DEFAULT '{}'::character varying[] NOT NULL,
    memo_type character varying DEFAULT 'none'::character varying NOT NULL,
    memo character varying,
    time_bounds int8range,
    successful boolean,
    fee_charged integer
);


--
-- Name: history_assets id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY history_assets ALTER COLUMN id SET DEFAULT nextval('history_assets_id_seq'::regclass);


--
-- Name: history_operation_participants id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY history_operation_participants ALTER COLUMN id SET DEFAULT nextval('history_operation_participants_id_seq'::regclass);


--
-- Name: history_transaction_participants id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY history_transaction_participants ALTER COLUMN id SET DEFAULT nextval('history_transaction_participants_id_seq'::regclass);


--
-- Data for Name: asset_stats; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO asset_stats VALUES (1, '50000000000', 1, 0, '');
INSERT INTO asset_stats VALUES (2, '50000000000', 1, 0, '');
INSERT INTO asset_stats VALUES (3, '50000000000', 2, 0, '');
INSERT INTO asset_stats VALUES (4, '50000000000', 1, 0, '');
INSERT INTO asset_stats VALUES (5, '50000000000', 2, 0, '');
INSERT INTO asset_stats VALUES (6, '50000000000', 1, 0, '');
INSERT INTO asset_stats VALUES (7, '50000000000', 1, 0, '');
INSERT INTO asset_stats VALUES (8, '50000000000', 1, 0, '');
INSERT INTO asset_stats VALUES (9, '50000000000', 1, 0, '');
INSERT INTO asset_stats VALUES (10, '50000000000', 2, 0, '');
INSERT INTO asset_stats VALUES (11, '50000000000', 2, 0, '');


--
-- Data for Name: gorp_migrations; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO gorp_migrations VALUES ('1_initial_schema.sql', '2019-06-03 18:28:47.032496+02');
INSERT INTO gorp_migrations VALUES ('2_index_participants_by_toid.sql', '2019-06-03 18:28:47.039657+02');
INSERT INTO gorp_migrations VALUES ('3_use_sequence_in_history_accounts.sql', '2019-06-03 18:28:47.044048+02');
INSERT INTO gorp_migrations VALUES ('4_add_protocol_version.sql', '2019-06-03 18:28:47.054532+02');
INSERT INTO gorp_migrations VALUES ('5_create_trades_table.sql', '2019-06-03 18:28:47.063028+02');
INSERT INTO gorp_migrations VALUES ('6_create_assets_table.sql', '2019-06-03 18:28:47.068415+02');
INSERT INTO gorp_migrations VALUES ('7_modify_trades_table.sql', '2019-06-03 18:28:47.081625+02');
INSERT INTO gorp_migrations VALUES ('8_create_asset_stats_table.sql', '2019-06-03 18:28:47.087463+02');
INSERT INTO gorp_migrations VALUES ('8_add_aggregators.sql', '2019-06-03 18:28:47.090109+02');
INSERT INTO gorp_migrations VALUES ('9_add_header_xdr.sql', '2019-06-03 18:28:47.092718+02');
INSERT INTO gorp_migrations VALUES ('10_add_trades_price.sql', '2019-06-03 18:28:47.095973+02');
INSERT INTO gorp_migrations VALUES ('11_add_trades_account_index.sql', '2019-06-03 18:28:47.099698+02');
INSERT INTO gorp_migrations VALUES ('12_asset_stats_amount_string.sql', '2019-06-03 18:28:47.107549+02');
INSERT INTO gorp_migrations VALUES ('13_trade_offer_ids.sql', '2019-06-03 18:28:47.112768+02');
INSERT INTO gorp_migrations VALUES ('14_fix_asset_toml_field.sql', '2019-06-03 18:28:47.115116+02');
INSERT INTO gorp_migrations VALUES ('15_ledger_failed_txs.sql', '2019-06-03 18:28:47.116796+02');
INSERT INTO gorp_migrations VALUES ('16_ingest_failed_transactions.sql', '2019-06-03 18:28:47.117989+02');
INSERT INTO gorp_migrations VALUES ('17_transaction_fee_paid.sql', '2019-06-03 18:28:47.120034+02');


--
-- Data for Name: history_accounts; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO history_accounts VALUES (1, 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN');
INSERT INTO history_accounts VALUES (2, 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL');
INSERT INTO history_accounts VALUES (3, 'GARSFJNXJIHO6ULUBK3DBYKVSIZE7SC72S5DYBCHU7DKL22UXKVD7MXP');
INSERT INTO history_accounts VALUES (4, 'GAEDTJ4PPEFVW5XV2S7LUXBEHNQMX5Q2GM562RJGOQG7GVCE5H3HIB4V');
INSERT INTO history_accounts VALUES (5, 'GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H');


--
-- Name: history_accounts_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('history_accounts_id_seq', 5, true);


--
-- Data for Name: history_assets; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO history_assets VALUES (1, 'credit_alphanum4', '33', 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN');
INSERT INTO history_assets VALUES (2, 'credit_alphanum4', '32', 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN');
INSERT INTO history_assets VALUES (3, 'credit_alphanum4', 'AAA', 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN');
INSERT INTO history_assets VALUES (4, 'credit_alphanum4', '31', 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN');
INSERT INTO history_assets VALUES (5, 'credit_alphanum4', 'EUR', 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN');
INSERT INTO history_assets VALUES (6, 'credit_alphanum4', '22', 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN');
INSERT INTO history_assets VALUES (7, 'credit_alphanum4', 'BBB', 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN');
INSERT INTO history_assets VALUES (8, 'credit_alphanum4', '1', 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN');
INSERT INTO history_assets VALUES (9, 'credit_alphanum4', '21', 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN');
INSERT INTO history_assets VALUES (10, 'credit_alphanum4', 'USD', 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN');
INSERT INTO history_assets VALUES (11, 'credit_alphanum4', 'CCC', 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN');


--
-- Name: history_assets_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('history_assets_id_seq', 11, true);


--
-- Data for Name: history_effects; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO history_effects VALUES (3, 17179873281, 1, 2, '{"amount": "5000.0000000", "asset_code": "USD", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (1, 17179873281, 2, 3, '{"amount": "5000.0000000", "asset_code": "USD", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (2, 17179877377, 1, 2, '{"amount": "5000.0000000", "asset_code": "EUR", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (1, 17179877377, 2, 3, '{"amount": "5000.0000000", "asset_code": "EUR", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (2, 17179881473, 1, 2, '{"amount": "5000.0000000", "asset_code": "1", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (1, 17179881473, 2, 3, '{"amount": "5000.0000000", "asset_code": "1", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (2, 17179885569, 1, 2, '{"amount": "5000.0000000", "asset_code": "21", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (1, 17179885569, 2, 3, '{"amount": "5000.0000000", "asset_code": "21", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (2, 17179889665, 1, 2, '{"amount": "5000.0000000", "asset_code": "22", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (1, 17179889665, 2, 3, '{"amount": "5000.0000000", "asset_code": "22", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (2, 17179893761, 1, 2, '{"amount": "5000.0000000", "asset_code": "31", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (1, 17179893761, 2, 3, '{"amount": "5000.0000000", "asset_code": "31", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (2, 17179897857, 1, 2, '{"amount": "5000.0000000", "asset_code": "32", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (1, 17179897857, 2, 3, '{"amount": "5000.0000000", "asset_code": "32", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (2, 17179901953, 1, 2, '{"amount": "5000.0000000", "asset_code": "33", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (1, 17179901953, 2, 3, '{"amount": "5000.0000000", "asset_code": "33", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (2, 17179906049, 1, 2, '{"amount": "5000.0000000", "asset_code": "AAA", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (1, 17179906049, 2, 3, '{"amount": "5000.0000000", "asset_code": "AAA", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (2, 17179910145, 1, 2, '{"amount": "5000.0000000", "asset_code": "BBB", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (1, 17179910145, 2, 3, '{"amount": "5000.0000000", "asset_code": "BBB", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (2, 17179914241, 1, 2, '{"amount": "5000.0000000", "asset_code": "CCC", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (1, 17179914241, 2, 3, '{"amount": "5000.0000000", "asset_code": "CCC", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (2, 12884905985, 1, 20, '{"limit": "922337203685.4775807", "asset_code": "USD", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (4, 12884910081, 1, 20, '{"limit": "922337203685.4775807", "asset_code": "EUR", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (3, 12884914177, 1, 20, '{"limit": "922337203685.4775807", "asset_code": "USD", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (4, 12884918273, 1, 20, '{"limit": "922337203685.4775807", "asset_code": "CCC", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (3, 12884922369, 1, 20, '{"limit": "922337203685.4775807", "asset_code": "AAA", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (2, 12884926465, 1, 20, '{"limit": "922337203685.4775807", "asset_code": "EUR", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (2, 12884930561, 1, 20, '{"limit": "922337203685.4775807", "asset_code": "1", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (2, 12884934657, 1, 20, '{"limit": "922337203685.4775807", "asset_code": "21", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (2, 12884938753, 1, 20, '{"limit": "922337203685.4775807", "asset_code": "22", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (2, 12884942849, 1, 20, '{"limit": "922337203685.4775807", "asset_code": "31", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (2, 12884946945, 1, 20, '{"limit": "922337203685.4775807", "asset_code": "32", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (2, 12884951041, 1, 20, '{"limit": "922337203685.4775807", "asset_code": "33", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (2, 12884955137, 1, 20, '{"limit": "922337203685.4775807", "asset_code": "AAA", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (2, 12884959233, 1, 20, '{"limit": "922337203685.4775807", "asset_code": "BBB", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (2, 12884963329, 1, 20, '{"limit": "922337203685.4775807", "asset_code": "CCC", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (1, 8589938689, 1, 0, '{"starting_balance": "10000.0000000"}');
INSERT INTO history_effects VALUES (5, 8589938689, 2, 3, '{"amount": "10000.0000000", "asset_type": "native"}');
INSERT INTO history_effects VALUES (1, 8589938689, 3, 10, '{"weight": 1, "public_key": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}');
INSERT INTO history_effects VALUES (3, 8589942785, 1, 0, '{"starting_balance": "1000.0000000"}');
INSERT INTO history_effects VALUES (5, 8589942785, 2, 3, '{"amount": "1000.0000000", "asset_type": "native"}');
INSERT INTO history_effects VALUES (3, 8589942785, 3, 10, '{"weight": 1, "public_key": "GARSFJNXJIHO6ULUBK3DBYKVSIZE7SC72S5DYBCHU7DKL22UXKVD7MXP"}');
INSERT INTO history_effects VALUES (4, 8589946881, 1, 0, '{"starting_balance": "1000.0000000"}');
INSERT INTO history_effects VALUES (5, 8589946881, 2, 3, '{"amount": "1000.0000000", "asset_type": "native"}');
INSERT INTO history_effects VALUES (4, 8589946881, 3, 10, '{"weight": 1, "public_key": "GAEDTJ4PPEFVW5XV2S7LUXBEHNQMX5Q2GM562RJGOQG7GVCE5H3HIB4V"}');
INSERT INTO history_effects VALUES (2, 8589950977, 1, 0, '{"starting_balance": "1000.0000000"}');
INSERT INTO history_effects VALUES (5, 8589950977, 2, 3, '{"amount": "1000.0000000", "asset_type": "native"}');
INSERT INTO history_effects VALUES (2, 8589950977, 3, 10, '{"weight": 1, "public_key": "GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL"}');


--
-- Data for Name: history_ledgers; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO history_ledgers VALUES (5, 'bd1228398c2ae39a6f483ff0ed100732432de2aafb4690498cbb2cc095828af5', '1682cf1bdff3a29952f78d3cebfb29025c4af85516e0cf196f0eeec92faf0356', 15, 15, '2019-06-03 16:38:11', '2019-06-03 16:38:11.911058', '2019-06-03 16:38:11.911059', 21474836480, 16, 1000000000000000000, 4500, 100, 100000000, 1000000, 11, 'AAAACxaCzxvf86KZUveNPOv7KQJcSvhVFuDPGW8O7skvrwNWENsioz2ltnCClQFRkKqzIt3YRaLKoCjfJHkamgnGALAAAAAAXPVM8wAAAAAAAAAAmPb/IdFVU6DmHBuxXut8PCRUBW9WgS/0Xucy3DNxDm5RYvF/Q6bRIZ2X43sfa0yFBI8ELiU8upRod5EQdKXzNAAAAAUN4Lazp2QAAAAAAAAAABGUAAAAAAAAAAAAAAAPAAAAZAX14QAAD0JAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA', 15, 0);
INSERT INTO history_ledgers VALUES (4, '1682cf1bdff3a29952f78d3cebfb29025c4af85516e0cf196f0eeec92faf0356', '55aaea38f94bd5daf575810390ba2399588532ca118d7bbe219e71dbc909f5db', 11, 11, '2019-06-03 16:38:10', '2019-06-03 16:38:11.945185', '2019-06-03 16:38:11.945185', 17179869184, 16, 1000000000000000000, 3000, 100, 100000000, 1000000, 11, 'AAAAC1Wq6jj5S9Xa9XWBA5C6I5lYhTLKEY17viGecdvJCfXbxB0mxnuHmjyUQaVikKU+PcS/+ikSiJZOux3hV+4VltMAAAAAXPVM8gAAAAAAAAAA4+bAkQbZh/sznT0pfrtwonK6PyJOPO/u91w/KMUkHHW6/cb1OkBEaIr5EukZYqvq3UIW7uGUWzIq2cz6//7MxwAAAAQN4Lazp2QAAAAAAAAAAAu4AAAAAAAAAAAAAAAAAAAAZAX14QAAD0JAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA', 11, 0);
INSERT INTO history_ledgers VALUES (3, '55aaea38f94bd5daf575810390ba2399588532ca118d7bbe219e71dbc909f5db', 'e3b5c5d2acfa8689353c520aa5aa6c8736080303224a7221301463798a3687d3', 15, 15, '2019-06-03 16:38:09', '2019-06-03 16:38:11.969083', '2019-06-03 16:38:11.969083', 12884901888, 16, 1000000000000000000, 1900, 100, 100000000, 1000000, 11, 'AAAAC+O1xdKs+oaJNTxSCqWqbIc2CAMDIkpyITAUY3mKNofTUK0HBm1SUYgq4BLE+92XcAwxFWREVEAEQH2uK+jheJgAAAAAXPVM8QAAAAAAAAAA9N5SNo8Ql8/Li/reP2oZ5eQO7+sGLMSpq7GdBSxy5IKQlguBbfcOLEvp5q3WpaqaApYx+UTuY3jPjwCSl+L3HwAAAAMN4Lazp2QAAAAAAAAAAAdsAAAAAAAAAAAAAAAAAAAAZAX14QAAD0JAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA', 15, 0);
INSERT INTO history_ledgers VALUES (2, 'e3b5c5d2acfa8689353c520aa5aa6c8736080303224a7221301463798a3687d3', '63d98f536ee68d1b27b5b89f23af5311b7569a24faf1403ad0b52b633b07be99', 4, 4, '2019-06-03 16:38:08', '2019-06-03 16:38:11.987022', '2019-06-03 16:38:11.987023', 8589934592, 16, 1000000000000000000, 400, 100, 100000000, 1000000, 11, 'AAAAC2PZj1Nu5o0bJ7W4nyOvUxG3Vpok+vFAOtC1K2M7B76ZEgIGEiV+mcm4b2nIL9vjp47MoLQRzOY75wUGqnsob9EAAAAAXPVM8AAAAAIAAAAIAAAAAQAAAAsAAAAIAAAAAwAPQkAAAAAAyInJ6Wk07o04mgDy1803oo/6UGNrb+83mDh0MPMoMykAE9qXjtFtt9xer/uYvhTO2R27voGxoFlSrOs9S+WHJQAAAAIN4Lazp2QAAAAAAAAAAAGQAAAAAAAAAAAAAAAAAAAAZAX14QAAD0JAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA', 4, 0);
INSERT INTO history_ledgers VALUES (1, '63d98f536ee68d1b27b5b89f23af5311b7569a24faf1403ad0b52b633b07be99', NULL, 0, 0, '1970-01-01 00:00:00', '2019-06-03 16:38:12.002721', '2019-06-03 16:38:12.002721', 4294967296, 16, 1000000000000000000, 0, 100, 100000000, 100, 0, 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABXKi4y/ySKB7DnD9H20xjB+s0gtswIwz1XdSWYaBJaFgAAAAEN4Lazp2QAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAZAX14QAAAABkAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA', 0, 0);


--
-- Data for Name: history_operation_participants; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO history_operation_participants VALUES (1, 21474840577, 1);
INSERT INTO history_operation_participants VALUES (2, 21474844673, 2);
INSERT INTO history_operation_participants VALUES (3, 21474848769, 2);
INSERT INTO history_operation_participants VALUES (4, 21474852865, 1);
INSERT INTO history_operation_participants VALUES (5, 21474856961, 2);
INSERT INTO history_operation_participants VALUES (6, 21474861057, 1);
INSERT INTO history_operation_participants VALUES (7, 21474865153, 2);
INSERT INTO history_operation_participants VALUES (8, 21474869249, 2);
INSERT INTO history_operation_participants VALUES (9, 21474873345, 2);
INSERT INTO history_operation_participants VALUES (10, 21474877441, 2);
INSERT INTO history_operation_participants VALUES (11, 21474881537, 2);
INSERT INTO history_operation_participants VALUES (12, 21474885633, 2);
INSERT INTO history_operation_participants VALUES (13, 21474889729, 2);
INSERT INTO history_operation_participants VALUES (14, 21474893825, 2);
INSERT INTO history_operation_participants VALUES (15, 21474897921, 2);
INSERT INTO history_operation_participants VALUES (16, 17179873281, 1);
INSERT INTO history_operation_participants VALUES (17, 17179873281, 3);
INSERT INTO history_operation_participants VALUES (18, 17179877377, 2);
INSERT INTO history_operation_participants VALUES (19, 17179877377, 1);
INSERT INTO history_operation_participants VALUES (20, 17179881473, 1);
INSERT INTO history_operation_participants VALUES (21, 17179881473, 2);
INSERT INTO history_operation_participants VALUES (22, 17179885569, 1);
INSERT INTO history_operation_participants VALUES (23, 17179885569, 2);
INSERT INTO history_operation_participants VALUES (24, 17179889665, 1);
INSERT INTO history_operation_participants VALUES (25, 17179889665, 2);
INSERT INTO history_operation_participants VALUES (26, 17179893761, 1);
INSERT INTO history_operation_participants VALUES (27, 17179893761, 2);
INSERT INTO history_operation_participants VALUES (28, 17179897857, 2);
INSERT INTO history_operation_participants VALUES (29, 17179897857, 1);
INSERT INTO history_operation_participants VALUES (30, 17179901953, 1);
INSERT INTO history_operation_participants VALUES (31, 17179901953, 2);
INSERT INTO history_operation_participants VALUES (32, 17179906049, 1);
INSERT INTO history_operation_participants VALUES (33, 17179906049, 2);
INSERT INTO history_operation_participants VALUES (34, 17179910145, 1);
INSERT INTO history_operation_participants VALUES (35, 17179910145, 2);
INSERT INTO history_operation_participants VALUES (36, 17179914241, 1);
INSERT INTO history_operation_participants VALUES (37, 17179914241, 2);
INSERT INTO history_operation_participants VALUES (38, 12884905985, 2);
INSERT INTO history_operation_participants VALUES (39, 12884910081, 4);
INSERT INTO history_operation_participants VALUES (40, 12884914177, 3);
INSERT INTO history_operation_participants VALUES (41, 12884918273, 4);
INSERT INTO history_operation_participants VALUES (42, 12884922369, 3);
INSERT INTO history_operation_participants VALUES (43, 12884926465, 2);
INSERT INTO history_operation_participants VALUES (44, 12884930561, 2);
INSERT INTO history_operation_participants VALUES (45, 12884934657, 2);
INSERT INTO history_operation_participants VALUES (46, 12884938753, 2);
INSERT INTO history_operation_participants VALUES (47, 12884942849, 2);
INSERT INTO history_operation_participants VALUES (48, 12884946945, 2);
INSERT INTO history_operation_participants VALUES (49, 12884951041, 2);
INSERT INTO history_operation_participants VALUES (50, 12884955137, 2);
INSERT INTO history_operation_participants VALUES (51, 12884959233, 2);
INSERT INTO history_operation_participants VALUES (52, 12884963329, 2);
INSERT INTO history_operation_participants VALUES (53, 8589938689, 5);
INSERT INTO history_operation_participants VALUES (54, 8589938689, 1);
INSERT INTO history_operation_participants VALUES (55, 8589942785, 5);
INSERT INTO history_operation_participants VALUES (56, 8589942785, 3);
INSERT INTO history_operation_participants VALUES (57, 8589946881, 5);
INSERT INTO history_operation_participants VALUES (58, 8589946881, 4);
INSERT INTO history_operation_participants VALUES (59, 8589950977, 5);
INSERT INTO history_operation_participants VALUES (60, 8589950977, 2);


--
-- Name: history_operation_participants_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('history_operation_participants_id_seq', 60, true);


--
-- Data for Name: history_operations; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO history_operations VALUES (21474840577, 21474840576, 1, 3, '{"price": "1.0000000", "amount": "10.0000000", "price_r": {"d": 1, "n": 1}, "offer_id": 0, "buying_asset_code": "USD", "buying_asset_type": "credit_alphanum4", "selling_asset_code": "EUR", "selling_asset_type": "credit_alphanum4", "buying_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "selling_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN');
INSERT INTO history_operations VALUES (21474844673, 21474844672, 1, 3, '{"price": "0.5000000", "amount": "10.0000000", "price_r": {"d": 2, "n": 1}, "offer_id": 0, "buying_asset_code": "USD", "buying_asset_type": "credit_alphanum4", "selling_asset_code": "EUR", "selling_asset_type": "credit_alphanum4", "buying_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "selling_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL');
INSERT INTO history_operations VALUES (21474848769, 21474848768, 1, 3, '{"price": "1.0000000", "amount": "20.0000000", "price_r": {"d": 1, "n": 1}, "offer_id": 0, "buying_asset_code": "USD", "buying_asset_type": "credit_alphanum4", "selling_asset_code": "1", "selling_asset_type": "credit_alphanum4", "buying_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "selling_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL');
INSERT INTO history_operations VALUES (21474852865, 21474852864, 1, 3, '{"price": "0.5000000", "amount": "10.0000000", "price_r": {"d": 2, "n": 1}, "offer_id": 0, "buying_asset_code": "USD", "buying_asset_type": "credit_alphanum4", "selling_asset_code": "EUR", "selling_asset_type": "credit_alphanum4", "buying_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "selling_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN');
INSERT INTO history_operations VALUES (21474856961, 21474856960, 1, 3, '{"price": "1.0000000", "amount": "20.0000000", "price_r": {"d": 1, "n": 1}, "offer_id": 0, "buying_asset_code": "1", "buying_asset_type": "credit_alphanum4", "selling_asset_code": "EUR", "selling_asset_type": "credit_alphanum4", "buying_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "selling_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL');
INSERT INTO history_operations VALUES (21474861057, 21474861056, 1, 3, '{"price": "0.1000000", "amount": "1000.0000000", "price_r": {"d": 10, "n": 1}, "offer_id": 0, "buying_asset_code": "USD", "buying_asset_type": "credit_alphanum4", "selling_asset_type": "native", "buying_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN');
INSERT INTO history_operations VALUES (21474865153, 21474865152, 1, 3, '{"price": "1.0000000", "amount": "30.0000000", "price_r": {"d": 1, "n": 1}, "offer_id": 0, "buying_asset_code": "USD", "buying_asset_type": "credit_alphanum4", "selling_asset_code": "21", "selling_asset_type": "credit_alphanum4", "buying_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "selling_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL');
INSERT INTO history_operations VALUES (21474869249, 21474869248, 1, 3, '{"price": "1.0000000", "amount": "30.0000000", "price_r": {"d": 1, "n": 1}, "offer_id": 0, "buying_asset_code": "21", "buying_asset_type": "credit_alphanum4", "selling_asset_code": "22", "selling_asset_type": "credit_alphanum4", "buying_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "selling_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL');
INSERT INTO history_operations VALUES (21474873345, 21474873344, 1, 3, '{"price": "1.0000000", "amount": "30.0000000", "price_r": {"d": 1, "n": 1}, "offer_id": 0, "buying_asset_code": "22", "buying_asset_type": "credit_alphanum4", "selling_asset_code": "EUR", "selling_asset_type": "credit_alphanum4", "buying_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "selling_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL');
INSERT INTO history_operations VALUES (21474877441, 21474877440, 1, 3, '{"price": "2.0000000", "amount": "40.0000000", "price_r": {"d": 1, "n": 2}, "offer_id": 0, "buying_asset_code": "USD", "buying_asset_type": "credit_alphanum4", "selling_asset_code": "31", "selling_asset_type": "credit_alphanum4", "buying_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "selling_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL');
INSERT INTO history_operations VALUES (21474881537, 21474881536, 1, 3, '{"price": "2.0000000", "amount": "40.0000000", "price_r": {"d": 1, "n": 2}, "offer_id": 0, "buying_asset_code": "31", "buying_asset_type": "credit_alphanum4", "selling_asset_code": "32", "selling_asset_type": "credit_alphanum4", "buying_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "selling_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL');
INSERT INTO history_operations VALUES (21474885633, 21474885632, 1, 3, '{"price": "2.0000000", "amount": "40.0000000", "price_r": {"d": 1, "n": 2}, "offer_id": 0, "buying_asset_code": "32", "buying_asset_type": "credit_alphanum4", "selling_asset_code": "33", "selling_asset_type": "credit_alphanum4", "buying_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "selling_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL');
INSERT INTO history_operations VALUES (21474889729, 21474889728, 1, 3, '{"price": "2.0000000", "amount": "40.0000000", "price_r": {"d": 1, "n": 2}, "offer_id": 0, "buying_asset_code": "33", "buying_asset_type": "credit_alphanum4", "selling_asset_code": "EUR", "selling_asset_type": "credit_alphanum4", "buying_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "selling_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL');
INSERT INTO history_operations VALUES (21474893825, 21474893824, 1, 3, '{"price": "11.0000000", "amount": "1.0000000", "price_r": {"d": 1, "n": 11}, "offer_id": 0, "buying_asset_code": "AAA", "buying_asset_type": "credit_alphanum4", "selling_asset_code": "BBB", "selling_asset_type": "credit_alphanum4", "buying_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "selling_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL');
INSERT INTO history_operations VALUES (21474897921, 21474897920, 1, 3, '{"price": "0.1000000", "amount": "10.0000000", "price_r": {"d": 10, "n": 1}, "offer_id": 0, "buying_asset_code": "BBB", "buying_asset_type": "credit_alphanum4", "selling_asset_code": "CCC", "selling_asset_type": "credit_alphanum4", "buying_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "selling_asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL');
INSERT INTO history_operations VALUES (17179873281, 17179873280, 1, 1, '{"to": "GARSFJNXJIHO6ULUBK3DBYKVSIZE7SC72S5DYBCHU7DKL22UXKVD7MXP", "from": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "amount": "5000.0000000", "asset_code": "USD", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN');
INSERT INTO history_operations VALUES (17179877377, 17179877376, 1, 1, '{"to": "GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL", "from": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "amount": "5000.0000000", "asset_code": "EUR", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN');
INSERT INTO history_operations VALUES (17179881473, 17179881472, 1, 1, '{"to": "GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL", "from": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "amount": "5000.0000000", "asset_code": "1", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN');
INSERT INTO history_operations VALUES (17179885569, 17179885568, 1, 1, '{"to": "GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL", "from": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "amount": "5000.0000000", "asset_code": "21", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN');
INSERT INTO history_operations VALUES (17179889665, 17179889664, 1, 1, '{"to": "GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL", "from": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "amount": "5000.0000000", "asset_code": "22", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN');
INSERT INTO history_operations VALUES (17179893761, 17179893760, 1, 1, '{"to": "GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL", "from": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "amount": "5000.0000000", "asset_code": "31", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN');
INSERT INTO history_operations VALUES (17179897857, 17179897856, 1, 1, '{"to": "GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL", "from": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "amount": "5000.0000000", "asset_code": "32", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN');
INSERT INTO history_operations VALUES (17179901953, 17179901952, 1, 1, '{"to": "GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL", "from": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "amount": "5000.0000000", "asset_code": "33", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN');
INSERT INTO history_operations VALUES (17179906049, 17179906048, 1, 1, '{"to": "GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL", "from": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "amount": "5000.0000000", "asset_code": "AAA", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN');
INSERT INTO history_operations VALUES (17179910145, 17179910144, 1, 1, '{"to": "GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL", "from": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "amount": "5000.0000000", "asset_code": "BBB", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN');
INSERT INTO history_operations VALUES (17179914241, 17179914240, 1, 1, '{"to": "GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL", "from": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "amount": "5000.0000000", "asset_code": "CCC", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN');
INSERT INTO history_operations VALUES (12884905985, 12884905984, 1, 6, '{"limit": "922337203685.4775807", "trustee": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "trustor": "GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL", "asset_code": "USD", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL');
INSERT INTO history_operations VALUES (12884910081, 12884910080, 1, 6, '{"limit": "922337203685.4775807", "trustee": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "trustor": "GAEDTJ4PPEFVW5XV2S7LUXBEHNQMX5Q2GM562RJGOQG7GVCE5H3HIB4V", "asset_code": "EUR", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GAEDTJ4PPEFVW5XV2S7LUXBEHNQMX5Q2GM562RJGOQG7GVCE5H3HIB4V');
INSERT INTO history_operations VALUES (12884914177, 12884914176, 1, 6, '{"limit": "922337203685.4775807", "trustee": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "trustor": "GARSFJNXJIHO6ULUBK3DBYKVSIZE7SC72S5DYBCHU7DKL22UXKVD7MXP", "asset_code": "USD", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GARSFJNXJIHO6ULUBK3DBYKVSIZE7SC72S5DYBCHU7DKL22UXKVD7MXP');
INSERT INTO history_operations VALUES (12884918273, 12884918272, 1, 6, '{"limit": "922337203685.4775807", "trustee": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "trustor": "GAEDTJ4PPEFVW5XV2S7LUXBEHNQMX5Q2GM562RJGOQG7GVCE5H3HIB4V", "asset_code": "CCC", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GAEDTJ4PPEFVW5XV2S7LUXBEHNQMX5Q2GM562RJGOQG7GVCE5H3HIB4V');
INSERT INTO history_operations VALUES (12884922369, 12884922368, 1, 6, '{"limit": "922337203685.4775807", "trustee": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "trustor": "GARSFJNXJIHO6ULUBK3DBYKVSIZE7SC72S5DYBCHU7DKL22UXKVD7MXP", "asset_code": "AAA", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GARSFJNXJIHO6ULUBK3DBYKVSIZE7SC72S5DYBCHU7DKL22UXKVD7MXP');
INSERT INTO history_operations VALUES (12884926465, 12884926464, 1, 6, '{"limit": "922337203685.4775807", "trustee": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "trustor": "GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL", "asset_code": "EUR", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL');
INSERT INTO history_operations VALUES (12884930561, 12884930560, 1, 6, '{"limit": "922337203685.4775807", "trustee": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "trustor": "GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL", "asset_code": "1", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL');
INSERT INTO history_operations VALUES (12884934657, 12884934656, 1, 6, '{"limit": "922337203685.4775807", "trustee": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "trustor": "GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL", "asset_code": "21", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL');
INSERT INTO history_operations VALUES (12884938753, 12884938752, 1, 6, '{"limit": "922337203685.4775807", "trustee": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "trustor": "GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL", "asset_code": "22", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL');
INSERT INTO history_operations VALUES (12884942849, 12884942848, 1, 6, '{"limit": "922337203685.4775807", "trustee": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "trustor": "GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL", "asset_code": "31", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL');
INSERT INTO history_operations VALUES (12884946945, 12884946944, 1, 6, '{"limit": "922337203685.4775807", "trustee": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "trustor": "GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL", "asset_code": "32", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL');
INSERT INTO history_operations VALUES (12884951041, 12884951040, 1, 6, '{"limit": "922337203685.4775807", "trustee": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "trustor": "GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL", "asset_code": "33", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL');
INSERT INTO history_operations VALUES (12884955137, 12884955136, 1, 6, '{"limit": "922337203685.4775807", "trustee": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "trustor": "GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL", "asset_code": "AAA", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL');
INSERT INTO history_operations VALUES (12884959233, 12884959232, 1, 6, '{"limit": "922337203685.4775807", "trustee": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "trustor": "GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL", "asset_code": "BBB", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL');
INSERT INTO history_operations VALUES (12884963329, 12884963328, 1, 6, '{"limit": "922337203685.4775807", "trustee": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "trustor": "GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL", "asset_code": "CCC", "asset_type": "credit_alphanum4", "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"}', 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL');
INSERT INTO history_operations VALUES (8589938689, 8589938688, 1, 0, '{"funder": "GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H", "account": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN", "starting_balance": "10000.0000000"}', 'GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H');
INSERT INTO history_operations VALUES (8589942785, 8589942784, 1, 0, '{"funder": "GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H", "account": "GARSFJNXJIHO6ULUBK3DBYKVSIZE7SC72S5DYBCHU7DKL22UXKVD7MXP", "starting_balance": "1000.0000000"}', 'GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H');
INSERT INTO history_operations VALUES (8589946881, 8589946880, 1, 0, '{"funder": "GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H", "account": "GAEDTJ4PPEFVW5XV2S7LUXBEHNQMX5Q2GM562RJGOQG7GVCE5H3HIB4V", "starting_balance": "1000.0000000"}', 'GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H');
INSERT INTO history_operations VALUES (8589950977, 8589950976, 1, 0, '{"funder": "GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H", "account": "GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL", "starting_balance": "1000.0000000"}', 'GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H');


--
-- Data for Name: history_trades; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: history_transaction_participants; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO history_transaction_participants VALUES (1, 21474840576, 1);
INSERT INTO history_transaction_participants VALUES (2, 21474844672, 2);
INSERT INTO history_transaction_participants VALUES (3, 21474848768, 2);
INSERT INTO history_transaction_participants VALUES (4, 21474852864, 1);
INSERT INTO history_transaction_participants VALUES (5, 21474856960, 2);
INSERT INTO history_transaction_participants VALUES (6, 21474861056, 1);
INSERT INTO history_transaction_participants VALUES (7, 21474865152, 2);
INSERT INTO history_transaction_participants VALUES (8, 21474869248, 2);
INSERT INTO history_transaction_participants VALUES (9, 21474873344, 2);
INSERT INTO history_transaction_participants VALUES (10, 21474877440, 2);
INSERT INTO history_transaction_participants VALUES (11, 21474881536, 2);
INSERT INTO history_transaction_participants VALUES (12, 21474885632, 2);
INSERT INTO history_transaction_participants VALUES (13, 21474889728, 2);
INSERT INTO history_transaction_participants VALUES (14, 21474893824, 2);
INSERT INTO history_transaction_participants VALUES (15, 21474897920, 2);
INSERT INTO history_transaction_participants VALUES (16, 17179873280, 1);
INSERT INTO history_transaction_participants VALUES (17, 17179873280, 3);
INSERT INTO history_transaction_participants VALUES (18, 17179877376, 1);
INSERT INTO history_transaction_participants VALUES (19, 17179877376, 2);
INSERT INTO history_transaction_participants VALUES (20, 17179881472, 2);
INSERT INTO history_transaction_participants VALUES (21, 17179881472, 1);
INSERT INTO history_transaction_participants VALUES (22, 17179885568, 1);
INSERT INTO history_transaction_participants VALUES (23, 17179885568, 2);
INSERT INTO history_transaction_participants VALUES (24, 17179889664, 1);
INSERT INTO history_transaction_participants VALUES (25, 17179889664, 2);
INSERT INTO history_transaction_participants VALUES (26, 17179893760, 1);
INSERT INTO history_transaction_participants VALUES (27, 17179893760, 2);
INSERT INTO history_transaction_participants VALUES (28, 17179897856, 1);
INSERT INTO history_transaction_participants VALUES (29, 17179897856, 2);
INSERT INTO history_transaction_participants VALUES (30, 17179901952, 1);
INSERT INTO history_transaction_participants VALUES (31, 17179901952, 2);
INSERT INTO history_transaction_participants VALUES (32, 17179906048, 1);
INSERT INTO history_transaction_participants VALUES (33, 17179906048, 2);
INSERT INTO history_transaction_participants VALUES (34, 17179910144, 1);
INSERT INTO history_transaction_participants VALUES (35, 17179910144, 2);
INSERT INTO history_transaction_participants VALUES (36, 17179914240, 1);
INSERT INTO history_transaction_participants VALUES (37, 17179914240, 2);
INSERT INTO history_transaction_participants VALUES (38, 12884905984, 2);
INSERT INTO history_transaction_participants VALUES (39, 12884910080, 4);
INSERT INTO history_transaction_participants VALUES (40, 12884914176, 3);
INSERT INTO history_transaction_participants VALUES (41, 12884918272, 4);
INSERT INTO history_transaction_participants VALUES (42, 12884922368, 3);
INSERT INTO history_transaction_participants VALUES (43, 12884926464, 2);
INSERT INTO history_transaction_participants VALUES (44, 12884930560, 2);
INSERT INTO history_transaction_participants VALUES (45, 12884934656, 2);
INSERT INTO history_transaction_participants VALUES (46, 12884938752, 2);
INSERT INTO history_transaction_participants VALUES (47, 12884942848, 2);
INSERT INTO history_transaction_participants VALUES (48, 12884946944, 2);
INSERT INTO history_transaction_participants VALUES (49, 12884951040, 2);
INSERT INTO history_transaction_participants VALUES (50, 12884955136, 2);
INSERT INTO history_transaction_participants VALUES (51, 12884959232, 2);
INSERT INTO history_transaction_participants VALUES (52, 12884963328, 2);
INSERT INTO history_transaction_participants VALUES (53, 8589938688, 5);
INSERT INTO history_transaction_participants VALUES (54, 8589938688, 1);
INSERT INTO history_transaction_participants VALUES (55, 8589942784, 5);
INSERT INTO history_transaction_participants VALUES (56, 8589942784, 3);
INSERT INTO history_transaction_participants VALUES (57, 8589946880, 5);
INSERT INTO history_transaction_participants VALUES (58, 8589946880, 4);
INSERT INTO history_transaction_participants VALUES (59, 8589950976, 5);
INSERT INTO history_transaction_participants VALUES (60, 8589950976, 2);


--
-- Name: history_transaction_participants_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('history_transaction_participants_id_seq', 60, true);


--
-- Data for Name: history_transactions; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO history_transactions VALUES ('66687e8782d2b8be7b857aa9d807f8cbde4190fc61bd15873485d946a97d76cc', 5, 1, 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN', 8589934604, 100, 1, '2019-06-03 16:38:11.911268', '2019-06-03 16:38:11.911268', 21474840576, 'AAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAZAAAAAIAAAAMAAAAAAAAAAAAAAABAAAAAAAAAAMAAAABRVVSAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAFVU0QAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAX14QAAAAABAAAAAQAAAAAAAAAAAAAAAAAAAAH7AVp7AAAAQP/lenQGgG9KZUpqysX/k7eLvdjBqhOFsN6mi+PcTdOw61xYLICIKnzszsAGN/sk1s1la3lbjzPy3LkxhAYsEQ0=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAADAAAAAAAAAAAAAAAAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAEAAAABRVVSAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAFVU0QAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAX14QAAAAABAAAAAQAAAAAAAAAAAAAAAA==', 'AAAAAQAAAAIAAAADAAAABQAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbiiAAAAAIAAAALAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAABQAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbiiAAAAAIAAAAMAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAMAAAAFAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduKIAAAAAgAAAAwAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAAFAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduKIAAAAAgAAAAwAAAABAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFAAAAAgAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAAAAAABAAAAAUVVUgAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAABVVNEAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAAF9eEAAAAAAQAAAAEAAAAAAAAAAAAAAAA=', 'AAAAAgAAAAMAAAAEAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduO0AAAAAgAAAAsAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAAFAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduNQAAAAAgAAAAsAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{/+V6dAaAb0plSmrKxf+Tt4u92MGqE4Ww3qaL49xN07DrXFgsgIgqfOzOwAY3+yTWzWVreVuPM/LcuTGEBiwRDQ==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('ea241da3b384577ebd2eab456fdf0fee575096a7d235c4156daac895327f44a8', 5, 2, 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL', 8589934604, 100, 1, '2019-06-03 16:38:11.911663', '2019-06-03 16:38:11.911663', 21474844672, 'AAAAADTRcy67LrrAEIdRBSb8+zb1BhXTe0r3PHXXBOxKYEGnAAAAZAAAAAIAAAAMAAAAAAAAAAAAAAABAAAAAAAAAAMAAAABRVVSAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAFVU0QAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAX14QAAAAABAAAAAgAAAAAAAAAAAAAAAAAAAAFKYEGnAAAAQAaXv/gvoeKTm52lxwUYg6KAGvlVrljxWIK9elXp6ojMcOzkFs058weNIvNxbVyT/EcDnqZuHONkYjP3gQgqnQQ=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAADAAAAAAAAAAAAAAAAAAAAADTRcy67LrrAEIdRBSb8+zb1BhXTe0r3PHXXBOxKYEGnAAAAAAAAAAIAAAABRVVSAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAFVU0QAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAX14QAAAAABAAAAAgAAAAAAAAAAAAAAAA==', 'AAAAAQAAAAIAAAADAAAABQAAAAAAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAACVAvbBAAAAAIAAAALAAAACwAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAABQAAAAAAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAACVAvbBAAAAAIAAAAMAAAACwAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAABwAAAAMAAAAFAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC9sEAAAAAgAAAAwAAAALAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAAFAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC9sEAAAAAgAAAAwAAAAMAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFAAAAAgAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAAAAAACAAAAAUVVUgAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAABVVNEAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAAF9eEAAAAAAQAAAAIAAAAAAAAAAAAAAAAAAAADAAAABAAAAAEAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAABRVVSAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAukO3QAf/////////8AAAABAAAAAAAAAAAAAAABAAAABQAAAAEAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAABRVVSAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAukO3QAf/////////8AAAABAAAAAQAAAAAAAAAAAAAAAAX14QAAAAAAAAAAAAAAAAMAAAADAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAFVU0QAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAAAAAAEAAAAFAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAFVU0QAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAABAAAAAAL68IAAAAAAAAAAAAAAAAAAAAAA', 'AAAAAgAAAAMAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC9+0AAAAAgAAAAsAAAALAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAAFAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC99QAAAAAgAAAAsAAAALAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{Bpe/+C+h4pObnaXHBRiDooAa+VWuWPFYgr16VenqiMxw7OQWzTnzB40i83FtXJP8RwOepm4c42RiM/eBCCqdBA==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('ca445d728775307838e20a7b08048daf95b05c7d029a7aeb23686dc7c13f7946', 5, 3, 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL', 8589934605, 100, 1, '2019-06-03 16:38:11.91196', '2019-06-03 16:38:11.91196', 21474848768, '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', '{PXUi0i71YIhj5VhFHyhMsH0/L9YyrQgqja/UhjHSJYfleRN3SSkho2QEPOY67YGv1f3nrSv037UjY4rj/yeWBQ==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('031f714b398d3978c1a4bf3d787439b051bffebfe4b2aee671837f3391fb5154', 5, 4, 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN', 8589934605, 100, 1, '2019-06-03 16:38:11.912122', '2019-06-03 16:38:11.912123', 21474852864, 'AAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAZAAAAAIAAAANAAAAAAAAAAAAAAABAAAAAAAAAAMAAAABRVVSAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAFVU0QAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAX14QAAAAABAAAAAgAAAAAAAAAAAAAAAAAAAAH7AVp7AAAAQJyxsaU+sucl6bKcIF7NGFUrUAOCaoQ/VWe2u+4O1fUPzYYuXYnk5NbyMR62i1khschZfM/lgCH8MxANgPiJOw8=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAADAAAAAAAAAAAAAAAAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAQAAAABRVVSAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAFVU0QAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAX14QAAAAABAAAAAgAAAAAAAAAAAAAAAA==', 'AAAAAQAAAAIAAAADAAAABQAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbiiAAAAAIAAAAMAAAAAQAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAABQAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbiiAAAAAIAAAANAAAAAQAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAMAAAAFAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduKIAAAAAgAAAA0AAAABAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAAFAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduKIAAAAAgAAAA0AAAACAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFAAAAAgAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAAAAAAEAAAAAUVVUgAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAABVVNEAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAAF9eEAAAAAAQAAAAIAAAAAAAAAAAAAAAA=', 'AAAAAgAAAAMAAAAFAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduNQAAAAAgAAAAsAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAAFAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduLsAAAAAgAAAAsAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{nLGxpT6y5yXpspwgXs0YVStQA4JqhD9VZ7a77g7V9Q/Nhi5dieTk1vIxHraLWSGxyFl8z+WAIfwzEA2A+Ik7Dw==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('626e306845586c8214e50d749cca97049fb268ea44b798620cf467ac8cbf0aee', 5, 5, 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL', 8589934606, 100, 1, '2019-06-03 16:38:11.912302', '2019-06-03 16:38:11.912303', 21474856960, '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', '{qfjHnjQ7+vcARozEzqyatzW7xl5NRupVa3MXm1q7t9Karp6FbU0aR0jpOTMPWbb6WS0Jh3Xo5fyoT9D1ZKGsDw==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('70408d14e85d343cf95853595c645a75f6b5dd756f9ddba4c17f817529742cd6', 5, 6, 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN', 8589934606, 100, 1, '2019-06-03 16:38:11.912431', '2019-06-03 16:38:11.912431', 21474861056, 'AAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAZAAAAAIAAAAOAAAAAAAAAAAAAAABAAAAAAAAAAMAAAAAAAAAAVVTRAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAACVAvkAAAAAAEAAAAKAAAAAAAAAAAAAAAAAAAAAfsBWnsAAABAW43U2pQcr7t+SyEZQ5LSDUzq4H/7CXpEF8G7dcphO0YglPinQCVcce8g5JeiNj8yn3cvL+q6/LuI+Lh9ax8TBw==', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAADAAAAAAAAAAAAAAAAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAYAAAAAAAAAAVVTRAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAACVAvkAAAAAAEAAAAKAAAAAAAAAAAAAAAA', 'AAAAAQAAAAIAAAADAAAABQAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbiiAAAAAIAAAANAAAAAgAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAABQAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbiiAAAAAIAAAAOAAAAAgAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAMAAAAFAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduKIAAAAAgAAAA4AAAACAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAAFAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduKIAAAAAgAAAA4AAAADAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAABAAAAAAAAAAAAAAACVAvkAAAAAAAAAAAAAAAAAAAAAAUAAAACAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAYAAAAAAAAAAVVTRAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAACVAvkAAAAAAEAAAAKAAAAAAAAAAAAAAAA', 'AAAAAgAAAAMAAAAFAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduLsAAAAAgAAAAsAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAAFAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduKIAAAAAgAAAAsAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{W43U2pQcr7t+SyEZQ5LSDUzq4H/7CXpEF8G7dcphO0YglPinQCVcce8g5JeiNj8yn3cvL+q6/LuI+Lh9ax8TBw==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('bf576e57ca2318c2a1588626ae32a3fa82d224dc25734fbd8eeb25a46aedb562', 5, 7, 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL', 8589934607, 100, 1, '2019-06-03 16:38:11.912682', '2019-06-03 16:38:11.912682', 21474865152, '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', '{jgCio0VRbVLvhUTL9IvtDiiq26/WWkZUd+JKYNdC57h5rUHFGJlTJaZOXMOHHP+bN9zrewHDdfAiYEyikfSNDA==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('7607b7d68147007a5c52c0b21a3fabca1cea2a7568902ae04666a64f339f008c', 5, 8, 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL', 8589934608, 100, 1, '2019-06-03 16:38:11.912884', '2019-06-03 16:38:11.912884', 21474869248, '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', '{9YWqlVxuQPBgCRt3KQvEkOWa2PxsGtG2nnYyXeHVmi++reAT3QD9cw4Z9fu5mXs+9a6U5qq1/ce68g5127jQDw==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('9bbd809c9538c71a96d909f13f8c8df87b81dacc15c007c80905b462eca3863b', 5, 9, 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL', 8589934609, 100, 1, '2019-06-03 16:38:11.913037', '2019-06-03 16:38:11.913038', 21474873344, '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', '{ryW0hp3NwrlZqJav98WDrarFLhIrYTu2LC1sFkEQzR8chyQAhflfG8ZpXuefim149C9l1cKUKVrXApFK1terAQ==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('60b2a3b58a61ce915af20dc41d361d33190c95c8bee9a96cf2eaa1600429ae02', 5, 10, 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL', 8589934610, 100, 1, '2019-06-03 16:38:11.913205', '2019-06-03 16:38:11.913205', 21474877440, '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', '{f6wTOt02bDVxovwkvHzKqcPFkjEw861AJ/nDFeWX6Xsuytx5aOTlHCEmvUFnkaxmrjrBTs6EQNHM2jfh8AOWBA==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('04c0d8b80c1638ff078eb31ecc7725cae30645afa4ce018a2668604dd7537a2f', 5, 11, 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL', 8589934611, 100, 1, '2019-06-03 16:38:11.91337', '2019-06-03 16:38:11.91337', 21474881536, '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', '{FN49KIb9dlaWqbdKZ/rhxCMXGVY7ydV8hccU89NMXxpee5PlLDOVmkK+0f/4XTum+A+KqzHtZOkS2BSmjkPiDw==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('242ad17993b95d60ee9535229b549ecbdcd9efe95dc35c227c1257da84809077', 5, 12, 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL', 8589934612, 100, 1, '2019-06-03 16:38:11.913534', '2019-06-03 16:38:11.913534', 21474885632, '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', '{MvrTCtWLWjEGbFsjy18QqVe9xjUiYXslAumldhkwZWswZaE9FobJXR8SXTUxLwZYglA0lRGIHBz/3cBvzPROCA==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('0c425218bbbc030a8ecd97ba1e6fd2870a99fccd51c24aab6a4f59be61b245f8', 5, 13, 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL', 8589934613, 100, 1, '2019-06-03 16:38:11.913687', '2019-06-03 16:38:11.913687', 21474889728, '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', '{ZN64qE+gtZNvTixFTKCtjxOTH/pscCVo0/vUkVuZLCjdIQMN3/Bf7auLg1k187sTLYtYWTY87QK5+IXK3YMoCA==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('b49c655f0c6f78f56aca78e3b60952452918045e8e371b964558836357ce6ed0', 5, 14, 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL', 8589934614, 100, 1, '2019-06-03 16:38:11.913832', '2019-06-03 16:38:11.913833', 21474893824, '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', '{Xjc+wUIyqOk+FtWDBh2GtPPaaACf622L9E46BH70UZuNTuOLtPGGr6kjueIoRS/fyQ91Qb4EVdpbEABe9x4ZCg==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('c0be568472fbca2b56607b6dac37acf64d4f9001f60a61a5af3005a50327b617', 5, 15, 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL', 8589934615, 100, 1, '2019-06-03 16:38:11.914012', '2019-06-03 16:38:11.914012', 21474897920, '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', '{fST5pg8afF/qIrfgBGWs9BTedJUF+k9v2ETz1jxbQuw848CK0OUMcRcJWU8dyZZCfHSOE0L33ghFoOIJNt61AA==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('a3fc459ac9baa9c327254ae6eb182fb60dd597ac3761159393de1564f59efa77', 4, 1, 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN', 8589934593, 100, 1, '2019-06-03 16:38:11.945398', '2019-06-03 16:38:11.945398', 17179873280, 'AAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAZAAAAAIAAAABAAAAAAAAAAAAAAABAAAAAAAAAAEAAAAAIyKlt0oO71F0CrYw4VWSMk/IX9S6PARHp8al61S6qj8AAAABVVNEAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAukO3QAAAAAAAAAAAH7AVp7AAAAQB0rZ3swcCS/5dUtKqTdrBYxUjI6b9C0K9bdynGsnTMDO4wLsV8IOjO+ZrSBXt9FUQHlTKgHRE5ydrNL2f4eRQM=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAABAAAAAAAAAAA=', 'AAAAAQAAAAIAAAADAAAABAAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbjtAAAAAIAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAABAAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbjtAAAAAIAAAABAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAgAAAAMAAAADAAAAAQAAAAAjIqW3Sg7vUXQKtjDhVZIyT8hf1Lo8BEenxqXrVLqqPwAAAAFVU0QAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAAAAAAEAAAAEAAAAAQAAAAAjIqW3Sg7vUXQKtjDhVZIyT8hf1Lo8BEenxqXrVLqqPwAAAAFVU0QAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAC6Q7dAB//////////wAAAAEAAAAAAAAAAA==', 'AAAAAgAAAAMAAAACAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIdugAAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAAEAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduecAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{HStnezBwJL/l1S0qpN2sFjFSMjpv0LQr1t3KcaydMwM7jAuxXwg6M75mtIFe30VRAeVMqAdETnJ2s0vZ/h5FAw==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('9dc1746a8a1fae0abf8756d189c3d421bb5b5a0d5c0f8a3834da6d6aae8311c7', 4, 2, 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN', 8589934594, 100, 1, '2019-06-03 16:38:11.945666', '2019-06-03 16:38:11.945667', 17179877376, 'AAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAZAAAAAIAAAACAAAAAAAAAAAAAAABAAAAAAAAAAEAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAABRVVSAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAukO3QAAAAAAAAAAAH7AVp7AAAAQC5RCYrBzLMJMo9N+MYClr/hpgXNaLNAOBtG59qk+qyavtT5Euhsa+X+ukdhYkJrWJ+Hh4ngX/1WO1SpE8q/Bws=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAABAAAAAAAAAAA=', 'AAAAAQAAAAIAAAADAAAABAAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbjtAAAAAIAAAABAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAABAAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbjtAAAAAIAAAACAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAgAAAAMAAAADAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAFFVVIAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAAAAAAEAAAAEAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAFFVVIAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAC6Q7dAB//////////wAAAAEAAAAAAAAAAA==', 'AAAAAgAAAAMAAAAEAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduecAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAAEAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduc4AAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{LlEJisHMswkyj034xgKWv+GmBc1os0A4G0bn2qT6rJq+1PkS6Gxr5f66R2FiQmtYn4eHieBf/VY7VKkTyr8HCw==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('335970f27c1141c865b9a77189ffe71591d13c49c8d62d57239228562bc43cf2', 4, 3, 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN', 8589934595, 100, 1, '2019-06-03 16:38:11.94582', '2019-06-03 16:38:11.94582', 17179881472, 'AAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAZAAAAAIAAAADAAAAAAAAAAAAAAABAAAAAAAAAAEAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAABMQAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAukO3QAAAAAAAAAAAH7AVp7AAAAQJwPDiB+yuKFlFstdLMotpJlApsm34y8aJSlr1PjJEUi6fz6FucRMktKYHhJMINqUGzzm0qEwHesc6Uo1wwRygc=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAABAAAAAAAAAAA=', 'AAAAAQAAAAIAAAADAAAABAAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbjtAAAAAIAAAACAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAABAAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbjtAAAAAIAAAADAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAgAAAAMAAAADAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAExAAAAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAAAAAAEAAAAEAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAExAAAAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAC6Q7dAB//////////wAAAAEAAAAAAAAAAA==', 'AAAAAgAAAAMAAAAEAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduc4AAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAAEAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIdubUAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{nA8OIH7K4oWUWy10syi2kmUCmybfjLxolKWvU+MkRSLp/PoW5xEyS0pgeEkwg2pQbPObSoTAd6xzpSjXDBHKBw==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('63fd31091c6a3333a8ee23297c255e1d899a275519c48defd3ce0708a904c12b', 4, 4, 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN', 8589934596, 100, 1, '2019-06-03 16:38:11.946024', '2019-06-03 16:38:11.946024', 17179885568, 'AAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAZAAAAAIAAAAEAAAAAAAAAAAAAAABAAAAAAAAAAEAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAABMjEAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAukO3QAAAAAAAAAAAH7AVp7AAAAQBqkEspJ7xLLBgXnC77mo+US4DBjZsB89rAXQO3zmcyH5qEc4nf/D4DMS8R1nHf8c8staaUV3/S1dpr8AR93CQA=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAABAAAAAAAAAAA=', 'AAAAAQAAAAIAAAADAAAABAAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbjtAAAAAIAAAADAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAABAAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbjtAAAAAIAAAAEAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAgAAAAMAAAADAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAEyMQAAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAAAAAAEAAAAEAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAEyMQAAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAC6Q7dAB//////////wAAAAEAAAAAAAAAAA==', 'AAAAAgAAAAMAAAAEAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIdubUAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAAEAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduZwAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{GqQSyknvEssGBecLvuaj5RLgMGNmwHz2sBdA7fOZzIfmoRzid/8PgMxLxHWcd/xzyy1ppRXf9LV2mvwBH3cJAA==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('9b28e1c8e857b0efd958c70c36d777f36e053d4f0d2073d3276c18cfc85f9249', 4, 5, 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN', 8589934597, 100, 1, '2019-06-03 16:38:11.946158', '2019-06-03 16:38:11.946158', 17179889664, 'AAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAZAAAAAIAAAAFAAAAAAAAAAAAAAABAAAAAAAAAAEAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAABMjIAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAukO3QAAAAAAAAAAAH7AVp7AAAAQPtRQAtUfiEmaZWw4ux/nbuh8y8NoG8mBwPEyFcqXHD2+/LrQ5TA6OUcovIB48cu5wlE8S23KIfYK4nE1clR+gw=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAABAAAAAAAAAAA=', 'AAAAAQAAAAIAAAADAAAABAAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbjtAAAAAIAAAAEAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAABAAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbjtAAAAAIAAAAFAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAgAAAAMAAAADAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAEyMgAAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAAAAAAEAAAAEAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAEyMgAAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAC6Q7dAB//////////wAAAAEAAAAAAAAAAA==', 'AAAAAgAAAAMAAAAEAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduZwAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAAEAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduYMAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{+1FAC1R+ISZplbDi7H+du6HzLw2gbyYHA8TIVypccPb78utDlMDo5Ryi8gHjxy7nCUTxLbcoh9gricTVyVH6DA==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('8208aa13fc4d71b7ae079072b3a3af1bfd8806ba6fd354d3fcbd127b7c0fe232', 4, 6, 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN', 8589934598, 100, 1, '2019-06-03 16:38:11.946294', '2019-06-03 16:38:11.946294', 17179893760, 'AAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAZAAAAAIAAAAGAAAAAAAAAAAAAAABAAAAAAAAAAEAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAABMzEAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAukO3QAAAAAAAAAAAH7AVp7AAAAQMzLClp/m92NuPEoN1e7F87rGGHpCoXCIrw8t9cd8tC3ukKYQViHyj2F/X0Mu83o/ooo5cXpxLyn9IbZZXt+CgU=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAABAAAAAAAAAAA=', 'AAAAAQAAAAIAAAADAAAABAAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbjtAAAAAIAAAAFAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAABAAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbjtAAAAAIAAAAGAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAgAAAAMAAAADAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAEzMQAAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAAAAAAEAAAAEAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAEzMQAAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAC6Q7dAB//////////wAAAAEAAAAAAAAAAA==', 'AAAAAgAAAAMAAAAEAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduYMAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAAEAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduWoAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{zMsKWn+b3Y248Sg3V7sXzusYYekKhcIivDy31x3y0Le6QphBWIfKPYX9fQy7zej+iijlxenEvKf0htlle34KBQ==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('7c663b4fe86794c57b4f3d718b5048be1aa1e084e52f3468dd470829a057cfeb', 4, 7, 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN', 8589934599, 100, 1, '2019-06-03 16:38:11.946425', '2019-06-03 16:38:11.946425', 17179897856, 'AAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAZAAAAAIAAAAHAAAAAAAAAAAAAAABAAAAAAAAAAEAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAABMzIAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAukO3QAAAAAAAAAAAH7AVp7AAAAQPYL/L7JhjheV0aToVoB98Iuwa4Z3rcSsuf4zY0ndS8v2fy1mDzGOuS4n0Fwdcdhzy77qOdmFV3kSsR75YqLawY=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAABAAAAAAAAAAA=', 'AAAAAQAAAAIAAAADAAAABAAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbjtAAAAAIAAAAGAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAABAAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbjtAAAAAIAAAAHAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAgAAAAMAAAADAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAEzMgAAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAAAAAAEAAAAEAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAEzMgAAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAC6Q7dAB//////////wAAAAEAAAAAAAAAAA==', 'AAAAAgAAAAMAAAAEAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduWoAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAAEAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduVEAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{9gv8vsmGOF5XRpOhWgH3wi7BrhnetxKy5/jNjSd1Ly/Z/LWYPMY65LifQXB1x2HPLvuo52YVXeRKxHvliotrBg==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('57d54e294ca2d6206226a8950976bbec3852a05f977c895b68510e329ed1cf9b', 4, 8, 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN', 8589934600, 100, 1, '2019-06-03 16:38:11.947236', '2019-06-03 16:38:11.947236', 17179901952, 'AAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAZAAAAAIAAAAIAAAAAAAAAAAAAAABAAAAAAAAAAEAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAABMzMAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAukO3QAAAAAAAAAAAH7AVp7AAAAQDZLOY+f8aVbrWmcdsUz5KDWGikLBfYEESNyeF/4BlFppWpejMrS36byCy2yScN9zEhx5dZgVumSfp/bthUYbwM=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAABAAAAAAAAAAA=', 'AAAAAQAAAAIAAAADAAAABAAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbjtAAAAAIAAAAHAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAABAAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbjtAAAAAIAAAAIAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAgAAAAMAAAADAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAEzMwAAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAAAAAAEAAAAEAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAEzMwAAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAC6Q7dAB//////////wAAAAEAAAAAAAAAAA==', 'AAAAAgAAAAMAAAAEAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduVEAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAAEAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduTgAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{Nks5j5/xpVutaZx2xTPkoNYaKQsF9gQRI3J4X/gGUWmlal6MytLfpvILLbJJw33MSHHl1mBW6ZJ+n9u2FRhvAw==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('8791f2a4c7609b1e8aada156aae9ff29d72f46e2e491adba3c6e8d34e58ded7a', 4, 9, 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN', 8589934601, 100, 1, '2019-06-03 16:38:11.94762', '2019-06-03 16:38:11.947621', 17179906048, 'AAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAZAAAAAIAAAAJAAAAAAAAAAAAAAABAAAAAAAAAAEAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAABQUFBAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAukO3QAAAAAAAAAAAH7AVp7AAAAQB0hL+M53IFfa66sUCpmkalJ5PPiF1PUZ/U4/PO+CyssQwaAhcuQsuLtsJO5GJGJ9W0qfjd1yQHEoI5w/XP0CQw=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAABAAAAAAAAAAA=', 'AAAAAQAAAAIAAAADAAAABAAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbjtAAAAAIAAAAIAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAABAAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbjtAAAAAIAAAAJAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAgAAAAMAAAADAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAFBQUEAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAAAAAAEAAAAEAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAFBQUEAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAC6Q7dAB//////////wAAAAEAAAAAAAAAAA==', 'AAAAAgAAAAMAAAAEAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduTgAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAAEAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduR8AAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{HSEv4zncgV9rrqxQKmaRqUnk8+IXU9Rn9Tj8874LKyxDBoCFy5Cy4u2wk7kYkYn1bSp+N3XJAcSgjnD9c/QJDA==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('e134c8f8511bde00c3d8c42f742f10c28fa356bd73a5422d249a3e751e969a54', 4, 10, 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN', 8589934602, 100, 1, '2019-06-03 16:38:11.948031', '2019-06-03 16:38:11.948032', 17179910144, 'AAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAZAAAAAIAAAAKAAAAAAAAAAAAAAABAAAAAAAAAAEAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAABQkJCAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAukO3QAAAAAAAAAAAH7AVp7AAAAQLcF8nyUmEdlyn4NXHoTQKgW1vxDSpFdGN6MsYzCUqU7RRtKpElixkULS98nJxrTXkQalSNxnTLt1qgObBs5awA=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAABAAAAAAAAAAA=', 'AAAAAQAAAAIAAAADAAAABAAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbjtAAAAAIAAAAJAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAABAAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbjtAAAAAIAAAAKAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAgAAAAMAAAADAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAFCQkIAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAAAAAAEAAAAEAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAFCQkIAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAC6Q7dAB//////////wAAAAEAAAAAAAAAAA==', 'AAAAAgAAAAMAAAAEAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduR8AAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAAEAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduQYAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{twXyfJSYR2XKfg1cehNAqBbW/ENKkV0Y3oyxjMJSpTtFG0qkSWLGRQtL3ycnGtNeRBqVI3GdMu3WqA5sGzlrAA==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('bf87743ad238e79d244506e9335d6d27d31d6f7641275397749499065dc41d8c', 4, 11, 'GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN', 8589934603, 100, 1, '2019-06-03 16:38:11.948558', '2019-06-03 16:38:11.948559', 17179914240, 'AAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAZAAAAAIAAAALAAAAAAAAAAAAAAABAAAAAAAAAAEAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAABQ0NDAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAukO3QAAAAAAAAAAAH7AVp7AAAAQOvgQpBB2jnTH5XAsFX+hmHiWW7yDXADatkZ2pKuCCPxDRipc2jmMQmcCCuovne1eXa6QEAbnyOyYuxKEOfgPAE=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAABAAAAAAAAAAA=', 'AAAAAQAAAAIAAAADAAAABAAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbjtAAAAAIAAAAKAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAABAAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHbjtAAAAAIAAAALAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAgAAAAMAAAADAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAFDQ0MAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAAAAAAEAAAAEAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAFDQ0MAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAC6Q7dAB//////////wAAAAEAAAAAAAAAAA==', 'AAAAAgAAAAMAAAAEAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduQYAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAAEAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIduO0AAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{6+BCkEHaOdMflcCwVf6GYeJZbvINcANq2Rnakq4II/ENGKlzaOYxCZwIK6i+d7V5drpAQBufI7Ji7EoQ5+A8AQ==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('d76450086b21a849e31a168577a9c112a913c3317a3f0c1a80c6cb43e16cf348', 3, 1, 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL', 8589934593, 100, 1, '2019-06-03 16:38:11.969229', '2019-06-03 16:38:11.969229', 12884905984, 'AAAAADTRcy67LrrAEIdRBSb8+zb1BhXTe0r3PHXXBOxKYEGnAAAAZAAAAAIAAAABAAAAAAAAAAAAAAABAAAAAAAAAAYAAAABVVNEAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFae3//////////AAAAAAAAAAFKYEGnAAAAQC0FIi0RCGDQ0EuuBT5Kg2XxzHxLXRrCxYGj+hY7/sB2Y+JtWyWRjlq3DYL4ajSFE8Na1KKm42oM/gjZUx+1PQU=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAAGAAAAAAAAAAA=', 'AAAAAQAAAAIAAAADAAAAAwAAAAAAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAACVAvftAAAAAIAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAAAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAACVAvftAAAAAIAAAABAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAMAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC9+0AAAAAgAAAAEAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC9+0AAAAAgAAAAEAAAABAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAFVU0QAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAA==', 'AAAAAgAAAAMAAAACAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC+QAAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC+OcAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{LQUiLREIYNDQS64FPkqDZfHMfEtdGsLFgaP6Fjv+wHZj4m1bJZGOWrcNgvhqNIUTw1rUoqbjagz+CNlTH7U9BQ==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('cdb0ce2cef61c5cab6566c1c65c5bc632f943b2bd43cfdd80984a66994cb7484', 3, 2, 'GAEDTJ4PPEFVW5XV2S7LUXBEHNQMX5Q2GM562RJGOQG7GVCE5H3HIB4V', 8589934593, 100, 1, '2019-06-03 16:38:11.969408', '2019-06-03 16:38:11.969409', 12884910080, 'AAAAAAg5p495C1t29dS+ulwkO2DL9hozO+1FJnQN81RE6fZ0AAAAZAAAAAIAAAABAAAAAAAAAAAAAAABAAAAAAAAAAYAAAABRVVSAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFae3//////////AAAAAAAAAAFE6fZ0AAAAQNSm/BIxP9LwabXoBKigrTG85o/PUp6VOWh/ne6mMaT5hvehDUvbRHQghJ/SZTDfjD+FPPjbe3nzcJEun1EX4g0=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAAGAAAAAAAAAAA=', 'AAAAAQAAAAIAAAADAAAAAwAAAAAAAAAACDmnj3kLW3b11L66XCQ7YMv2GjM77UUmdA3zVETp9nQAAAACVAvjOAAAAAIAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAAAAAAACDmnj3kLW3b11L66XCQ7YMv2GjM77UUmdA3zVETp9nQAAAACVAvjOAAAAAIAAAABAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAMAAAADAAAAAAAAAAAIOaePeQtbdvXUvrpcJDtgy/YaMzvtRSZ0DfNUROn2dAAAAAJUC+M4AAAAAgAAAAEAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAAIOaePeQtbdvXUvrpcJDtgy/YaMzvtRSZ0DfNUROn2dAAAAAJUC+M4AAAAAgAAAAEAAAABAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAAAAAQAAAAAIOaePeQtbdvXUvrpcJDtgy/YaMzvtRSZ0DfNUROn2dAAAAAFFVVIAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAA==', 'AAAAAgAAAAMAAAACAAAAAAAAAAAIOaePeQtbdvXUvrpcJDtgy/YaMzvtRSZ0DfNUROn2dAAAAAJUC+QAAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAAIOaePeQtbdvXUvrpcJDtgy/YaMzvtRSZ0DfNUROn2dAAAAAJUC+OcAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{1Kb8EjE/0vBptegEqKCtMbzmj89SnpU5aH+d7qYxpPmG96ENS9tEdCCEn9JlMN+MP4U8+Nt7efNwkS6fURfiDQ==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('563ab1cf87e65377e015813499a6027081e63177a9ef35a0c5e0c0c2a5a918f7', 3, 3, 'GARSFJNXJIHO6ULUBK3DBYKVSIZE7SC72S5DYBCHU7DKL22UXKVD7MXP', 8589934593, 100, 1, '2019-06-03 16:38:11.96952', '2019-06-03 16:38:11.96952', 12884914176, 'AAAAACMipbdKDu9RdAq2MOFVkjJPyF/UujwER6fGpetUuqo/AAAAZAAAAAIAAAABAAAAAAAAAAAAAAABAAAAAAAAAAYAAAABVVNEAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFae3//////////AAAAAAAAAAFUuqo/AAAAQKblh64EFK4tp2+xohOkNaSdfMFDId/y4nop7dVmZRsbe4f/eVCpUrKQCRWLJ4bXzMpPTOTsjHRIxOa8WekP+ww=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAAGAAAAAAAAAAA=', 'AAAAAQAAAAIAAAADAAAAAwAAAAAAAAAAIyKlt0oO71F0CrYw4VWSMk/IX9S6PARHp8al61S6qj8AAAACVAvjOAAAAAIAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAAAAAAAIyKlt0oO71F0CrYw4VWSMk/IX9S6PARHp8al61S6qj8AAAACVAvjOAAAAAIAAAABAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAMAAAADAAAAAAAAAAAjIqW3Sg7vUXQKtjDhVZIyT8hf1Lo8BEenxqXrVLqqPwAAAAJUC+M4AAAAAgAAAAEAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAAjIqW3Sg7vUXQKtjDhVZIyT8hf1Lo8BEenxqXrVLqqPwAAAAJUC+M4AAAAAgAAAAEAAAABAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAAAAAQAAAAAjIqW3Sg7vUXQKtjDhVZIyT8hf1Lo8BEenxqXrVLqqPwAAAAFVU0QAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAA==', 'AAAAAgAAAAMAAAACAAAAAAAAAAAjIqW3Sg7vUXQKtjDhVZIyT8hf1Lo8BEenxqXrVLqqPwAAAAJUC+QAAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAAjIqW3Sg7vUXQKtjDhVZIyT8hf1Lo8BEenxqXrVLqqPwAAAAJUC+OcAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{puWHrgQUri2nb7GiE6Q1pJ18wUMh3/Lieint1WZlGxt7h/95UKlSspAJFYsnhtfMyk9M5OyMdEjE5rxZ6Q/7DA==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('8e8608cc05a772eeb8762e4658dfcd6a6b3d28db43c258df24c7c29038d6f007', 3, 4, 'GAEDTJ4PPEFVW5XV2S7LUXBEHNQMX5Q2GM562RJGOQG7GVCE5H3HIB4V', 8589934594, 100, 1, '2019-06-03 16:38:11.969639', '2019-06-03 16:38:11.969639', 12884918272, 'AAAAAAg5p495C1t29dS+ulwkO2DL9hozO+1FJnQN81RE6fZ0AAAAZAAAAAIAAAACAAAAAAAAAAAAAAABAAAAAAAAAAYAAAABQ0NDAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFae3//////////AAAAAAAAAAFE6fZ0AAAAQC/jr177vmkPeAUf6RY2BTktAqXgxddTOC5/L5FkCN+2w7PYW/c1zb6ckJa6mGEa+HZokL62wYjp1aT/DSiWdgo=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAAGAAAAAAAAAAA=', 'AAAAAQAAAAIAAAADAAAAAwAAAAAAAAAACDmnj3kLW3b11L66XCQ7YMv2GjM77UUmdA3zVETp9nQAAAACVAvjOAAAAAIAAAABAAAAAQAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAAAAAAACDmnj3kLW3b11L66XCQ7YMv2GjM77UUmdA3zVETp9nQAAAACVAvjOAAAAAIAAAACAAAAAQAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAMAAAADAAAAAAAAAAAIOaePeQtbdvXUvrpcJDtgy/YaMzvtRSZ0DfNUROn2dAAAAAJUC+M4AAAAAgAAAAIAAAABAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAAIOaePeQtbdvXUvrpcJDtgy/YaMzvtRSZ0DfNUROn2dAAAAAJUC+M4AAAAAgAAAAIAAAACAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAAAAAQAAAAAIOaePeQtbdvXUvrpcJDtgy/YaMzvtRSZ0DfNUROn2dAAAAAFDQ0MAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAA==', 'AAAAAgAAAAMAAAADAAAAAAAAAAAIOaePeQtbdvXUvrpcJDtgy/YaMzvtRSZ0DfNUROn2dAAAAAJUC+OcAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAAIOaePeQtbdvXUvrpcJDtgy/YaMzvtRSZ0DfNUROn2dAAAAAJUC+M4AAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{L+OvXvu+aQ94BR/pFjYFOS0CpeDF11M4Ln8vkWQI37bDs9hb9zXNvpyQlrqYYRr4dmiQvrbBiOnVpP8NKJZ2Cg==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('ae11ad12c012e34f4c660f28d4dbb1df40a354af1739d5e9581d3964427e712d', 3, 5, 'GARSFJNXJIHO6ULUBK3DBYKVSIZE7SC72S5DYBCHU7DKL22UXKVD7MXP', 8589934594, 100, 1, '2019-06-03 16:38:11.969745', '2019-06-03 16:38:11.969746', 12884922368, 'AAAAACMipbdKDu9RdAq2MOFVkjJPyF/UujwER6fGpetUuqo/AAAAZAAAAAIAAAACAAAAAAAAAAAAAAABAAAAAAAAAAYAAAABQUFBAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFae3//////////AAAAAAAAAAFUuqo/AAAAQBb9GRRvLUwbFHEKoehfAFNy9OieTI5pu5S+7P6s1voFRbvxDEbOTjjSlx0b6ThhUcSyVnwcqDNx82qB8PSMkwo=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAAGAAAAAAAAAAA=', 'AAAAAQAAAAIAAAADAAAAAwAAAAAAAAAAIyKlt0oO71F0CrYw4VWSMk/IX9S6PARHp8al61S6qj8AAAACVAvjOAAAAAIAAAABAAAAAQAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAAAAAAAIyKlt0oO71F0CrYw4VWSMk/IX9S6PARHp8al61S6qj8AAAACVAvjOAAAAAIAAAACAAAAAQAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAMAAAADAAAAAAAAAAAjIqW3Sg7vUXQKtjDhVZIyT8hf1Lo8BEenxqXrVLqqPwAAAAJUC+M4AAAAAgAAAAIAAAABAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAAjIqW3Sg7vUXQKtjDhVZIyT8hf1Lo8BEenxqXrVLqqPwAAAAJUC+M4AAAAAgAAAAIAAAACAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAAAAAQAAAAAjIqW3Sg7vUXQKtjDhVZIyT8hf1Lo8BEenxqXrVLqqPwAAAAFBQUEAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAA==', 'AAAAAgAAAAMAAAADAAAAAAAAAAAjIqW3Sg7vUXQKtjDhVZIyT8hf1Lo8BEenxqXrVLqqPwAAAAJUC+OcAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAAjIqW3Sg7vUXQKtjDhVZIyT8hf1Lo8BEenxqXrVLqqPwAAAAJUC+M4AAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{Fv0ZFG8tTBsUcQqh6F8AU3L06J5Mjmm7lL7s/qzW+gVFu/EMRs5OONKXHRvpOGFRxLJWfByoM3HzaoHw9IyTCg==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('ffe9db595b5e8be4d54b738e78652704dd96ad8729bc6b6adc17c300162f2c98', 3, 6, 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL', 8589934594, 100, 1, '2019-06-03 16:38:11.969854', '2019-06-03 16:38:11.969854', 12884926464, 'AAAAADTRcy67LrrAEIdRBSb8+zb1BhXTe0r3PHXXBOxKYEGnAAAAZAAAAAIAAAACAAAAAAAAAAAAAAABAAAAAAAAAAYAAAABRVVSAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFae3//////////AAAAAAAAAAFKYEGnAAAAQNdoCZqF5REF4OZCu65uUph5WUYmUKAcNCwgaJe4Mbakx92kViEe9kaJ13EhVOP7U+yjno/L1v1wpTaTa7YDiAc=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAAGAAAAAAAAAAA=', 'AAAAAQAAAAIAAAADAAAAAwAAAAAAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAACVAvftAAAAAIAAAABAAAAAQAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAAAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAACVAvftAAAAAIAAAACAAAAAQAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAMAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC9+0AAAAAgAAAAIAAAABAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC9+0AAAAAgAAAAIAAAACAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAFFVVIAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAA==', 'AAAAAgAAAAMAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC+OcAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC+M4AAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{12gJmoXlEQXg5kK7rm5SmHlZRiZQoBw0LCBol7gxtqTH3aRWIR72RonXcSFU4/tT7KOej8vW/XClNpNrtgOIBw==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('1a1dc052649f98314a12a17911d3868523d7806887d49a52749221b17cd713ce', 3, 7, 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL', 8589934595, 100, 1, '2019-06-03 16:38:11.969957', '2019-06-03 16:38:11.969957', 12884930560, 'AAAAADTRcy67LrrAEIdRBSb8+zb1BhXTe0r3PHXXBOxKYEGnAAAAZAAAAAIAAAADAAAAAAAAAAAAAAABAAAAAAAAAAYAAAABMQAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFae3//////////AAAAAAAAAAFKYEGnAAAAQMXWsXqylJxEXU0J+7PHlS/xBkw6/LXU23Q3sv7ew9z8KZZJbFEqDgrh340V57e/Xo1CIOdcRWHYpSZ0YZf5jg0=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAAGAAAAAAAAAAA=', 'AAAAAQAAAAIAAAADAAAAAwAAAAAAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAACVAvftAAAAAIAAAACAAAAAgAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAAAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAACVAvftAAAAAIAAAADAAAAAgAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAMAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC9+0AAAAAgAAAAMAAAACAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC9+0AAAAAgAAAAMAAAADAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAExAAAAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAA==', 'AAAAAgAAAAMAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC+M4AAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC+LUAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{xdaxerKUnERdTQn7s8eVL/EGTDr8tdTbdDey/t7D3PwplklsUSoOCuHfjRXnt79ejUIg51xFYdilJnRhl/mODQ==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('6bfdfd3aeff0f617532caf59ee28fe60b7c20ffa6574e08beecda9a890d8f2b2', 3, 8, 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL', 8589934596, 100, 1, '2019-06-03 16:38:11.970072', '2019-06-03 16:38:11.970072', 12884934656, 'AAAAADTRcy67LrrAEIdRBSb8+zb1BhXTe0r3PHXXBOxKYEGnAAAAZAAAAAIAAAAEAAAAAAAAAAAAAAABAAAAAAAAAAYAAAABMjEAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFae3//////////AAAAAAAAAAFKYEGnAAAAQI7tUQkz1Dedybs2Y9fYkjeZMkPxYJyQ21NNYmdl4rV3nI+pu3ymAiPtqrLWk21sXWSh8//1rx4sXH0aG+faQQ0=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAAGAAAAAAAAAAA=', 'AAAAAQAAAAIAAAADAAAAAwAAAAAAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAACVAvftAAAAAIAAAADAAAAAwAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAAAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAACVAvftAAAAAIAAAAEAAAAAwAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAMAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC9+0AAAAAgAAAAQAAAADAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC9+0AAAAAgAAAAQAAAAEAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAEyMQAAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAA==', 'AAAAAgAAAAMAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC+LUAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC+JwAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{ju1RCTPUN53JuzZj19iSN5kyQ/FgnJDbU01iZ2XitXecj6m7fKYCI+2qstaTbWxdZKHz//WvHixcfRob59pBDQ==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('b48686858ee595141fb7d5d9999d5d40b78cacc4e27092e48824c8a0ac01370e', 3, 9, 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL', 8589934597, 100, 1, '2019-06-03 16:38:11.970176', '2019-06-03 16:38:11.970176', 12884938752, 'AAAAADTRcy67LrrAEIdRBSb8+zb1BhXTe0r3PHXXBOxKYEGnAAAAZAAAAAIAAAAFAAAAAAAAAAAAAAABAAAAAAAAAAYAAAABMjIAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFae3//////////AAAAAAAAAAFKYEGnAAAAQKacYiD/dSVGNuCAcHUal15ITf0UzHebfNWNudVk5DPvnxFMOqXBkxAVOyVqqXKslFAtco/hI9WNB5yaf2i8eQU=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAAGAAAAAAAAAAA=', 'AAAAAQAAAAIAAAADAAAAAwAAAAAAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAACVAvftAAAAAIAAAAEAAAABAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAAAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAACVAvftAAAAAIAAAAFAAAABAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAMAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC9+0AAAAAgAAAAUAAAAEAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC9+0AAAAAgAAAAUAAAAFAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAEyMgAAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAA==', 'AAAAAgAAAAMAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC+JwAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC+IMAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{ppxiIP91JUY24IBwdRqXXkhN/RTMd5t81Y251WTkM++fEUw6pcGTEBU7JWqpcqyUUC1yj+Ej1Y0HnJp/aLx5BQ==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('e6315711266014d425b16b339bba063ad1b8d8823b5dfb8cd3399644754be459', 3, 10, 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL', 8589934598, 100, 1, '2019-06-03 16:38:11.970279', '2019-06-03 16:38:11.970279', 12884942848, 'AAAAADTRcy67LrrAEIdRBSb8+zb1BhXTe0r3PHXXBOxKYEGnAAAAZAAAAAIAAAAGAAAAAAAAAAAAAAABAAAAAAAAAAYAAAABMzEAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFae3//////////AAAAAAAAAAFKYEGnAAAAQI84TOk9J4ISBmnHAPy5KLkGnho/BjBWrZuYFB9I/6Z6QI9B25/mQw0EUOxBoX6X9q74a48TpIrf6n9LgonC8ws=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAAGAAAAAAAAAAA=', 'AAAAAQAAAAIAAAADAAAAAwAAAAAAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAACVAvftAAAAAIAAAAFAAAABQAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAAAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAACVAvftAAAAAIAAAAGAAAABQAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAMAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC9+0AAAAAgAAAAYAAAAFAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC9+0AAAAAgAAAAYAAAAGAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAEzMQAAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAA==', 'AAAAAgAAAAMAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC+IMAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC+GoAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{jzhM6T0nghIGaccA/LkouQaeGj8GMFatm5gUH0j/pnpAj0Hbn+ZDDQRQ7EGhfpf2rvhrjxOkit/qf0uCicLzCw==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('344a58add4f45245924cef37386866e0fd10f38f797615551c59fc0b5521e517', 3, 11, 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL', 8589934599, 100, 1, '2019-06-03 16:38:11.970383', '2019-06-03 16:38:11.970384', 12884946944, 'AAAAADTRcy67LrrAEIdRBSb8+zb1BhXTe0r3PHXXBOxKYEGnAAAAZAAAAAIAAAAHAAAAAAAAAAAAAAABAAAAAAAAAAYAAAABMzIAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFae3//////////AAAAAAAAAAFKYEGnAAAAQIYa/VDKchzI2ce0Hi7EZejOglDpZLsh3oLRGRkitrOM6zm9hJSZEfPZGPY/9+b14+fNnAyiHXGuVJK2ybiTJgU=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAAGAAAAAAAAAAA=', 'AAAAAQAAAAIAAAADAAAAAwAAAAAAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAACVAvftAAAAAIAAAAGAAAABgAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAAAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAACVAvftAAAAAIAAAAHAAAABgAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAMAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC9+0AAAAAgAAAAcAAAAGAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC9+0AAAAAgAAAAcAAAAHAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAEzMgAAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAA==', 'AAAAAgAAAAMAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC+GoAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC+FEAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{hhr9UMpyHMjZx7QeLsRl6M6CUOlkuyHegtEZGSK2s4zrOb2ElJkR89kY9j/35vXj582cDKIdca5UkrbJuJMmBQ==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('254f4bd646c27c730501f50bf5c9543b3510c10f8b39152b74a75cd78ec5aa7a', 3, 12, 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL', 8589934600, 100, 1, '2019-06-03 16:38:11.970486', '2019-06-03 16:38:11.970486', 12884951040, 'AAAAADTRcy67LrrAEIdRBSb8+zb1BhXTe0r3PHXXBOxKYEGnAAAAZAAAAAIAAAAIAAAAAAAAAAAAAAABAAAAAAAAAAYAAAABMzMAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFae3//////////AAAAAAAAAAFKYEGnAAAAQKAZjNrWsdz1Pve6US/r4vC01x+rDl62xDI2NxieKesGxMS+yIs4Rn39O6iWlApGqGJ5LlC7QRidIVWIaF8V6gA=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAAGAAAAAAAAAAA=', 'AAAAAQAAAAIAAAADAAAAAwAAAAAAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAACVAvftAAAAAIAAAAHAAAABwAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAAAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAACVAvftAAAAAIAAAAIAAAABwAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAMAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC9+0AAAAAgAAAAgAAAAHAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC9+0AAAAAgAAAAgAAAAIAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAEzMwAAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAA==', 'AAAAAgAAAAMAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC+FEAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC+DgAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{oBmM2tax3PU+97pRL+vi8LTXH6sOXrbEMjY3GJ4p6wbExL7IizhGff07qJaUCkaoYnkuULtBGJ0hVYhoXxXqAA==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('5b2de34cf9d44fb6bc9665ae2461e975264ca845d0c4bcba0570031fddca71a8', 3, 13, 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL', 8589934601, 100, 1, '2019-06-03 16:38:11.970587', '2019-06-03 16:38:11.970587', 12884955136, 'AAAAADTRcy67LrrAEIdRBSb8+zb1BhXTe0r3PHXXBOxKYEGnAAAAZAAAAAIAAAAJAAAAAAAAAAAAAAABAAAAAAAAAAYAAAABQUFBAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFae3//////////AAAAAAAAAAFKYEGnAAAAQBRmWJTg3iTl8d97BuJuLnpKhy1O+uCDJP2M7CKhF6e9QRUW3IsjS+r/mLpd2Gwvph5ISS8Q17O3QQrS/C/jGgc=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAAGAAAAAAAAAAA=', 'AAAAAQAAAAIAAAADAAAAAwAAAAAAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAACVAvftAAAAAIAAAAIAAAACAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAAAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAACVAvftAAAAAIAAAAJAAAACAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAMAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC9+0AAAAAgAAAAkAAAAIAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC9+0AAAAAgAAAAkAAAAJAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAFBQUEAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAA==', 'AAAAAgAAAAMAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC+DgAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC+B8AAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{FGZYlODeJOXx33sG4m4uekqHLU764IMk/YzsIqEXp71BFRbciyNL6v+Yul3YbC+mHkhJLxDXs7dBCtL8L+MaBw==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('07ce5e7469a0c0ed352e99d9adbae2baf86d9fcdb598407d47ce4671998aed5a', 3, 14, 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL', 8589934602, 100, 1, '2019-06-03 16:38:11.970712', '2019-06-03 16:38:11.970712', 12884959232, 'AAAAADTRcy67LrrAEIdRBSb8+zb1BhXTe0r3PHXXBOxKYEGnAAAAZAAAAAIAAAAKAAAAAAAAAAAAAAABAAAAAAAAAAYAAAABQkJCAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFae3//////////AAAAAAAAAAFKYEGnAAAAQOFkkm0Xf05YdBWq5Oi2o3ohXWHqakB9AFhQZaWkeHgVAlbRus962cOT/zkf+QYGba1Qk4nIfcI2Y9RT18SXQw4=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAAGAAAAAAAAAAA=', 'AAAAAQAAAAIAAAADAAAAAwAAAAAAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAACVAvftAAAAAIAAAAJAAAACQAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAAAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAACVAvftAAAAAIAAAAKAAAACQAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAMAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC9+0AAAAAgAAAAoAAAAJAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC9+0AAAAAgAAAAoAAAAKAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAFCQkIAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAA==', 'AAAAAgAAAAMAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC+B8AAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC+AYAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{4WSSbRd/Tlh0Fark6LajeiFdYepqQH0AWFBlpaR4eBUCVtG6z3rZw5P/OR/5BgZtrVCTich9wjZj1FPXxJdDDg==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('5ec32263bcc318abdc2b52844387c39856d9a15af175b8f121f3ff4848100a09', 3, 15, 'GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL', 8589934603, 100, 1, '2019-06-03 16:38:11.970822', '2019-06-03 16:38:11.970822', 12884963328, 'AAAAADTRcy67LrrAEIdRBSb8+zb1BhXTe0r3PHXXBOxKYEGnAAAAZAAAAAIAAAALAAAAAAAAAAAAAAABAAAAAAAAAAYAAAABQ0NDAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFae3//////////AAAAAAAAAAFKYEGnAAAAQFj+oqL3U4JBJvBjRlcgKq4N8Vd1kvHP9dRlhwF8Ptqhe3TPikyOQwhrYnBjrzOtJeehTgg9QDZ1R/Hf8zZEewQ=', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAAGAAAAAAAAAAA=', 'AAAAAQAAAAIAAAADAAAAAwAAAAAAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAACVAvftAAAAAIAAAAKAAAACgAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAAAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAACVAvftAAAAAIAAAALAAAACgAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAMAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC9+0AAAAAgAAAAsAAAAKAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC9+0AAAAAgAAAAsAAAALAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAAAAAQAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAFDQ0MAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAB//////////wAAAAEAAAAAAAAAAA==', 'AAAAAgAAAAMAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC+AYAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC9+0AAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{WP6iovdTgkEm8GNGVyAqrg3xV3WS8c/11GWHAXw+2qF7dM+KTI5DCGticGOvM60l56FOCD1ANnVH8d/zNkR7BA==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('4cecf4a146240111f6a035062dda70036d97008fb6bde073fe569463bcac9434', 2, 1, 'GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H', 1, 100, 1, '2019-06-03 16:38:11.987176', '2019-06-03 16:38:11.987176', 8589938688, 'AAAAAGL8HQvQkbK2HA3WVjRrKmjX00fG8sLI7m0ERwJW/AX3AAAAZAAAAAAAAAABAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAXSHboAAAAAAAAAAABVvwF9wAAAEALHgF/PRvl1Uxc06538IE+3POEzTvO5mKT5kCHdfb7kMCPzhZSpyJZ8xFyClYuBdKtdau9MFxVpdq347/q81MI', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAA=', 'AAAAAQAAAAAAAAABAAAAAwAAAAMAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtrOnY/5wAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtpxe7RZwAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAABdIdugAAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', 'AAAAAgAAAAMAAAABAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtrOnZAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtrOnY/+cAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{Cx4Bfz0b5dVMXNOud/CBPtzzhM07zuZik+ZAh3X2+5DAj84WUqciWfMRcgpWLgXSrXWrvTBcVaXat+O/6vNTCA==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('afc983f7f88809442fc616c1ce425b6c8cf6d8f7493b33ff6a809003a845ec16', 2, 2, 'GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H', 2, 100, 1, '2019-06-03 16:38:11.987336', '2019-06-03 16:38:11.987336', 8589942784, 'AAAAAGL8HQvQkbK2HA3WVjRrKmjX00fG8sLI7m0ERwJW/AX3AAAAZAAAAAAAAAACAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAIyKlt0oO71F0CrYw4VWSMk/IX9S6PARHp8al61S6qj8AAAACVAvkAAAAAAAAAAABVvwF9wAAAECXh9/+55FmQjeMx9IMQfBn42fYY1fKAT/gb+e7P3jdSMCKiKhwoxj1bub733a2XuxPETpe79uzatzm8/KI0asI', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAA=', 'AAAAAQAAAAAAAAABAAAAAwAAAAMAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtpxe7RZwAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtpoK4TJwAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAAAAAAAAAAAjIqW3Sg7vUXQKtjDhVZIyT8hf1Lo8BEenxqXrVLqqPwAAAAJUC+QAAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', 'AAAAAgAAAAMAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtrOnY/+cAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtrOnY/84AAAAAAAAAAIAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{l4ff/ueRZkI3jMfSDEHwZ+Nn2GNXygE/4G/nuz943UjAioiocKMY9W7m+992tl7sTxE6Xu/bs2rc5vPyiNGrCA==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('5c71f1b198a9be6b1c7e36737bb3e32c7f258530573bc11f20827d0913ba1433', 2, 3, 'GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H', 3, 100, 1, '2019-06-03 16:38:11.987456', '2019-06-03 16:38:11.987456', 8589946880, 'AAAAAGL8HQvQkbK2HA3WVjRrKmjX00fG8sLI7m0ERwJW/AX3AAAAZAAAAAAAAAADAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAACDmnj3kLW3b11L66XCQ7YMv2GjM77UUmdA3zVETp9nQAAAACVAvkAAAAAAAAAAABVvwF9wAAAEBdnUBKCV3HlHOqp5iNLsP8auaNCvxDeBp+0C+lwPYUNrzRALQRDJDWuKfExmsRrEnW8LKJbMdeW9ilLaEc2lAO', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAA=', 'AAAAAQAAAAAAAAABAAAAAwAAAAMAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtpoK4TJwAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtpe21U5wAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAAAAAAAAAAAIOaePeQtbdvXUvrpcJDtgy/YaMzvtRSZ0DfNUROn2dAAAAAJUC+QAAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', 'AAAAAgAAAAMAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtrOnY/84AAAAAAAAAAIAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtrOnY/7UAAAAAAAAAAMAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{XZ1ASgldx5RzqqeYjS7D/GrmjQr8Q3gaftAvpcD2FDa80QC0EQyQ1rinxMZrEaxJ1vCyiWzHXlvYpS2hHNpQDg==}', 'none', NULL, NULL, true, 100);
INSERT INTO history_transactions VALUES ('06445dc688fc2e583567e4a4940440d87cf4877a57af306ece7c64f1f061def4', 2, 4, 'GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H', 4, 100, 1, '2019-06-03 16:38:11.987566', '2019-06-03 16:38:11.987566', 8589950976, 'AAAAAGL8HQvQkbK2HA3WVjRrKmjX00fG8sLI7m0ERwJW/AX3AAAAZAAAAAAAAAAEAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAANNFzLrsuusAQh1EFJvz7NvUGFdN7Svc8ddcE7EpgQacAAAACVAvkAAAAAAAAAAABVvwF9wAAAEAfH3XJo+xZ8adtaA2PZZ3T0kM57CXfxGl+JItnC6nP9LziLQMKgl1EEhuiPqfjn14KK5WfEv0E0op5BJFAkd8O', 'AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAA=', 'AAAAAQAAAAAAAAABAAAAAwAAAAMAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtpe21U5wAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtpViyWpwAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAAAAAAAAAAA00XMuuy66wBCHUQUm/Ps29QYV03tK9zx11wTsSmBBpwAAAAJUC+QAAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', 'AAAAAgAAAAMAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtrOnY/7UAAAAAAAAAAMAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtrOnY/5wAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==', '{Hx91yaPsWfGnbWgNj2Wd09JDOewl38RpfiSLZwupz/S84i0DCoJdRBIboj6n459eCiuVnxL9BNKKeQSRQJHfDg==}', 'none', NULL, NULL, true, 100);


--
-- Name: asset_stats asset_stats_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY asset_stats
    ADD CONSTRAINT asset_stats_pkey PRIMARY KEY (id);


--
-- Name: gorp_migrations gorp_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY gorp_migrations
    ADD CONSTRAINT gorp_migrations_pkey PRIMARY KEY (id);


--
-- Name: history_assets history_assets_asset_code_asset_type_asset_issuer_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY history_assets
    ADD CONSTRAINT history_assets_asset_code_asset_type_asset_issuer_key UNIQUE (asset_code, asset_type, asset_issuer);


--
-- Name: history_assets history_assets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY history_assets
    ADD CONSTRAINT history_assets_pkey PRIMARY KEY (id);


--
-- Name: history_operation_participants history_operation_participants_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY history_operation_participants
    ADD CONSTRAINT history_operation_participants_pkey PRIMARY KEY (id);


--
-- Name: history_transaction_participants history_transaction_participants_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY history_transaction_participants
    ADD CONSTRAINT history_transaction_participants_pkey PRIMARY KEY (id);


--
-- Name: asset_by_code; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX asset_by_code ON history_assets USING btree (asset_code);


--
-- Name: asset_by_issuer; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX asset_by_issuer ON history_assets USING btree (asset_issuer);


--
-- Name: by_account; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX by_account ON history_transactions USING btree (account, account_sequence);


--
-- Name: by_hash; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX by_hash ON history_transactions USING btree (transaction_hash);


--
-- Name: by_ledger; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX by_ledger ON history_transactions USING btree (ledger_sequence, application_order);


--
-- Name: hist_e_by_order; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX hist_e_by_order ON history_effects USING btree (history_operation_id, "order");


--
-- Name: hist_e_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX hist_e_id ON history_effects USING btree (history_account_id, history_operation_id, "order");


--
-- Name: hist_op_p_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX hist_op_p_id ON history_operation_participants USING btree (history_account_id, history_operation_id);


--
-- Name: hist_tx_p_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX hist_tx_p_id ON history_transaction_participants USING btree (history_account_id, history_transaction_id);


--
-- Name: hop_by_hoid; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX hop_by_hoid ON history_operation_participants USING btree (history_operation_id);


--
-- Name: hs_ledger_by_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX hs_ledger_by_id ON history_ledgers USING btree (id);


--
-- Name: hs_transaction_by_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX hs_transaction_by_id ON history_transactions USING btree (id);


--
-- Name: htp_by_htid; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX htp_by_htid ON history_transaction_participants USING btree (history_transaction_id);


--
-- Name: htrd_by_base_account; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX htrd_by_base_account ON history_trades USING btree (base_account_id);


--
-- Name: htrd_by_base_offer; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX htrd_by_base_offer ON history_trades USING btree (base_offer_id);


--
-- Name: htrd_by_counter_account; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX htrd_by_counter_account ON history_trades USING btree (counter_account_id);


--
-- Name: htrd_by_counter_offer; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX htrd_by_counter_offer ON history_trades USING btree (counter_offer_id);


--
-- Name: htrd_by_offer; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX htrd_by_offer ON history_trades USING btree (offer_id);


--
-- Name: htrd_counter_lookup; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX htrd_counter_lookup ON history_trades USING btree (counter_asset_id);


--
-- Name: htrd_pair_time_lookup; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX htrd_pair_time_lookup ON history_trades USING btree (base_asset_id, counter_asset_id, ledger_closed_at);


--
-- Name: htrd_pid; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX htrd_pid ON history_trades USING btree (history_operation_id, "order");


--
-- Name: htrd_time_lookup; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX htrd_time_lookup ON history_trades USING btree (ledger_closed_at);


--
-- Name: index_history_accounts_on_address; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_history_accounts_on_address ON history_accounts USING btree (address);


--
-- Name: index_history_accounts_on_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_history_accounts_on_id ON history_accounts USING btree (id);


--
-- Name: index_history_effects_on_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_history_effects_on_type ON history_effects USING btree (type);


--
-- Name: index_history_ledgers_on_closed_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_history_ledgers_on_closed_at ON history_ledgers USING btree (closed_at);


--
-- Name: index_history_ledgers_on_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_history_ledgers_on_id ON history_ledgers USING btree (id);


--
-- Name: index_history_ledgers_on_importer_version; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_history_ledgers_on_importer_version ON history_ledgers USING btree (importer_version);


--
-- Name: index_history_ledgers_on_ledger_hash; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_history_ledgers_on_ledger_hash ON history_ledgers USING btree (ledger_hash);


--
-- Name: index_history_ledgers_on_previous_ledger_hash; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_history_ledgers_on_previous_ledger_hash ON history_ledgers USING btree (previous_ledger_hash);


--
-- Name: index_history_ledgers_on_sequence; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_history_ledgers_on_sequence ON history_ledgers USING btree (sequence);


--
-- Name: index_history_operations_on_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_history_operations_on_id ON history_operations USING btree (id);


--
-- Name: index_history_operations_on_transaction_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_history_operations_on_transaction_id ON history_operations USING btree (transaction_id);


--
-- Name: index_history_operations_on_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_history_operations_on_type ON history_operations USING btree (type);


--
-- Name: index_history_transactions_on_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_history_transactions_on_id ON history_transactions USING btree (id);


--
-- Name: trade_effects_by_order_book; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX trade_effects_by_order_book ON history_effects USING btree (((details ->> 'sold_asset_type'::text)), ((details ->> 'sold_asset_code'::text)), ((details ->> 'sold_asset_issuer'::text)), ((details ->> 'bought_asset_type'::text)), ((details ->> 'bought_asset_code'::text)), ((details ->> 'bought_asset_issuer'::text))) WHERE (type = 33);


--
-- Name: asset_stats asset_stats_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY asset_stats
    ADD CONSTRAINT asset_stats_id_fkey FOREIGN KEY (id) REFERENCES history_assets(id) ON UPDATE RESTRICT ON DELETE CASCADE;


--
-- Name: history_trades history_trades_base_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY history_trades
    ADD CONSTRAINT history_trades_base_account_id_fkey FOREIGN KEY (base_account_id) REFERENCES history_accounts(id);


--
-- Name: history_trades history_trades_base_asset_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY history_trades
    ADD CONSTRAINT history_trades_base_asset_id_fkey FOREIGN KEY (base_asset_id) REFERENCES history_assets(id);


--
-- Name: history_trades history_trades_counter_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY history_trades
    ADD CONSTRAINT history_trades_counter_account_id_fkey FOREIGN KEY (counter_account_id) REFERENCES history_accounts(id);


--
-- Name: history_trades history_trades_counter_asset_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY history_trades
    ADD CONSTRAINT history_trades_counter_asset_id_fkey FOREIGN KEY (counter_asset_id) REFERENCES history_assets(id);


--
-- PostgreSQL database dump complete
--

