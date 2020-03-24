--
-- PostgreSQL database dump
--

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

SET search_path = public, pg_catalog;

DROP INDEX public.signersaccount;
DROP INDEX public.sellingissuerindex;
DROP INDEX public.scpenvsbyseq;
DROP INDEX public.priceindex;
DROP INDEX public.ledgersbyseq;
DROP INDEX public.histfeebyseq;
DROP INDEX public.histbyseq;
DROP INDEX public.buyingissuerindex;
DROP INDEX public.accountbalances;
ALTER TABLE ONLY public.txhistory DROP CONSTRAINT txhistory_pkey;
ALTER TABLE ONLY public.txfeehistory DROP CONSTRAINT txfeehistory_pkey;
ALTER TABLE ONLY public.trustlines DROP CONSTRAINT trustlines_pkey;
ALTER TABLE ONLY public.storestate DROP CONSTRAINT storestate_pkey;
ALTER TABLE ONLY public.signers DROP CONSTRAINT signers_pkey;
ALTER TABLE ONLY public.scpquorums DROP CONSTRAINT scpquorums_pkey;
ALTER TABLE ONLY public.pubsub DROP CONSTRAINT pubsub_pkey;
ALTER TABLE ONLY public.publishqueue DROP CONSTRAINT publishqueue_pkey;
ALTER TABLE ONLY public.peers DROP CONSTRAINT peers_pkey;
ALTER TABLE ONLY public.offers DROP CONSTRAINT offers_pkey;
ALTER TABLE ONLY public.ledgerheaders DROP CONSTRAINT ledgerheaders_pkey;
ALTER TABLE ONLY public.ledgerheaders DROP CONSTRAINT ledgerheaders_ledgerseq_key;
ALTER TABLE ONLY public.accounts DROP CONSTRAINT accounts_pkey;
DROP TABLE public.txhistory;
DROP TABLE public.txfeehistory;
DROP TABLE public.trustlines;
DROP TABLE public.storestate;
DROP TABLE public.signers;
DROP TABLE public.scpquorums;
DROP TABLE public.scphistory;
DROP TABLE public.pubsub;
DROP TABLE public.publishqueue;
DROP TABLE public.peers;
DROP TABLE public.offers;
DROP TABLE public.ledgerheaders;
DROP TABLE public.accounts;
DROP EXTENSION plpgsql;
DROP SCHEMA public;
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

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: accounts; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE accounts (
    accountid character varying(56) NOT NULL,
    balance bigint NOT NULL,
    seqnum bigint NOT NULL,
    numsubentries integer NOT NULL,
    inflationdest character varying(56),
    homedomain character varying(32) NOT NULL,
    thresholds text NOT NULL,
    flags integer NOT NULL,
    lastmodified integer NOT NULL,
    CONSTRAINT accounts_balance_check CHECK ((balance >= 0)),
    CONSTRAINT accounts_numsubentries_check CHECK ((numsubentries >= 0))
);


--
-- Name: ledgerheaders; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE ledgerheaders (
    ledgerhash character(64) NOT NULL,
    prevhash character(64) NOT NULL,
    bucketlisthash character(64) NOT NULL,
    ledgerseq integer,
    closetime bigint NOT NULL,
    data text NOT NULL,
    CONSTRAINT ledgerheaders_closetime_check CHECK ((closetime >= 0)),
    CONSTRAINT ledgerheaders_ledgerseq_check CHECK ((ledgerseq >= 0))
);


--
-- Name: offers; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE offers (
    sellerid character varying(56) NOT NULL,
    offerid bigint NOT NULL,
    sellingassettype integer NOT NULL,
    sellingassetcode character varying(12),
    sellingissuer character varying(56),
    buyingassettype integer NOT NULL,
    buyingassetcode character varying(12),
    buyingissuer character varying(56),
    amount bigint NOT NULL,
    pricen integer NOT NULL,
    priced integer NOT NULL,
    price double precision NOT NULL,
    flags integer NOT NULL,
    lastmodified integer NOT NULL,
    CONSTRAINT offers_amount_check CHECK ((amount >= 0)),
    CONSTRAINT offers_offerid_check CHECK ((offerid >= 0))
);


--
-- Name: peers; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE peers (
    ip character varying(15) NOT NULL,
    port integer DEFAULT 0 NOT NULL,
    nextattempt timestamp without time zone NOT NULL,
    numfailures integer DEFAULT 0 NOT NULL,
    CONSTRAINT peers_numfailures_check CHECK ((numfailures >= 0)),
    CONSTRAINT peers_port_check CHECK (((port > 0) AND (port <= 65535)))
);


--
-- Name: publishqueue; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE publishqueue (
    ledger integer NOT NULL,
    state text
);


--
-- Name: pubsub; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE pubsub (
    resid character(32) NOT NULL,
    lastread integer
);


--
-- Name: scphistory; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE scphistory (
    nodeid character(56) NOT NULL,
    ledgerseq integer NOT NULL,
    envelope text NOT NULL,
    CONSTRAINT scphistory_ledgerseq_check CHECK ((ledgerseq >= 0))
);


--
-- Name: scpquorums; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE scpquorums (
    qsethash character(64) NOT NULL,
    lastledgerseq integer NOT NULL,
    qset text NOT NULL,
    CONSTRAINT scpquorums_lastledgerseq_check CHECK ((lastledgerseq >= 0))
);


--
-- Name: signers; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE signers (
    accountid character varying(56) NOT NULL,
    publickey character varying(56) NOT NULL,
    weight integer NOT NULL
);


--
-- Name: storestate; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE storestate (
    statename character(32) NOT NULL,
    state text
);


--
-- Name: trustlines; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE trustlines (
    accountid character varying(56) NOT NULL,
    assettype integer NOT NULL,
    issuer character varying(56) NOT NULL,
    assetcode character varying(12) NOT NULL,
    tlimit bigint NOT NULL,
    balance bigint NOT NULL,
    flags integer NOT NULL,
    lastmodified integer NOT NULL,
    CONSTRAINT trustlines_balance_check CHECK ((balance >= 0)),
    CONSTRAINT trustlines_tlimit_check CHECK ((tlimit > 0))
);


--
-- Name: txfeehistory; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE txfeehistory (
    txid character(64) NOT NULL,
    ledgerseq integer NOT NULL,
    txindex integer NOT NULL,
    txchanges text NOT NULL,
    CONSTRAINT txfeehistory_ledgerseq_check CHECK ((ledgerseq >= 0))
);


--
-- Name: txhistory; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE txhistory (
    txid character(64) NOT NULL,
    ledgerseq integer NOT NULL,
    txindex integer NOT NULL,
    txbody text NOT NULL,
    txresult text NOT NULL,
    txmeta text NOT NULL,
    CONSTRAINT txhistory_ledgerseq_check CHECK ((ledgerseq >= 0))
);


--
-- Data for Name: accounts; Type: TABLE DATA; Schema: public; Owner: -
--

COPY accounts (accountid, balance, seqnum, numsubentries, inflationdest, homedomain, thresholds, flags, lastmodified) FROM stdin;
GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H	999999969999999700	3	0	\N		AQAAAA==	0	2
GAEDTJ4PPEFVW5XV2S7LUXBEHNQMX5Q2GM562RJGOQG7GVCE5H3HIB4V	9999999900	8589934593	1	\N		AQAAAA==	0	3
GARSFJNXJIHO6ULUBK3DBYKVSIZE7SC72S5DYBCHU7DKL22UXKVD7MXP	9999999900	8589934593	1	\N		AQAAAA==	0	3
GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN	9999999800	8589934594	1	\N		AQAAAA==	0	4
\.


--
-- Data for Name: ledgerheaders; Type: TABLE DATA; Schema: public; Owner: -
--

COPY ledgerheaders (ledgerhash, prevhash, bucketlisthash, ledgerseq, closetime, data) FROM stdin;
63d98f536ee68d1b27b5b89f23af5311b7569a24faf1403ad0b52b633b07be99	0000000000000000000000000000000000000000000000000000000000000000	572a2e32ff248a07b0e70fd1f6d318c1facd20b6cc08c33d5775259868125a16	1	0	AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABXKi4y/ySKB7DnD9H20xjB+s0gtswIwz1XdSWYaBJaFgAAAAEN4Lazp2QAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAZAX14QAAAABkAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
eb6c4b32da2ede2c9596ebfd6c7c66ec9d62e67cd61567964a6614b131185129	63d98f536ee68d1b27b5b89f23af5311b7569a24faf1403ad0b52b633b07be99	9c66dee9eee8a04fdf1e48def94192998ae143c4bc57d49e348f54222606098a	2	1453986751	AAAAAWPZj1Nu5o0bJ7W4nyOvUxG3Vpok+vFAOtC1K2M7B76ZL6euDJxI5oXP7YrOhEKxY+jfD+d9oPDDwAMA/kIl6TgAAAAAVqoTvwAAAAIAAAAIAAAAAQAAAAEAAAAIAAAAAwAAADIAAAAA9g6B9DENzDtqHD436koiLxRKDYIzhsjyqT0b5tTQIlucZt7p7uigT98eSN75QZKZiuFDxLxX1J40j1QiJgYJigAAAAIN4Lazp2QAAAAAAAAAAAEsAAAAAAAAAAAAAAAAAAAAZAX14QAAAAAyAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
e5437432f99ef3bab13b626676c6e71d7de08ab3950fc3f9ae56dc9a94aedc82	eb6c4b32da2ede2c9596ebfd6c7c66ec9d62e67cd61567964a6614b131185129	a9a0c8769f35057cd15f06b1cd9ec500bdddb1b08e1b285b4223b45bb18e70fd	3	1453986752	AAAAAetsSzLaLt4slZbr/Wx8ZuydYuZ81hVnlkpmFLExGFEpxYeML1mE0FFOO6x4OXLziurfFRTRzQfnxDkO3hlp5ScAAAAAVqoTwAAAAAAAAAAAKwBWfKKqAAVrA8ptV4sxg1LDFvO378bXmV7+NQIXA42poMh2nzUFfNFfBrHNnsUAvd2xsI4bKFtCI7RbsY5w/QAAAAMN4Lazp2QAAAAAAAAAAAH0AAAAAAAAAAAAAAAAAAAAZAX14QAAAAAyAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
47983630e6df8e9ca7f30ecdef498bac0c6bf2e8aaf2e5f77782e941efc50030	e5437432f99ef3bab13b626676c6e71d7de08ab3950fc3f9ae56dc9a94aedc82	26d21904b6a9d6ca0f7aaba8176e083e2eb0e44bf4cad5884a6822c6680c7f96	4	1453986753	AAAAAeVDdDL5nvO6sTtiZnbG5x194IqzlQ/D+a5W3JqUrtyCgweDN3suE/1iz7oOHgm55NMWPVBPThvZZOC0kgViCF4AAAAAVqoTwQAAAAAAAAAAPXm8IcCHW9YauF7/m6T+TQSgvRq7Cq53p0J9GD4vnXAm0hkEtqnWyg96q6gXbgg+LrDkS/TK1YhKaCLGaAx/lgAAAAQN4Lazp2QAAAAAAAAAAAK8AAAAAAAAAAAAAAABAAAAZAX14QAAAAAyAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
\.


--
-- Data for Name: offers; Type: TABLE DATA; Schema: public; Owner: -
--

COPY offers (sellerid, offerid, sellingassettype, sellingassetcode, sellingissuer, buyingassettype, buyingassetcode, buyingissuer, amount, pricen, priced, price, flags, lastmodified) FROM stdin;
GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN	1	1	EUR	GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN	1	USD	GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN	5000000000000000	1	200	0.0050000000000000001	0	4
\.


--
-- Data for Name: peers; Type: TABLE DATA; Schema: public; Owner: -
--

COPY peers (ip, port, nextattempt, numfailures) FROM stdin;
\.


--
-- Data for Name: publishqueue; Type: TABLE DATA; Schema: public; Owner: -
--

COPY publishqueue (ledger, state) FROM stdin;
\.


--
-- Data for Name: pubsub; Type: TABLE DATA; Schema: public; Owner: -
--

COPY pubsub (resid, lastread) FROM stdin;
\.


--
-- Data for Name: scphistory; Type: TABLE DATA; Schema: public; Owner: -
--

COPY scphistory (nodeid, ledgerseq, envelope) FROM stdin;
GDKQNDBAOXL5APAWQQYZZW4VAU6LONLCHULCHAG7KX24GGYQRQWK6MCW	2	AAAAANUGjCB119A8FoQxnNuVBTy3NWI9FiOA31X1wxsQjCyvAAAAAAAAAAIAAAACAAAAAQAAAEgvp64MnEjmhc/tis6EQrFj6N8P532g8MPAAwD+QiXpOAAAAABWqhO/AAAAAgAAAAgAAAABAAAAAQAAAAgAAAADAAAAMgAAAAAAAAABu0eV85Ba8OQ0CxFZASBY9MubD6wAgEk/CDE5W/ZQMV4AAABAvcnJ/rSleuTQmKKi48xHl3mcXBL8zRHVx2juQ09cX+5zGeZjeeix5X6EIpJhkNRjrxlNaBJGHAaa8RdIix0/BA==
GDKQNDBAOXL5APAWQQYZZW4VAU6LONLCHULCHAG7KX24GGYQRQWK6MCW	3	AAAAANUGjCB119A8FoQxnNuVBTy3NWI9FiOA31X1wxsQjCyvAAAAAAAAAAMAAAACAAAAAQAAADDFh4wvWYTQUU47rHg5cvOK6t8VFNHNB+fEOQ7eGWnlJwAAAABWqhPAAAAAAAAAAAAAAAABu0eV85Ba8OQ0CxFZASBY9MubD6wAgEk/CDE5W/ZQMV4AAABA+jycpb0qGCtUJwErlVT207WglNJ8i+O0zYoSLbnTu+0IyXz8YcMb6XDtmpNdsowVpeZDqgiOb3oPkAtta7TDBg==
GDKQNDBAOXL5APAWQQYZZW4VAU6LONLCHULCHAG7KX24GGYQRQWK6MCW	4	AAAAANUGjCB119A8FoQxnNuVBTy3NWI9FiOA31X1wxsQjCyvAAAAAAAAAAQAAAACAAAAAQAAADCDB4M3ey4T/WLPug4eCbnk0xY9UE9OG9lk4LSSBWIIXgAAAABWqhPBAAAAAAAAAAAAAAABu0eV85Ba8OQ0CxFZASBY9MubD6wAgEk/CDE5W/ZQMV4AAABAz4rN48bzisCYkedS1tCegHKavrd/SumJ1TUwtIT6CFG0Yy4YvLK68VZLwqVotF71ltCF9r55bEa9dmrivwLnBg==
\.


--
-- Data for Name: scpquorums; Type: TABLE DATA; Schema: public; Owner: -
--

COPY scpquorums (qsethash, lastledgerseq, qset) FROM stdin;
bb4795f3905af0e4340b1159012058f4cb9b0fac0080493f0831395bf650315e	4	AAAAAQAAAAEAAAAA1QaMIHXX0DwWhDGc25UFPLc1Yj0WI4DfVfXDGxCMLK8AAAAA
\.


--
-- Data for Name: signers; Type: TABLE DATA; Schema: public; Owner: -
--

COPY signers (accountid, publickey, weight) FROM stdin;
\.


--
-- Data for Name: storestate; Type: TABLE DATA; Schema: public; Owner: -
--

COPY storestate (statename, state) FROM stdin;
databaseschema                  	2
forcescponnextlaunch            	false
historyarchivestate             	{\n    "version": 1,\n    "server": "v0.3.3-7-geed8964",\n    "currentLedger": 4,\n    "currentBuckets": [\n        {\n            "curr": "859b29707263bb214f2b6918a599a5bd8fe4cd8bfc64f04b2fd00e9485954edc",\n            "next": {\n                "state": 0\n            },\n            "snap": "25e05f9a3a34d835b3632a1fc37510c2930afb832874a2891fb1097ff7761b00"\n        },\n        {\n            "curr": "ef31a20a398ee73ce22275ea8177786bac54656f33dcc4f3fec60d55ddf163d9",\n            "next": {\n                "state": 1,\n                "output": "25e05f9a3a34d835b3632a1fc37510c2930afb832874a2891fb1097ff7761b00"\n            },\n            "snap": "0000000000000000000000000000000000000000000000000000000000000000"\n        },\n        {\n            "curr": "0000000000000000000000000000000000000000000000000000000000000000",\n            "next": {\n                "state": 0\n            },\n            "snap": "0000000000000000000000000000000000000000000000000000000000000000"\n        },\n        {\n            "curr": "0000000000000000000000000000000000000000000000000000000000000000",\n            "next": {\n                "state": 0\n            },\n            "snap": "0000000000000000000000000000000000000000000000000000000000000000"\n        },\n        {\n            "curr": "0000000000000000000000000000000000000000000000000000000000000000",\n            "next": {\n                "state": 0\n            },\n            "snap": "0000000000000000000000000000000000000000000000000000000000000000"\n        },\n        {\n            "curr": "0000000000000000000000000000000000000000000000000000000000000000",\n            "next": {\n                "state": 0\n            },\n            "snap": "0000000000000000000000000000000000000000000000000000000000000000"\n        },\n        {\n            "curr": "0000000000000000000000000000000000000000000000000000000000000000",\n            "next": {\n                "state": 0\n            },\n            "snap": "0000000000000000000000000000000000000000000000000000000000000000"\n        },\n        {\n            "curr": "0000000000000000000000000000000000000000000000000000000000000000",\n            "next": {\n                "state": 0\n            },\n            "snap": "0000000000000000000000000000000000000000000000000000000000000000"\n        },\n        {\n            "curr": "0000000000000000000000000000000000000000000000000000000000000000",\n            "next": {\n                "state": 0\n            },\n            "snap": "0000000000000000000000000000000000000000000000000000000000000000"\n        },\n        {\n            "curr": "0000000000000000000000000000000000000000000000000000000000000000",\n            "next": {\n                "state": 0\n            },\n            "snap": "0000000000000000000000000000000000000000000000000000000000000000"\n        },\n        {\n            "curr": "0000000000000000000000000000000000000000000000000000000000000000",\n            "next": {\n                "state": 0\n            },\n            "snap": "0000000000000000000000000000000000000000000000000000000000000000"\n        }\n    ]\n}
lastscpdata                     	AAAAAgAAAADVBowgddfQPBaEMZzblQU8tzViPRYjgN9V9cMbEIwsrwAAAAAAAAAEAAAAA7tHlfOQWvDkNAsRWQEgWPTLmw+sAIBJPwgxOVv2UDFeAAAAAQAAADCDB4M3ey4T/WLPug4eCbnk0xY9UE9OG9lk4LSSBWIIXgAAAABWqhPBAAAAAAAAAAAAAAABAAAAMIMHgzd7LhP9Ys+6Dh4JueTTFj1QT04b2WTgtJIFYgheAAAAAFaqE8EAAAAAAAAAAAAAAEB3N0Vmd9yXrEKucXM2/fcSHhJiome8Gz2vaU6LC1XAsN88LYCdUigaXaaIupOiLmb70OHY7wuAlkiTt1i2NGkCAAAAANUGjCB119A8FoQxnNuVBTy3NWI9FiOA31X1wxsQjCyvAAAAAAAAAAQAAAACAAAAAQAAADCDB4M3ey4T/WLPug4eCbnk0xY9UE9OG9lk4LSSBWIIXgAAAABWqhPBAAAAAAAAAAAAAAABu0eV85Ba8OQ0CxFZASBY9MubD6wAgEk/CDE5W/ZQMV4AAABAz4rN48bzisCYkedS1tCegHKavrd/SumJ1TUwtIT6CFG0Yy4YvLK68VZLwqVotF71ltCF9r55bEa9dmrivwLnBgAAAAHlQ3Qy+Z7zurE7YmZ2xucdfeCKs5UPw/muVtyalK7cggAAAAIAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAABkAAAAAgAAAAIAAAAAAAAAAAAAAAEAAAAAAAAAAwAAAAFFVVIAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAVVTRAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAEcN5N+CAAAAAAAEAAADIAAAAAAAAAAAAAAAAAAAAAfsBWnsAAABAYasXryVdVS5JURHjPaXWlWxLZhGnaE4LO5W8ODJ3iK8g17i0DH/6VLo9jCin6YnTKW7K6UTaCxQ1p4pVPuFMDAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAGQAAAACAAAAAQAAAAAAAAAAAAAAAQAAAAAAAAABAAAAACMipbdKDu9RdAq2MOFVkjJPyF/UujwER6fGpetUuqo/AAAAAVVTRAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAAO5rKAAAAAAAAAAAB+wFaewAAAEArvbcTeceld8fpsIeGwDo6ToD1HADd9Xci56O3TvuceyCvxIjTR6nmcl+S1MxYk684gKnOb217poABFAE/8koPAAAAAQAAAAEAAAABAAAAANUGjCB119A8FoQxnNuVBTy3NWI9FiOA31X1wxsQjCyvAAAAAA==
lastclosedledger                	47983630e6df8e9ca7f30ecdef498bac0c6bf2e8aaf2e5f77782e941efc50030
\.


--
-- Data for Name: trustlines; Type: TABLE DATA; Schema: public; Owner: -
--

COPY trustlines (accountid, assettype, issuer, assetcode, tlimit, balance, flags, lastmodified) FROM stdin;
GAEDTJ4PPEFVW5XV2S7LUXBEHNQMX5Q2GM562RJGOQG7GVCE5H3HIB4V	1	GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN	EUR	9223372036854775807	0	1	3
GARSFJNXJIHO6ULUBK3DBYKVSIZE7SC72S5DYBCHU7DKL22UXKVD7MXP	1	GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN	USD	9223372036854775807	1000000000	1	4
\.


--
-- Data for Name: txfeehistory; Type: TABLE DATA; Schema: public; Owner: -
--

COPY txfeehistory (txid, ledgerseq, txindex, txchanges) FROM stdin;
9ff6b71bba6b24c8afc6ca53b12702566d40e9c8abf9b8439e9a3b08300e4188	2	1	AAAAAgAAAAMAAAABAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtrOnZAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtrOnY/+cAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==
afc983f7f88809442fc616c1ce425b6c8cf6d8f7493b33ff6a809003a845ec16	2	2	AAAAAQAAAAEAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtrOnY/84AAAAAAAAAAIAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==
5c71f1b198a9be6b1c7e36737bb3e32c7f258530573bc11f20827d0913ba1433	2	3	AAAAAQAAAAEAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtrOnY/7UAAAAAAAAAAMAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==
cdb0ce2cef61c5cab6566c1c65c5bc632f943b2bd43cfdd80984a66994cb7484	3	1	AAAAAgAAAAMAAAACAAAAAAAAAAAIOaePeQtbdvXUvrpcJDtgy/YaMzvtRSZ0DfNUROn2dAAAAAJUC+QAAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAAIOaePeQtbdvXUvrpcJDtgy/YaMzvtRSZ0DfNUROn2dAAAAAJUC+OcAAAAAgAAAAEAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==
563ab1cf87e65377e015813499a6027081e63177a9ef35a0c5e0c0c2a5a918f7	3	2	AAAAAgAAAAMAAAACAAAAAAAAAAAjIqW3Sg7vUXQKtjDhVZIyT8hf1Lo8BEenxqXrVLqqPwAAAAJUC+QAAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAAjIqW3Sg7vUXQKtjDhVZIyT8hf1Lo8BEenxqXrVLqqPwAAAAJUC+OcAAAAAgAAAAEAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==
378d6013b2a9593a06a39af1a7b08a0c3e0343aa4d40a25a4be7783c54772db0	4	1	AAAAAgAAAAMAAAACAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAJUC+QAAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAAEAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAJUC+OcAAAAAgAAAAEAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==
2cbac484f2328faeb48b6737c450772d59d7e1e629cd4fa184c33c9a382e8088	4	2	AAAAAQAAAAEAAAAEAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAJUC+M4AAAAAgAAAAIAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==
\.


--
-- Data for Name: txhistory; Type: TABLE DATA; Schema: public; Owner: -
--

COPY txhistory (txid, ledgerseq, txindex, txbody, txresult, txmeta) FROM stdin;
9ff6b71bba6b24c8afc6ca53b12702566d40e9c8abf9b8439e9a3b08300e4188	2	1	AAAAAGL8HQvQkbK2HA3WVjRrKmjX00fG8sLI7m0ERwJW/AX3AAAAZAAAAAAAAAABAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAACVAvkAAAAAAAAAAABVvwF9wAAAEC96/+BcbMflvMQfFAQTbAKGu+6BR1M6SG/KVzTJSlIY8ovSVywuthk9dOW9jm23siTiIZE0IAl84wK83gnAcEK	n/a3G7prJMivxspTsScCVm1A6cir+bhDnpo7CDAOQYgAAAAAAAAAZAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAA==	AAAAAAAAAAEAAAACAAAAAAAAAAIAAAAAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAlQL5AAAAAACAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAIAAAAAAAAAAGL8HQvQkbK2HA3WVjRrKmjX00fG8sLI7m0ERwJW/AX3DeC2sVNYGtQAAAAAAAAAAwAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAA
afc983f7f88809442fc616c1ce425b6c8cf6d8f7493b33ff6a809003a845ec16	2	2	AAAAAGL8HQvQkbK2HA3WVjRrKmjX00fG8sLI7m0ERwJW/AX3AAAAZAAAAAAAAAACAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAIyKlt0oO71F0CrYw4VWSMk/IX9S6PARHp8al61S6qj8AAAACVAvkAAAAAAAAAAABVvwF9wAAAECXh9/+55FmQjeMx9IMQfBn42fYY1fKAT/gb+e7P3jdSMCKiKhwoxj1bub733a2XuxPETpe79uzatzm8/KI0asI	r8mD9/iICUQvxhbBzkJbbIz22PdJOzP/aoCQA6hF7BYAAAAAAAAAZAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAA==	AAAAAAAAAAEAAAACAAAAAAAAAAIAAAAAAAAAACMipbdKDu9RdAq2MOFVkjJPyF/UujwER6fGpetUuqo/AAAAAlQL5AAAAAACAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAIAAAAAAAAAAGL8HQvQkbK2HA3WVjRrKmjX00fG8sLI7m0ERwJW/AX3DeC2rv9MNtQAAAAAAAAAAwAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAA
5c71f1b198a9be6b1c7e36737bb3e32c7f258530573bc11f20827d0913ba1433	2	3	AAAAAGL8HQvQkbK2HA3WVjRrKmjX00fG8sLI7m0ERwJW/AX3AAAAZAAAAAAAAAADAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAACDmnj3kLW3b11L66XCQ7YMv2GjM77UUmdA3zVETp9nQAAAACVAvkAAAAAAAAAAABVvwF9wAAAEBdnUBKCV3HlHOqp5iNLsP8auaNCvxDeBp+0C+lwPYUNrzRALQRDJDWuKfExmsRrEnW8LKJbMdeW9ilLaEc2lAO	XHHxsZipvmscfjZze7PjLH8lhTBXO8EfIIJ9CRO6FDMAAAAAAAAAZAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAA==	AAAAAAAAAAEAAAACAAAAAAAAAAIAAAAAAAAAAAg5p495C1t29dS+ulwkO2DL9hozO+1FJnQN81RE6fZ0AAAAAlQL5AAAAAACAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAIAAAAAAAAAAGL8HQvQkbK2HA3WVjRrKmjX00fG8sLI7m0ERwJW/AX3DeC2rKtAUtQAAAAAAAAAAwAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAA
cdb0ce2cef61c5cab6566c1c65c5bc632f943b2bd43cfdd80984a66994cb7484	3	1	AAAAAAg5p495C1t29dS+ulwkO2DL9hozO+1FJnQN81RE6fZ0AAAAZAAAAAIAAAABAAAAAAAAAAAAAAABAAAAAAAAAAYAAAABRVVSAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFae3//////////AAAAAAAAAAFE6fZ0AAAAQNSm/BIxP9LwabXoBKigrTG85o/PUp6VOWh/ne6mMaT5hvehDUvbRHQghJ/SZTDfjD+FPPjbe3nzcJEun1EX4g0=	zbDOLO9hxcq2VmwcZcW8Yy+UOyvUPP3YCYSmaZTLdIQAAAAAAAAAZAAAAAAAAAABAAAAAAAAAAYAAAAAAAAAAA==	AAAAAAAAAAEAAAACAAAAAAAAAAMAAAABAAAAAAg5p495C1t29dS+ulwkO2DL9hozO+1FJnQN81RE6fZ0AAAAAUVVUgAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAAAAAAAH//////////AAAAAQAAAAAAAAAAAAAAAQAAAAMAAAAAAAAAAAg5p495C1t29dS+ulwkO2DL9hozO+1FJnQN81RE6fZ0AAAAAlQL45wAAAACAAAAAQAAAAEAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAA
563ab1cf87e65377e015813499a6027081e63177a9ef35a0c5e0c0c2a5a918f7	3	2	AAAAACMipbdKDu9RdAq2MOFVkjJPyF/UujwER6fGpetUuqo/AAAAZAAAAAIAAAABAAAAAAAAAAAAAAABAAAAAAAAAAYAAAABVVNEAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFae3//////////AAAAAAAAAAFUuqo/AAAAQKblh64EFK4tp2+xohOkNaSdfMFDId/y4nop7dVmZRsbe4f/eVCpUrKQCRWLJ4bXzMpPTOTsjHRIxOa8WekP+ww=	Vjqxz4fmU3fgFYE0maYCcIHmMXep7zWgxeDAwqWpGPcAAAAAAAAAZAAAAAAAAAABAAAAAAAAAAYAAAAAAAAAAA==	AAAAAAAAAAEAAAACAAAAAAAAAAMAAAABAAAAACMipbdKDu9RdAq2MOFVkjJPyF/UujwER6fGpetUuqo/AAAAAVVTRAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAAAAAAAH//////////AAAAAQAAAAAAAAAAAAAAAQAAAAMAAAAAAAAAACMipbdKDu9RdAq2MOFVkjJPyF/UujwER6fGpetUuqo/AAAAAlQL45wAAAACAAAAAQAAAAEAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAA
378d6013b2a9593a06a39af1a7b08a0c3e0343aa4d40a25a4be7783c54772db0	4	1	AAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAZAAAAAIAAAABAAAAAAAAAAAAAAABAAAAAAAAAAEAAAAAIyKlt0oO71F0CrYw4VWSMk/IX9S6PARHp8al61S6qj8AAAABVVNEAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAA7msoAAAAAAAAAAAH7AVp7AAAAQCu9txN5x6V3x+mwh4bAOjpOgPUcAN31dyLno7dO+5x7IK/EiNNHqeZyX5LUzFiTrziAqc5vbXumgAEUAT/ySg8=	N41gE7KpWToGo5rxp7CKDD4DQ6pNQKJaS+d4PFR3LbAAAAAAAAAAZAAAAAAAAAABAAAAAAAAAAEAAAAAAAAAAA==	AAAAAAAAAAEAAAACAAAAAwAAAAMAAAABAAAAACMipbdKDu9RdAq2MOFVkjJPyF/UujwER6fGpetUuqo/AAAAAVVTRAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAAAAAAAH//////////AAAAAQAAAAAAAAAAAAAAAQAAAAQAAAABAAAAACMipbdKDu9RdAq2MOFVkjJPyF/UujwER6fGpetUuqo/AAAAAVVTRAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAAO5rKAH//////////AAAAAQAAAAAAAAAA
2cbac484f2328faeb48b6737c450772d59d7e1e629cd4fa184c33c9a382e8088	4	2	AAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAZAAAAAIAAAACAAAAAAAAAAAAAAABAAAAAAAAAAMAAAABRVVSAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAFVU0QAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7ABHDeTfggAAAAAABAAAAyAAAAAAAAAAAAAAAAAAAAAH7AVp7AAAAQGGrF68lXVUuSVER4z2l1pVsS2YRp2hOCzuVvDgyd4ivINe4tAx/+lS6PYwop+mJ0yluyulE2gsUNaeKVT7hTAw=	LLrEhPIyj660i2c3xFB3LVnX4eYpzU+hhMM8mjgugIgAAAAAAAAAZAAAAAAAAAABAAAAAAAAAAMAAAAAAAAAAAAAAAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAAAAAAAAAAQAAAAFFVVIAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAVVTRAAAAAAA5BFB2+Hs81DQk/cAlJes5R0+3PUQaZ62NZJoKPsBWnsAEcN5N+CAAAAAAAEAAADIAAAAAAAAAAAAAAAA	AAAAAAAAAAEAAAACAAAAAAAAAAQAAAACAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7AAAAAAAAAAEAAAABRVVSAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAFVU0QAAAAAAOQRQdvh7PNQ0JP3AJSXrOUdPtz1EGmetjWSaCj7AVp7ABHDeTfggAAAAAABAAAAyAAAAAAAAAAAAAAAAAAAAAEAAAAEAAAAAAAAAADkEUHb4ezzUNCT9wCUl6zlHT7c9RBpnrY1kmgo+wFaewAAAAJUC+M4AAAAAgAAAAIAAAABAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==
\.


--
-- Name: accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: -; Tablespace: 
--

ALTER TABLE ONLY accounts
    ADD CONSTRAINT accounts_pkey PRIMARY KEY (accountid);


--
-- Name: ledgerheaders_ledgerseq_key; Type: CONSTRAINT; Schema: public; Owner: -; Tablespace: 
--

ALTER TABLE ONLY ledgerheaders
    ADD CONSTRAINT ledgerheaders_ledgerseq_key UNIQUE (ledgerseq);


--
-- Name: ledgerheaders_pkey; Type: CONSTRAINT; Schema: public; Owner: -; Tablespace: 
--

ALTER TABLE ONLY ledgerheaders
    ADD CONSTRAINT ledgerheaders_pkey PRIMARY KEY (ledgerhash);


--
-- Name: offers_pkey; Type: CONSTRAINT; Schema: public; Owner: -; Tablespace: 
--

ALTER TABLE ONLY offers
    ADD CONSTRAINT offers_pkey PRIMARY KEY (offerid);


--
-- Name: peers_pkey; Type: CONSTRAINT; Schema: public; Owner: -; Tablespace: 
--

ALTER TABLE ONLY peers
    ADD CONSTRAINT peers_pkey PRIMARY KEY (ip, port);


--
-- Name: publishqueue_pkey; Type: CONSTRAINT; Schema: public; Owner: -; Tablespace: 
--

ALTER TABLE ONLY publishqueue
    ADD CONSTRAINT publishqueue_pkey PRIMARY KEY (ledger);


--
-- Name: pubsub_pkey; Type: CONSTRAINT; Schema: public; Owner: -; Tablespace: 
--

ALTER TABLE ONLY pubsub
    ADD CONSTRAINT pubsub_pkey PRIMARY KEY (resid);


--
-- Name: scpquorums_pkey; Type: CONSTRAINT; Schema: public; Owner: -; Tablespace: 
--

ALTER TABLE ONLY scpquorums
    ADD CONSTRAINT scpquorums_pkey PRIMARY KEY (qsethash);


--
-- Name: signers_pkey; Type: CONSTRAINT; Schema: public; Owner: -; Tablespace: 
--

ALTER TABLE ONLY signers
    ADD CONSTRAINT signers_pkey PRIMARY KEY (accountid, publickey);


--
-- Name: storestate_pkey; Type: CONSTRAINT; Schema: public; Owner: -; Tablespace: 
--

ALTER TABLE ONLY storestate
    ADD CONSTRAINT storestate_pkey PRIMARY KEY (statename);


--
-- Name: trustlines_pkey; Type: CONSTRAINT; Schema: public; Owner: -; Tablespace: 
--

ALTER TABLE ONLY trustlines
    ADD CONSTRAINT trustlines_pkey PRIMARY KEY (accountid, issuer, assetcode);


--
-- Name: txfeehistory_pkey; Type: CONSTRAINT; Schema: public; Owner: -; Tablespace: 
--

ALTER TABLE ONLY txfeehistory
    ADD CONSTRAINT txfeehistory_pkey PRIMARY KEY (ledgerseq, txindex);


--
-- Name: txhistory_pkey; Type: CONSTRAINT; Schema: public; Owner: -; Tablespace: 
--

ALTER TABLE ONLY txhistory
    ADD CONSTRAINT txhistory_pkey PRIMARY KEY (ledgerseq, txindex);


--
-- Name: accountbalances; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX accountbalances ON accounts USING btree (balance) WHERE (balance >= 1000000000);


--
-- Name: buyingissuerindex; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX buyingissuerindex ON offers USING btree (buyingissuer);


--
-- Name: histbyseq; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX histbyseq ON txhistory USING btree (ledgerseq);


--
-- Name: histfeebyseq; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX histfeebyseq ON txfeehistory USING btree (ledgerseq);


--
-- Name: ledgersbyseq; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX ledgersbyseq ON ledgerheaders USING btree (ledgerseq);


--
-- Name: priceindex; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX priceindex ON offers USING btree (price);


--
-- Name: scpenvsbyseq; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX scpenvsbyseq ON scphistory USING btree (ledgerseq);


--
-- Name: sellingissuerindex; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX sellingissuerindex ON offers USING btree (sellingissuer);


--
-- Name: signersaccount; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX signersaccount ON signers USING btree (accountid);


--
-- PostgreSQL database dump complete
--

