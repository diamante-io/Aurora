--
-- PostgreSQL database dump
--

-- Dumped from database version 9.6.1
-- Dumped by pg_dump version 9.6.1

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET search_path = public, pg_catalog;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: gorp_migrations; Type: TABLE; Schema: public; Owner: bartek
--

CREATE TABLE gorp_migrations (
    id text NOT NULL,
    applied_at timestamp with time zone
);


ALTER TABLE gorp_migrations OWNER TO bartek;

--
-- Name: received_payment; Type: TABLE; Schema: public; Owner: bartek
--

CREATE TABLE received_payment (
    id bigint NOT NULL,
    operation_id character varying(255) NOT NULL,
    processed_at timestamp without time zone NOT NULL,
    paging_token character varying(255) NOT NULL,
    status character varying(255) NOT NULL,
    transaction_id character varying(64) DEFAULT 'N/A'::character varying
);


ALTER TABLE received_payment OWNER TO bartek;

--
-- Name: receivedpayment_id_seq; Type: SEQUENCE; Schema: public; Owner: bartek
--

CREATE SEQUENCE receivedpayment_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE receivedpayment_id_seq OWNER TO bartek;

--
-- Name: receivedpayment_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bartek
--

ALTER SEQUENCE receivedpayment_id_seq OWNED BY received_payment.id;


--
-- Name: sent_transaction; Type: TABLE; Schema: public; Owner: bartek
--

CREATE TABLE sent_transaction (
    id integer NOT NULL,
    transaction_id character varying(64) NOT NULL,
    status character varying(10) NOT NULL,
    source character varying(56) NOT NULL,
    submitted_at timestamp without time zone NOT NULL,
    succeeded_at timestamp without time zone,
    ledger bigint,
    envelope_xdr text NOT NULL,
    result_xdr character varying(255) DEFAULT NULL::character varying,
    payment_id character varying(255) DEFAULT NULL::character varying
);


ALTER TABLE sent_transaction OWNER TO bartek;

--
-- Name: senttransaction_id_seq; Type: SEQUENCE; Schema: public; Owner: bartek
--

CREATE SEQUENCE senttransaction_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE senttransaction_id_seq OWNER TO bartek;

--
-- Name: senttransaction_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bartek
--

ALTER SEQUENCE senttransaction_id_seq OWNED BY sent_transaction.id;


--
-- Name: received_payment id; Type: DEFAULT; Schema: public; Owner: bartek
--

ALTER TABLE ONLY received_payment ALTER COLUMN id SET DEFAULT nextval('receivedpayment_id_seq'::regclass);


--
-- Name: sent_transaction id; Type: DEFAULT; Schema: public; Owner: bartek
--

ALTER TABLE ONLY sent_transaction ALTER COLUMN id SET DEFAULT nextval('senttransaction_id_seq'::regclass);


--
-- Data for Name: gorp_migrations; Type: TABLE DATA; Schema: public; Owner: bartek
--

COPY gorp_migrations (id, applied_at) FROM stdin;
01_init.sql	2018-04-25 18:24:44.557981+02
02_payment_id.sql	2018-04-25 18:24:44.571645+02
03_transaction_id.sql	2018-04-25 18:24:44.578795+02
04_table_names.sql	2018-04-25 18:24:44.5814+02
\.


--
-- Data for Name: received_payment; Type: TABLE DATA; Schema: public; Owner: bartek
--

COPY received_payment (id, operation_id, processed_at, paging_token, status, transaction_id) FROM stdin;
\.


--
-- Name: receivedpayment_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bartek
--

SELECT pg_catalog.setval('receivedpayment_id_seq', 1, false);


--
-- Data for Name: sent_transaction; Type: TABLE DATA; Schema: public; Owner: bartek
--

COPY sent_transaction (id, transaction_id, status, source, submitted_at, succeeded_at, ledger, envelope_xdr, result_xdr, payment_id) FROM stdin;
\.


--
-- Name: senttransaction_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bartek
--

SELECT pg_catalog.setval('senttransaction_id_seq', 1, false);


--
-- Name: gorp_migrations gorp_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: bartek
--

ALTER TABLE ONLY gorp_migrations
    ADD CONSTRAINT gorp_migrations_pkey PRIMARY KEY (id);


--
-- Name: sent_transaction payment_id_unique; Type: CONSTRAINT; Schema: public; Owner: bartek
--

ALTER TABLE ONLY sent_transaction
    ADD CONSTRAINT payment_id_unique UNIQUE (payment_id);


--
-- Name: received_payment receivedpayment_operation_id_key; Type: CONSTRAINT; Schema: public; Owner: bartek
--

ALTER TABLE ONLY received_payment
    ADD CONSTRAINT receivedpayment_operation_id_key UNIQUE (operation_id);


--
-- Name: received_payment receivedpayment_pkey; Type: CONSTRAINT; Schema: public; Owner: bartek
--

ALTER TABLE ONLY received_payment
    ADD CONSTRAINT receivedpayment_pkey PRIMARY KEY (id);


--
-- Name: sent_transaction senttransaction_pkey; Type: CONSTRAINT; Schema: public; Owner: bartek
--

ALTER TABLE ONLY sent_transaction
    ADD CONSTRAINT senttransaction_pkey PRIMARY KEY (id);


--
-- PostgreSQL database dump complete
--

