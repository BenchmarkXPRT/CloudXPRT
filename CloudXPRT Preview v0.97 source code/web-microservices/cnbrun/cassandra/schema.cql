CREATE KEYSPACE cnb WITH replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };

CREATE TABLE cnb.user(
	id text,
	password text,
	email text,
	PRIMARY KEY(id)
);

CREATE TYPE cnb.results(
	stockprice double,
	strikeprice double,
	optionyear double,
	callresult double,
	callconfidence double
);

CREATE TABLE cnb.montecarlo(
	name text,
	date text,
	elapsed double,
	ops double,
	results list<frozen<results>>,
	PRIMARY KEY(name, date)
);
