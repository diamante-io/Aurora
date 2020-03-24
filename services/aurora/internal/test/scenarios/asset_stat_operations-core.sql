running recipe
recipe finished
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

SET search_path = public, pg_catalog;

DROP INDEX IF EXISTS public.upgradehistbyseq;
DROP INDEX IF EXISTS public.scpquorumsbyseq;
DROP INDEX IF EXISTS public.scpenvsbyseq;
DROP INDEX IF EXISTS public.ledgersbyseq;
DROP INDEX IF EXISTS public.histfeebyseq;
DROP INDEX IF EXISTS public.histbyseq;
DROP INDEX IF EXISTS public.bestofferindex;
DROP INDEX IF EXISTS public.accountbalances;
ALTER TABLE IF EXISTS ONLY public.upgradehistory DROP CONSTRAINT IF EXISTS upgradehistory_pkey;
ALTER TABLE IF EXISTS ONLY public.txhistory DROP CONSTRAINT IF EXISTS txhistory_pkey;
ALTER TABLE IF EXISTS ONLY public.txfeehistory DROP CONSTRAINT IF EXISTS txfeehistory_pkey;
ALTER TABLE IF EXISTS ONLY public.trustlines DROP CONSTRAINT IF EXISTS trustlines_pkey;
ALTER TABLE IF EXISTS ONLY public.storestate DROP CONSTRAINT IF EXISTS storestate_pkey;
ALTER TABLE IF EXISTS ONLY public.scpquorums DROP CONSTRAINT IF EXISTS scpquorums_pkey;
ALTER TABLE IF EXISTS ONLY public.quoruminfo DROP CONSTRAINT IF EXISTS quoruminfo_pkey;
ALTER TABLE IF EXISTS ONLY public.pubsub DROP CONSTRAINT IF EXISTS pubsub_pkey;
ALTER TABLE IF EXISTS ONLY public.publishqueue DROP CONSTRAINT IF EXISTS publishqueue_pkey;
ALTER TABLE IF EXISTS ONLY public.peers DROP CONSTRAINT IF EXISTS peers_pkey;
ALTER TABLE IF EXISTS ONLY public.offers DROP CONSTRAINT IF EXISTS offers_pkey;
ALTER TABLE IF EXISTS ONLY public.ledgerheaders DROP CONSTRAINT IF EXISTS ledgerheaders_pkey;
ALTER TABLE IF EXISTS ONLY public.ledgerheaders DROP CONSTRAINT IF EXISTS ledgerheaders_ledgerseq_key;
ALTER TABLE IF EXISTS ONLY public.ban DROP CONSTRAINT IF EXISTS ban_pkey;
ALTER TABLE IF EXISTS ONLY public.accounts DROP CONSTRAINT IF EXISTS accounts_pkey;
ALTER TABLE IF EXISTS ONLY public.accountdata DROP CONSTRAINT IF EXISTS accountdata_pkey;
DROP TABLE IF EXISTS public.upgradehistory;
DROP TABLE IF EXISTS public.txhistory;
DROP TABLE IF EXISTS public.txfeehistory;
DROP TABLE IF EXISTS public.trustlines;
DROP TABLE IF EXISTS public.storestate;
DROP TABLE IF EXISTS public.scpquorums;
DROP TABLE IF EXISTS public.scphistory;
DROP TABLE IF EXISTS public.quoruminfo;
DROP TABLE IF EXISTS public.pubsub;
DROP TABLE IF EXISTS public.publishqueue;
DROP TABLE IF EXISTS public.peers;
DROP TABLE IF EXISTS public.offers;
DROP TABLE IF EXISTS public.ledgerheaders;
DROP TABLE IF EXISTS public.ban;
DROP TABLE IF EXISTS public.accounts;
DROP TABLE IF EXISTS public.accountdata;
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

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: accountdata; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE accountdata (
    accountid character varying(56) NOT NULL,
    dataname character varying(88) NOT NULL,
    datavalue character varying(112) NOT NULL,
    lastmodified integer NOT NULL
);


--
-- Name: accounts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE accounts (
    accountid character varying(56) NOT NULL,
    balance bigint NOT NULL,
    seqnum bigint NOT NULL,
    numsubentries integer NOT NULL,
    inflationdest character varying(56),
    homedomain character varying(44) NOT NULL,
    thresholds text NOT NULL,
    flags integer NOT NULL,
    lastmodified integer NOT NULL,
    buyingliabilities bigint,
    sellingliabilities bigint,
    signers text,
    CONSTRAINT accounts_balance_check CHECK ((balance >= 0)),
    CONSTRAINT accounts_buyingliabilities_check CHECK ((buyingliabilities >= 0)),
    CONSTRAINT accounts_numsubentries_check CHECK ((numsubentries >= 0)),
    CONSTRAINT accounts_sellingliabilities_check CHECK ((sellingliabilities >= 0))
);


--
-- Name: ban; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE ban (
    nodeid character(56) NOT NULL
);


--
-- Name: ledgerheaders; Type: TABLE; Schema: public; Owner: -
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
-- Name: offers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE offers (
    sellerid character varying(56) NOT NULL,
    offerid bigint NOT NULL,
    sellingasset text NOT NULL,
    buyingasset text NOT NULL,
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
-- Name: peers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE peers (
    ip character varying(15) NOT NULL,
    port integer DEFAULT 0 NOT NULL,
    nextattempt timestamp without time zone NOT NULL,
    numfailures integer DEFAULT 0 NOT NULL,
    type integer NOT NULL,
    CONSTRAINT peers_numfailures_check CHECK ((numfailures >= 0)),
    CONSTRAINT peers_port_check CHECK (((port > 0) AND (port <= 65535)))
);


--
-- Name: publishqueue; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE publishqueue (
    ledger integer NOT NULL,
    state text
);


--
-- Name: pubsub; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE pubsub (
    resid character(32) NOT NULL,
    lastread integer
);


--
-- Name: quoruminfo; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE quoruminfo (
    nodeid character(56) NOT NULL,
    qsethash character(64) NOT NULL
);


--
-- Name: scphistory; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE scphistory (
    nodeid character(56) NOT NULL,
    ledgerseq integer NOT NULL,
    envelope text NOT NULL,
    CONSTRAINT scphistory_ledgerseq_check CHECK ((ledgerseq >= 0))
);


--
-- Name: scpquorums; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE scpquorums (
    qsethash character(64) NOT NULL,
    lastledgerseq integer NOT NULL,
    qset text NOT NULL,
    CONSTRAINT scpquorums_lastledgerseq_check CHECK ((lastledgerseq >= 0))
);


--
-- Name: storestate; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE storestate (
    statename character(32) NOT NULL,
    state text
);


--
-- Name: trustlines; Type: TABLE; Schema: public; Owner: -
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
    buyingliabilities bigint,
    sellingliabilities bigint,
    CONSTRAINT trustlines_balance_check CHECK ((balance >= 0)),
    CONSTRAINT trustlines_buyingliabilities_check CHECK ((buyingliabilities >= 0)),
    CONSTRAINT trustlines_sellingliabilities_check CHECK ((sellingliabilities >= 0)),
    CONSTRAINT trustlines_tlimit_check CHECK ((tlimit > 0))
);


--
-- Name: txfeehistory; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE txfeehistory (
    txid character(64) NOT NULL,
    ledgerseq integer NOT NULL,
    txindex integer NOT NULL,
    txchanges text NOT NULL,
    CONSTRAINT txfeehistory_ledgerseq_check CHECK ((ledgerseq >= 0))
);


--
-- Name: txhistory; Type: TABLE; Schema: public; Owner: -
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
-- Name: upgradehistory; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE upgradehistory (
    ledgerseq integer NOT NULL,
    upgradeindex integer NOT NULL,
    upgrade text NOT NULL,
    changes text NOT NULL,
    CONSTRAINT upgradehistory_ledgerseq_check CHECK ((ledgerseq >= 0))
);


--
-- Data for Name: accountdata; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: accounts; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO accounts VALUES ('GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H', 999999959999999600, 4, 0, NULL, '', 'AQAAAA==', 0, 2, NULL, NULL, NULL);
INSERT INTO accounts VALUES ('GCFZWN3AOVFQM2BZTZX7P47WSI4QMGJC62LILPKODTNDLVKZZNA5BQJ3', 10000000000, 8589934592, 0, NULL, '', 'AQAAAA==', 0, 2, NULL, NULL, NULL);
INSERT INTO accounts VALUES ('GAB7GMQPJ5YY2E4UJMLNAZPDEUKPK4AAIPRXIZHKZGUIRC6FP2LAQSDN', 10000000000, 8589934592, 0, NULL, '', 'AQAAAA==', 0, 2, NULL, NULL, NULL);
INSERT INTO accounts VALUES ('GCSX4PDUZP3BL522ZVMFXCEJ55NKEOHEMII7PSMJZNAAESJ444GSSJMO', 9999999900, 8589934593, 1, NULL, '', 'AQAAAA==', 0, 3, NULL, NULL, NULL);
INSERT INTO accounts VALUES ('GCYLTPOU7IVYHHA3XKQF4YB4W4ZWHFERMOQ7K47IWANKNBFBNJJNEOG5', 9999999900, 8589934593, 1, NULL, '', 'AQAAAA==', 0, 3, NULL, NULL, NULL);


--
-- Data for Name: ban; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: ledgerheaders; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO ledgerheaders VALUES ('63d98f536ee68d1b27b5b89f23af5311b7569a24faf1403ad0b52b633b07be99', '0000000000000000000000000000000000000000000000000000000000000000', '572a2e32ff248a07b0e70fd1f6d318c1facd20b6cc08c33d5775259868125a16', 1, 0, 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABXKi4y/ySKB7DnD9H20xjB+s0gtswIwz1XdSWYaBJaFgAAAAEN4Lazp2QAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAZAX14QAAAABkAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA');
INSERT INTO ledgerheaders VALUES ('0dc75706bf9958fdd2326dca73c429ba839b3565bf43959c972460e2e741d52b', '63d98f536ee68d1b27b5b89f23af5311b7569a24faf1403ad0b52b633b07be99', '4a4bacf70ee86458779f2ff4e54fd8710a5434bf0d58307e88dfc105d9038419', 2, 1559579667, 'AAAAC2PZj1Nu5o0bJ7W4nyOvUxG3Vpok+vFAOtC1K2M7B76ZAl4xVs7p/dHViPjkvi2jwT71A96KVxBFTiVfkaPKPPMAAAAAXPVMEwAAAAIAAAAIAAAAAQAAAAsAAAAIAAAAAwAPQkAAAAAAi4Ue20vNo0cmNykBLoF1m6yPEvTdv3MfCquWXwc7NpxKS6z3DuhkWHefL/TlT9hxClQ0vw1YMH6I38EF2QOEGQAAAAIN4Lazp2QAAAAAAAAAAAGQAAAAAAAAAAAAAAAAAAAAZAX14QAAD0JAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA');
INSERT INTO ledgerheaders VALUES ('18cbb37ab371dcb900c846a1fd8000c2a09c68ea23ba43858c0deed05df2d9cd', '0dc75706bf9958fdd2326dca73c429ba839b3565bf43959c972460e2e741d52b', '255d96414b66ee031f96ac428fd27dfeea5137eab44cd9e878180d9bac6c1a57', 3, 1559579668, 'AAAACw3HVwa/mVj90jJtynPEKbqDmzVlv0OVnJckYOLnQdUrEfAYHWfAn84njhZoaIFZlI1KsNuRDfce9f6vVb64XJ0AAAAAXPVMFAAAAAAAAAAAhSD66Lhzb7ZNorUyzTQqB11nd0srMHPv8kaiQLT5higlXZZBS2buAx+WrEKP0n3+6lE36rRM2eh4GA2brGwaVwAAAAMN4Lazp2QAAAAAAAAAAAJYAAAAAAAAAAAAAAAAAAAAZAX14QAAD0JAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA');


--
-- Data for Name: offers; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: peers; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: publishqueue; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: pubsub; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: quoruminfo; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: scphistory; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO scphistory VALUES ('GCHSPAWNLVYEOJ5QIWN5DRT2KFGOFRRXPBMGGD2C2X2ZCFXBXIN32XTI', 2, 'AAAAAI8ngs1dcEcnsEWb0cZ6UUzixjd4WGMPQtX1kRbhuhu9AAAAAAAAAAIAAAACAAAAAQAAAEgCXjFWzun90dWI+OS+LaPBPvUD3opXEEVOJV+Ro8o88wAAAABc9UwTAAAAAgAAAAgAAAABAAAACwAAAAgAAAADAA9CQAAAAAAAAAABZgcwdPbe7XdMFPZ8j87J0zEqCP++XEJkH0Ddvf9ZmLIAAABA8ubzb3l+hx93XhuSnGYNuHPuKwYEo46Lz/ZrK1ghbHD99B5+JzOokzf8BD+q8XBrBz8A0k4xF7H79EoOZ/QlDA==');
INSERT INTO scphistory VALUES ('GCHSPAWNLVYEOJ5QIWN5DRT2KFGOFRRXPBMGGD2C2X2ZCFXBXIN32XTI', 3, 'AAAAAI8ngs1dcEcnsEWb0cZ6UUzixjd4WGMPQtX1kRbhuhu9AAAAAAAAAAMAAAACAAAAAQAAADAR8BgdZ8CfzieOFmhogVmUjUqw25EN9x71/q9VvrhcnQAAAABc9UwUAAAAAAAAAAAAAAABZgcwdPbe7XdMFPZ8j87J0zEqCP++XEJkH0Ddvf9ZmLIAAABAaED18lLRkFoinjFqAE7h52NkZQKxOKA84a7hHsj9V3xBjc6g+MGz67Db7Cxo7wFPbRTwAnBJN+n9DY9t93eBBg==');


--
-- Data for Name: scpquorums; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO scpquorums VALUES ('66073074f6deed774c14f67c8fcec9d3312a08ffbe5c42641f40ddbdff5998b2', 3, 'AAAAAQAAAAEAAAAAjyeCzV1wRyewRZvRxnpRTOLGN3hYYw9C1fWRFuG6G70AAAAA');


--
-- Data for Name: storestate; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO storestate VALUES ('databaseschema                  ', '10');
INSERT INTO storestate VALUES ('networkpassphrase               ', 'Test SDF Network ; September 2015');
INSERT INTO storestate VALUES ('forcescponnextlaunch            ', 'false');
INSERT INTO storestate VALUES ('ledgerupgrades                  ', '{
    "time": 0,
    "version": {
        "has": false
    },
    "fee": {
        "has": false
    },
    "maxtxsize": {
        "has": false
    },
    "reserve": {
        "has": false
    }
}');
INSERT INTO storestate VALUES ('lastclosedledger                ', '18cbb37ab371dcb900c846a1fd8000c2a09c68ea23ba43858c0deed05df2d9cd');
INSERT INTO storestate VALUES ('historyarchivestate             ', '{
    "version": 1,
    "server": "v11.1.0",
    "currentLedger": 3,
    "currentBuckets": [
        {
            "curr": "f21f6db99c8ac25b8bf672e8f75360230a9864a52509699e1ba03a21d7d884bd",
            "next": {
                "state": 0
            },
            "snap": "ef31a20a398ee73ce22275ea8177786bac54656f33dcc4f3fec60d55ddf163d9"
        },
        {
            "curr": "0000000000000000000000000000000000000000000000000000000000000000",
            "next": {
                "state": 1,
                "output": "ef31a20a398ee73ce22275ea8177786bac54656f33dcc4f3fec60d55ddf163d9"
            },
            "snap": "0000000000000000000000000000000000000000000000000000000000000000"
        },
        {
            "curr": "0000000000000000000000000000000000000000000000000000000000000000",
            "next": {
                "state": 0
            },
            "snap": "0000000000000000000000000000000000000000000000000000000000000000"
        },
        {
            "curr": "0000000000000000000000000000000000000000000000000000000000000000",
            "next": {
                "state": 0
            },
            "snap": "0000000000000000000000000000000000000000000000000000000000000000"
        },
        {
            "curr": "0000000000000000000000000000000000000000000000000000000000000000",
            "next": {
                "state": 0
            },
            "snap": "0000000000000000000000000000000000000000000000000000000000000000"
        },
        {
            "curr": "0000000000000000000000000000000000000000000000000000000000000000",
            "next": {
                "state": 0
            },
            "snap": "0000000000000000000000000000000000000000000000000000000000000000"
        },
        {
            "curr": "0000000000000000000000000000000000000000000000000000000000000000",
            "next": {
                "state": 0
            },
            "snap": "0000000000000000000000000000000000000000000000000000000000000000"
        },
        {
            "curr": "0000000000000000000000000000000000000000000000000000000000000000",
            "next": {
                "state": 0
            },
            "snap": "0000000000000000000000000000000000000000000000000000000000000000"
        },
        {
            "curr": "0000000000000000000000000000000000000000000000000000000000000000",
            "next": {
                "state": 0
            },
            "snap": "0000000000000000000000000000000000000000000000000000000000000000"
        },
        {
            "curr": "0000000000000000000000000000000000000000000000000000000000000000",
            "next": {
                "state": 0
            },
            "snap": "0000000000000000000000000000000000000000000000000000000000000000"
        },
        {
            "curr": "0000000000000000000000000000000000000000000000000000000000000000",
            "next": {
                "state": 0
            },
            "snap": "0000000000000000000000000000000000000000000000000000000000000000"
        }
    ]
}');
INSERT INTO storestate VALUES ('lastscpdata                     ', 'AAAAAgAAAACPJ4LNXXBHJ7BFm9HGelFM4sY3eFhjD0LV9ZEW4bobvQAAAAAAAAADAAAAA2YHMHT23u13TBT2fI/OydMxKgj/vlxCZB9A3b3/WZiyAAAAAQAAAJgR8BgdZ8CfzieOFmhogVmUjUqw25EN9x71/q9VvrhcnQAAAABc9UwUAAAAAAAAAAEAAAAAjyeCzV1wRyewRZvRxnpRTOLGN3hYYw9C1fWRFuG6G70AAABAs4etSBgrBvS2X4JMge/H2K0HB3HoU/e8RNnqPmHNjyQBpxRSNsNUk8alITPvmo3nA4fyvJC4QlVUySFY+rvHCgAAAAEAAACYEfAYHWfAn84njhZoaIFZlI1KsNuRDfce9f6vVb64XJ0AAAAAXPVMFAAAAAAAAAABAAAAAI8ngs1dcEcnsEWb0cZ6UUzixjd4WGMPQtX1kRbhuhu9AAAAQLOHrUgYKwb0tl+CTIHvx9itBwdx6FP3vETZ6j5hzY8kAacUUjbDVJPGpSEz75qN5wOH8ryQuEJVVMkhWPq7xwoAAABA+IBefeM9vDmssNIeMVdovciRH8A3ilS42BRnH0zck8ZwZ3LtFhJlY45HIDC/TVMBZA5WNTQUWmoUY/OLCCI5DgAAAACPJ4LNXXBHJ7BFm9HGelFM4sY3eFhjD0LV9ZEW4bobvQAAAAAAAAADAAAAAgAAAAEAAAAwEfAYHWfAn84njhZoaIFZlI1KsNuRDfce9f6vVb64XJ0AAAAAXPVMFAAAAAAAAAAAAAAAAWYHMHT23u13TBT2fI/OydMxKgj/vlxCZB9A3b3/WZiyAAAAQGhA9fJS0ZBaIp4xagBO4edjZGUCsTigPOGu4R7I/Vd8QY3OoPjBs+uw2+wsaO8BT20U8AJwSTfp/Q2Pbfd3gQYAAAABDcdXBr+ZWP3SMm3Kc8QpuoObNWW/Q5WclyRg4udB1SsAAAACAAAAALC5vdT6K4OcG7qgXmA8tzNjlJFjofVz6LAapoShalLSAAAAZAAAAAIAAAABAAAAAAAAAAAAAAABAAAAAAAAAAYAAAABVVNEAAAAAACLmzdgdUsGaDmeb/fz9pI5Bhki9paFvU4c2jXVWctB0H//////////AAAAAAAAAAGhalLSAAAAQCnZlvEITlesuWctD/z9XMvJSujWqCYBR5F6tzeyStpEqiQXeyOEtJrGO3ulubYfi4kuwxETF5+CORR/2+0IKA4AAAAApX48dMv2FfdazVhbiInvWqI45GIR98mJy0ACSTznDSkAAABkAAAAAgAAAAEAAAAAAAAAAAAAAAEAAAAAAAAABgAAAAFVU0QAAAAAAAPzMg9PcY0TlEsW0GXjJRT1cABD43Rk6smoiIvFfpYIf/////////8AAAAAAAAAATznDSkAAABACD1Q1IWv8UnBj7YBwTe/bL9Tit2zfm4Mnzy8/z7qVZHEtNlkl9mSQ9CG1IbRVnjgc7c4sswvgy/6kYuGBeiaAAAAAAEAAAABAAAAAQAAAACPJ4LNXXBHJ7BFm9HGelFM4sY3eFhjD0LV9ZEW4bobvQAAAAA=');


--
-- Data for Name: trustlines; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO trustlines VALUES ('GCSX4PDUZP3BL522ZVMFXCEJ55NKEOHEMII7PSMJZNAAESJ444GSSJMO', 1, 'GAB7GMQPJ5YY2E4UJMLNAZPDEUKPK4AAIPRXIZHKZGUIRC6FP2LAQSDN', 'USD', 9223372036854775807, 0, 1, 3, NULL, NULL);
INSERT INTO trustlines VALUES ('GCYLTPOU7IVYHHA3XKQF4YB4W4ZWHFERMOQ7K47IWANKNBFBNJJNEOG5', 1, 'GCFZWN3AOVFQM2BZTZX7P47WSI4QMGJC62LILPKODTNDLVKZZNA5BQJ3', 'USD', 9223372036854775807, 0, 1, 3, NULL, NULL);


--
-- Data for Name: txfeehistory; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO txfeehistory VALUES ('c6c673dd3f0f5248f1e2c85bc88daebedafa4de71202bede4980667c10292821', 2, 1, 'AAAAAgAAAAMAAAABAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtrOnZAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtrOnY/+cAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==');
INSERT INTO txfeehistory VALUES ('790c7fd5024a5da930b7cd55ed31bff4dd4a279ee20a2c1fc089e97557573424', 2, 2, 'AAAAAgAAAAMAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtrOnY/+cAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtrOnY/84AAAAAAAAAAIAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==');
INSERT INTO txfeehistory VALUES ('834f6076a7870b4921347be0a6128487d900fabc1078be433f473b2f3139837f', 2, 3, 'AAAAAgAAAAMAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtrOnY/84AAAAAAAAAAIAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtrOnY/7UAAAAAAAAAAMAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==');
INSERT INTO txfeehistory VALUES ('ff8874bbd46e02dfe9b47aafd35b817d3f44bf769a0dff0bf73b16e6bc0a33ef', 2, 4, 'AAAAAgAAAAMAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtrOnY/7UAAAAAAAAAAMAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtrOnY/5wAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==');
INSERT INTO txfeehistory VALUES ('a407d1e59a681499d7b1af4752df7b73953c631772eb40ca758989b96018d692', 3, 1, 'AAAAAgAAAAMAAAACAAAAAAAAAACwub3U+iuDnBu6oF5gPLczY5SRY6H1c+iwGqaEoWpS0gAAAAJUC+QAAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAACwub3U+iuDnBu6oF5gPLczY5SRY6H1c+iwGqaEoWpS0gAAAAJUC+OcAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==');
INSERT INTO txfeehistory VALUES ('3cb0f6b31e0a73f6c3a316d930f5deb4d9a825abd80310dcf0c587d7e12c7624', 3, 2, 'AAAAAgAAAAMAAAACAAAAAAAAAAClfjx0y/YV91rNWFuIie9aojjkYhH3yYnLQAJJPOcNKQAAAAJUC+QAAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAClfjx0y/YV91rNWFuIie9aojjkYhH3yYnLQAJJPOcNKQAAAAJUC+OcAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==');


--
-- Data for Name: txhistory; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO txhistory VALUES ('c6c673dd3f0f5248f1e2c85bc88daebedafa4de71202bede4980667c10292821', 2, 1, 'AAAAAGL8HQvQkbK2HA3WVjRrKmjX00fG8sLI7m0ERwJW/AX3AAAAZAAAAAAAAAABAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAsLm91Porg5wbuqBeYDy3M2OUkWOh9XPosBqmhKFqUtIAAAACVAvkAAAAAAAAAAABVvwF9wAAAEDr7G099qIL7nPbxze2xCgJGbUVZ4rvCgTyz+zsKbHqYfSJAlqq+1LJAUIH9fbjTjbNObVelKRa90ri/p2RHXIJ', 'xsZz3T8PUkjx4shbyI2uvtr6TecSAr7eSYBmfBApKCEAAAAAAAAAZAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAA==', 'AAAAAQAAAAAAAAABAAAAAwAAAAMAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtrOnY/5wAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtrFTWBpwAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAAAAAAAAAACwub3U+iuDnBu6oF5gPLczY5SRY6H1c+iwGqaEoWpS0gAAAAJUC+QAAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==');
INSERT INTO txhistory VALUES ('790c7fd5024a5da930b7cd55ed31bff4dd4a279ee20a2c1fc089e97557573424', 2, 2, 'AAAAAGL8HQvQkbK2HA3WVjRrKmjX00fG8sLI7m0ERwJW/AX3AAAAZAAAAAAAAAACAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAApX48dMv2FfdazVhbiInvWqI45GIR98mJy0ACSTznDSkAAAACVAvkAAAAAAAAAAABVvwF9wAAAEAquSOIytKAx1/dElrYoSuRdTbRtVMQvtE9gAa4uM/elgMLXsdxSsar7SZL+p7I/+J8ZIwivYxnGjOXipt5f4UJ', 'eQx/1QJKXakwt81V7TG/9N1KJ57iCiwfwInpdVdXNCQAAAAAAAAAZAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAA==', 'AAAAAQAAAAAAAAABAAAAAwAAAAMAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtrFTWBpwAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtq7/TDZwAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAAAAAAAAAAClfjx0y/YV91rNWFuIie9aojjkYhH3yYnLQAJJPOcNKQAAAAJUC+QAAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==');
INSERT INTO txhistory VALUES ('834f6076a7870b4921347be0a6128487d900fabc1078be433f473b2f3139837f', 2, 3, 'AAAAAGL8HQvQkbK2HA3WVjRrKmjX00fG8sLI7m0ERwJW/AX3AAAAZAAAAAAAAAADAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAi5s3YHVLBmg5nm/38/aSOQYZIvaWhb1OHNo11VnLQdAAAAACVAvkAAAAAAAAAAABVvwF9wAAAEDSRx2tkaPmowO5+GO4a5qEZ6/6BwbkoAK+N9Ad4Ie8Gu6f4XWu7RjxQ7fbuLLW6kipomxVIfFJ8+Tiv5Zd4UME', 'g09gdqeHC0khNHvgphKEh9kA+rwQeL5DP0c7LzE5g38AAAAAAAAAZAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAA==', 'AAAAAQAAAAAAAAABAAAAAwAAAAMAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtq7/TDZwAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtqyrQFJwAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAAAAAAAAAACLmzdgdUsGaDmeb/fz9pI5Bhki9paFvU4c2jXVWctB0AAAAAJUC+QAAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==');
INSERT INTO txhistory VALUES ('ff8874bbd46e02dfe9b47aafd35b817d3f44bf769a0dff0bf73b16e6bc0a33ef', 2, 4, 'AAAAAGL8HQvQkbK2HA3WVjRrKmjX00fG8sLI7m0ERwJW/AX3AAAAZAAAAAAAAAAEAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAA/MyD09xjROUSxbQZeMlFPVwAEPjdGTqyaiIi8V+lggAAAACVAvkAAAAAAAAAAABVvwF9wAAAEABI+RIBTA9OBOHPtApTdYY2ABBRoR55l4dwEYszn8laVt52bZNa+1NfBVuZffVTyHbaww6Q0/sfo2OfF5c0HwN', '/4h0u9RuAt/ptHqv01uBfT9Ev3aaDf8L9zsW5rwKM+8AAAAAAAAAZAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAA==', 'AAAAAQAAAAAAAAABAAAAAwAAAAMAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtqyrQFJwAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtqpXNG5wAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAAAAAAAAAAAD8zIPT3GNE5RLFtBl4yUU9XAAQ+N0ZOrJqIiLxX6WCAAAAAJUC+QAAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==');
INSERT INTO txhistory VALUES ('a407d1e59a681499d7b1af4752df7b73953c631772eb40ca758989b96018d692', 3, 1, 'AAAAALC5vdT6K4OcG7qgXmA8tzNjlJFjofVz6LAapoShalLSAAAAZAAAAAIAAAABAAAAAAAAAAAAAAABAAAAAAAAAAYAAAABVVNEAAAAAACLmzdgdUsGaDmeb/fz9pI5Bhki9paFvU4c2jXVWctB0H//////////AAAAAAAAAAGhalLSAAAAQCnZlvEITlesuWctD/z9XMvJSujWqCYBR5F6tzeyStpEqiQXeyOEtJrGO3ulubYfi4kuwxETF5+CORR/2+0IKA4=', 'pAfR5ZpoFJnXsa9HUt97c5U8Yxdy60DKdYmJuWAY1pIAAAAAAAAAZAAAAAAAAAABAAAAAAAAAAYAAAAAAAAAAA==', 'AAAAAQAAAAIAAAADAAAAAwAAAAAAAAAAsLm91Porg5wbuqBeYDy3M2OUkWOh9XPosBqmhKFqUtIAAAACVAvjnAAAAAIAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAAAAAAAsLm91Porg5wbuqBeYDy3M2OUkWOh9XPosBqmhKFqUtIAAAACVAvjnAAAAAIAAAABAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAMAAAADAAAAAAAAAACwub3U+iuDnBu6oF5gPLczY5SRY6H1c+iwGqaEoWpS0gAAAAJUC+OcAAAAAgAAAAEAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAACwub3U+iuDnBu6oF5gPLczY5SRY6H1c+iwGqaEoWpS0gAAAAJUC+OcAAAAAgAAAAEAAAABAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAAAAAQAAAACwub3U+iuDnBu6oF5gPLczY5SRY6H1c+iwGqaEoWpS0gAAAAFVU0QAAAAAAIubN2B1SwZoOZ5v9/P2kjkGGSL2loW9ThzaNdVZy0HQAAAAAAAAAAB//////////wAAAAEAAAAAAAAAAA==');
INSERT INTO txhistory VALUES ('3cb0f6b31e0a73f6c3a316d930f5deb4d9a825abd80310dcf0c587d7e12c7624', 3, 2, 'AAAAAKV+PHTL9hX3Ws1YW4iJ71qiOORiEffJictAAkk85w0pAAAAZAAAAAIAAAABAAAAAAAAAAAAAAABAAAAAAAAAAYAAAABVVNEAAAAAAAD8zIPT3GNE5RLFtBl4yUU9XAAQ+N0ZOrJqIiLxX6WCH//////////AAAAAAAAAAE85w0pAAAAQAg9UNSFr/FJwY+2AcE3v2y/U4rds35uDJ88vP8+6lWRxLTZZJfZkkPQhtSG0VZ44HO3OLLML4Mv+pGLhgXomgA=', 'PLD2sx4Kc/bDoxbZMPXetNmoJavYAxDc8MWH1+EsdiQAAAAAAAAAZAAAAAAAAAABAAAAAAAAAAYAAAAAAAAAAA==', 'AAAAAQAAAAIAAAADAAAAAwAAAAAAAAAApX48dMv2FfdazVhbiInvWqI45GIR98mJy0ACSTznDSkAAAACVAvjnAAAAAIAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAAAAAAApX48dMv2FfdazVhbiInvWqI45GIR98mJy0ACSTznDSkAAAACVAvjnAAAAAIAAAABAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAMAAAADAAAAAAAAAAClfjx0y/YV91rNWFuIie9aojjkYhH3yYnLQAJJPOcNKQAAAAJUC+OcAAAAAgAAAAEAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAADAAAAAAAAAAClfjx0y/YV91rNWFuIie9aojjkYhH3yYnLQAJJPOcNKQAAAAJUC+OcAAAAAgAAAAEAAAABAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAAAAAQAAAAClfjx0y/YV91rNWFuIie9aojjkYhH3yYnLQAJJPOcNKQAAAAFVU0QAAAAAAAPzMg9PcY0TlEsW0GXjJRT1cABD43Rk6smoiIvFfpYIAAAAAAAAAAB//////////wAAAAEAAAAAAAAAAA==');


--
-- Data for Name: upgradehistory; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO upgradehistory VALUES (2, 1, 'AAAAAQAAAAs=', 'AAAAAA==');
INSERT INTO upgradehistory VALUES (2, 2, 'AAAAAwAPQkA=', 'AAAAAA==');


--
-- Name: accountdata accountdata_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY accountdata
    ADD CONSTRAINT accountdata_pkey PRIMARY KEY (accountid, dataname);


--
-- Name: accounts accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY accounts
    ADD CONSTRAINT accounts_pkey PRIMARY KEY (accountid);


--
-- Name: ban ban_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY ban
    ADD CONSTRAINT ban_pkey PRIMARY KEY (nodeid);


--
-- Name: ledgerheaders ledgerheaders_ledgerseq_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY ledgerheaders
    ADD CONSTRAINT ledgerheaders_ledgerseq_key UNIQUE (ledgerseq);


--
-- Name: ledgerheaders ledgerheaders_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY ledgerheaders
    ADD CONSTRAINT ledgerheaders_pkey PRIMARY KEY (ledgerhash);


--
-- Name: offers offers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY offers
    ADD CONSTRAINT offers_pkey PRIMARY KEY (offerid);


--
-- Name: peers peers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY peers
    ADD CONSTRAINT peers_pkey PRIMARY KEY (ip, port);


--
-- Name: publishqueue publishqueue_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY publishqueue
    ADD CONSTRAINT publishqueue_pkey PRIMARY KEY (ledger);


--
-- Name: pubsub pubsub_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY pubsub
    ADD CONSTRAINT pubsub_pkey PRIMARY KEY (resid);


--
-- Name: quoruminfo quoruminfo_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY quoruminfo
    ADD CONSTRAINT quoruminfo_pkey PRIMARY KEY (nodeid);


--
-- Name: scpquorums scpquorums_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY scpquorums
    ADD CONSTRAINT scpquorums_pkey PRIMARY KEY (qsethash);


--
-- Name: storestate storestate_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY storestate
    ADD CONSTRAINT storestate_pkey PRIMARY KEY (statename);


--
-- Name: trustlines trustlines_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY trustlines
    ADD CONSTRAINT trustlines_pkey PRIMARY KEY (accountid, issuer, assetcode);


--
-- Name: txfeehistory txfeehistory_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY txfeehistory
    ADD CONSTRAINT txfeehistory_pkey PRIMARY KEY (ledgerseq, txindex);


--
-- Name: txhistory txhistory_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY txhistory
    ADD CONSTRAINT txhistory_pkey PRIMARY KEY (ledgerseq, txindex);


--
-- Name: upgradehistory upgradehistory_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY upgradehistory
    ADD CONSTRAINT upgradehistory_pkey PRIMARY KEY (ledgerseq, upgradeindex);


--
-- Name: accountbalances; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX accountbalances ON accounts USING btree (balance) WHERE (balance >= 1000000000);


--
-- Name: bestofferindex; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX bestofferindex ON offers USING btree (sellingasset, buyingasset, price);


--
-- Name: histbyseq; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX histbyseq ON txhistory USING btree (ledgerseq);


--
-- Name: histfeebyseq; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX histfeebyseq ON txfeehistory USING btree (ledgerseq);


--
-- Name: ledgersbyseq; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ledgersbyseq ON ledgerheaders USING btree (ledgerseq);


--
-- Name: scpenvsbyseq; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX scpenvsbyseq ON scphistory USING btree (ledgerseq);


--
-- Name: scpquorumsbyseq; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX scpquorumsbyseq ON scpquorums USING btree (lastledgerseq);


--
-- Name: upgradehistbyseq; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX upgradehistbyseq ON upgradehistory USING btree (ledgerseq);


--
-- PostgreSQL database dump complete
--

