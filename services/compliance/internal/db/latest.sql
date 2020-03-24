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
-- Name: allowed_fi; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE allowed_fi (
    id bigint NOT NULL,
    name character varying(255) NOT NULL,
    domain character varying(255) NOT NULL,
    public_key character(56) NOT NULL,
    allowed_at timestamp without time zone NOT NULL
);


ALTER TABLE allowed_fi OWNER TO root;

--
-- Name: allowed_user; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE allowed_user (
    id bigint NOT NULL,
    fi_name character varying(255) NOT NULL,
    fi_domain character varying(255) NOT NULL,
    fi_public_key character(56) NOT NULL,
    user_id character varying(255) NOT NULL,
    allowed_at timestamp without time zone NOT NULL
);


ALTER TABLE allowed_user OWNER TO root;

--
-- Name: allowedfi_id_seq; Type: SEQUENCE; Schema: public; Owner: root
--

CREATE SEQUENCE allowedfi_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE allowedfi_id_seq OWNER TO root;

--
-- Name: allowedfi_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: root
--

ALTER SEQUENCE allowedfi_id_seq OWNED BY allowed_fi.id;


--
-- Name: alloweduser_id_seq; Type: SEQUENCE; Schema: public; Owner: root
--

CREATE SEQUENCE alloweduser_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE alloweduser_id_seq OWNER TO root;

--
-- Name: alloweduser_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: root
--

ALTER SEQUENCE alloweduser_id_seq OWNED BY allowed_user.id;


--
-- Name: auth_data; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE auth_data (
    id bigint NOT NULL,
    request_id character varying(255) NOT NULL,
    domain character varying(255) NOT NULL,
    auth_data text NOT NULL
);


ALTER TABLE auth_data OWNER TO root;

--
-- Name: authdata_id_seq; Type: SEQUENCE; Schema: public; Owner: root
--

CREATE SEQUENCE authdata_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE authdata_id_seq OWNER TO root;

--
-- Name: authdata_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: root
--

ALTER SEQUENCE authdata_id_seq OWNED BY auth_data.id;


--
-- Name: authorized_transaction; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE authorized_transaction (
    id bigint NOT NULL,
    transaction_id character varying(64) NOT NULL,
    memo character varying(64) NOT NULL,
    transaction_xdr text NOT NULL,
    authorized_at timestamp without time zone NOT NULL,
    data text NOT NULL
);


ALTER TABLE authorized_transaction OWNER TO root;

--
-- Name: authorizedtransaction_id_seq; Type: SEQUENCE; Schema: public; Owner: root
--

CREATE SEQUENCE authorizedtransaction_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE authorizedtransaction_id_seq OWNER TO root;

--
-- Name: authorizedtransaction_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: root
--

ALTER SEQUENCE authorizedtransaction_id_seq OWNED BY authorized_transaction.id;


--
-- Name: gorp_migrations; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE gorp_migrations (
    id text NOT NULL,
    applied_at timestamp with time zone
);


ALTER TABLE gorp_migrations OWNER TO root;

--
-- Name: allowed_fi id; Type: DEFAULT; Schema: public; Owner: root
--

ALTER TABLE ONLY allowed_fi ALTER COLUMN id SET DEFAULT nextval('allowedfi_id_seq'::regclass);


--
-- Name: allowed_user id; Type: DEFAULT; Schema: public; Owner: root
--

ALTER TABLE ONLY allowed_user ALTER COLUMN id SET DEFAULT nextval('alloweduser_id_seq'::regclass);


--
-- Name: auth_data id; Type: DEFAULT; Schema: public; Owner: root
--

ALTER TABLE ONLY auth_data ALTER COLUMN id SET DEFAULT nextval('authdata_id_seq'::regclass);


--
-- Name: authorized_transaction id; Type: DEFAULT; Schema: public; Owner: root
--

ALTER TABLE ONLY authorized_transaction ALTER COLUMN id SET DEFAULT nextval('authorizedtransaction_id_seq'::regclass);


--
-- Data for Name: allowed_fi; Type: TABLE DATA; Schema: public; Owner: root
--

COPY allowed_fi (id, name, domain, public_key, allowed_at) FROM stdin;
\.


--
-- Data for Name: allowed_user; Type: TABLE DATA; Schema: public; Owner: root
--

COPY allowed_user (id, fi_name, fi_domain, fi_public_key, user_id, allowed_at) FROM stdin;
\.


--
-- Name: allowedfi_id_seq; Type: SEQUENCE SET; Schema: public; Owner: root
--

SELECT pg_catalog.setval('allowedfi_id_seq', 1, false);


--
-- Name: alloweduser_id_seq; Type: SEQUENCE SET; Schema: public; Owner: root
--

SELECT pg_catalog.setval('alloweduser_id_seq', 1, false);


--
-- Data for Name: auth_data; Type: TABLE DATA; Schema: public; Owner: root
--

COPY auth_data (id, request_id, domain, auth_data) FROM stdin;
\.


--
-- Name: authdata_id_seq; Type: SEQUENCE SET; Schema: public; Owner: root
--

SELECT pg_catalog.setval('authdata_id_seq', 1, false);


--
-- Data for Name: authorized_transaction; Type: TABLE DATA; Schema: public; Owner: root
--

COPY authorized_transaction (id, transaction_id, memo, transaction_xdr, authorized_at, data) FROM stdin;
\.


--
-- Name: authorizedtransaction_id_seq; Type: SEQUENCE SET; Schema: public; Owner: root
--

SELECT pg_catalog.setval('authorizedtransaction_id_seq', 1, false);


--
-- Data for Name: gorp_migrations; Type: TABLE DATA; Schema: public; Owner: root
--

COPY gorp_migrations (id, applied_at) FROM stdin;
01_init.sql	2018-06-29 15:12:55.594578+02
02_auth_data.sql	2018-06-29 15:12:55.601753+02
03_table_names.sql	2018-06-29 15:12:55.604754+02
\.


--
-- Name: allowed_fi allowedfi_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY allowed_fi
    ADD CONSTRAINT allowedfi_pkey PRIMARY KEY (id);


--
-- Name: allowed_user alloweduser_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY allowed_user
    ADD CONSTRAINT alloweduser_pkey PRIMARY KEY (id);


--
-- Name: auth_data authdata_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY auth_data
    ADD CONSTRAINT authdata_pkey PRIMARY KEY (id);


--
-- Name: authorized_transaction authorizedtransaction_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY authorized_transaction
    ADD CONSTRAINT authorizedtransaction_pkey PRIMARY KEY (id);


--
-- Name: gorp_migrations gorp_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY gorp_migrations
    ADD CONSTRAINT gorp_migrations_pkey PRIMARY KEY (id);


--
-- Name: afi_by_domain; Type: INDEX; Schema: public; Owner: root
--

CREATE UNIQUE INDEX afi_by_domain ON allowed_fi USING btree (domain);


--
-- Name: afi_by_public_key; Type: INDEX; Schema: public; Owner: root
--

CREATE UNIQUE INDEX afi_by_public_key ON allowed_fi USING btree (public_key);


--
-- Name: au_by_fi_public_key_user_id; Type: INDEX; Schema: public; Owner: root
--

CREATE UNIQUE INDEX au_by_fi_public_key_user_id ON allowed_user USING btree (fi_public_key, user_id);


--
-- Name: request_id; Type: INDEX; Schema: public; Owner: root
--

CREATE UNIQUE INDEX request_id ON auth_data USING btree (request_id);


--
-- PostgreSQL database dump complete
--

