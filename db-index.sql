CREATE DATABASE index;

CREATE TABLE words(
    w_id serial PRIMARY KEY,
    word text
);

CREATE TABLE files(
    f_id serial PRIMARY KEY,
    name_file text
);

CREATE TABLE positions(
    w_id integer REFERENCES words(w_id),
    f_id integer REFERENCES files(f_id),
    position integer
);

