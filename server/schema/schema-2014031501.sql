CREATE TABLE flag (
    id SERIAL PRIMARY KEY,
    triggerid INTEGER,
    teamid INTEGER,
    code VARCHAR NOT NULL,
    flag VARCHAR,
    value INTEGER NOT NULL DEFAULT 0,
    writeup_value INTEGER NOT NULL DEFAULT 0,
    return_string VARCHAR,
    counter INTEGER,
    validator VARCHAR,
    description VARCHAR,
    tags VARCHAR
);

CREATE TABLE schema (
    version INTEGER PRIMARY KEY
);

CREATE TABLE score (
    id SERIAL PRIMARY KEY,
    teamid INTEGER NOT NULL,
    flagid INTEGER NOT NULL,
    value INTEGER NOT NULL DEFAULT 0,
    writeup_value INTEGER NOT NULL DEFAULT 0,
    submit_time TIMESTAMP WITH TIME ZONE,
    writeup_time TIMESTAMP WITH TIME ZONE
);

CREATE TABLE team (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    country VARCHAR(2),
    website VARCHAR(255),
    notes VARCHAR,
    subnets VARCHAR
);

CREATE TABLE trigger (
    id SERIAL PRIMARY KEY,
    flagid INTEGER NOT NULL,
    count VARCHAR NOT NULL,
    description VARCHAR
);

ALTER TABLE flag ADD FOREIGN KEY (triggerid) REFERENCES trigger(id);
ALTER TABLE flag ADD FOREIGN KEY (teamid) REFERENCES team(id);
ALTER TABLE flag ADD CONSTRAINT code_teamid UNIQUE (code, teamid);
CREATE UNIQUE INDEX flag_code_teamid ON flag (code, teamid)
WHERE teamid IS NOT NULL;
CREATE UNIQUE INDEX flag_code ON flag (code) WHERE teamid IS NULL;

ALTER TABLE score ADD FOREIGN KEY (teamid) REFERENCES team(id);
ALTER TABLE score ADD FOREIGN KEY (flagid) REFERENCES flag(id);
ALTER TABLE score ADD CONSTRAINT teamid_flagid UNIQUE (teamid, flagid);

ALTER SEQUENCE team_id_seq MINVALUE 0;
ALTER SEQUENCE team_id_seq RESTART WITH 0;

ALTER TABLE trigger ADD FOREIGN KEY (flagid) REFERENCES flag(id);

INSERT INTO schema (version) VALUES (2014031501);
