CREATE SCHEMA ports
    CREATE TABLE port_entries (
        id          bigint generated always as identity constraint port_entry_pkey primary key,
        slug        varchar(100) unique not null,
        code        varchar(100),
        name        varchar(200),
        city        varchar(200),
        province    varchar(200),
        country     varchar(200),
        alias       varchar(200)[],
        regions     varchar(200)[],
        latitude    numeric,
        longitude   numeric,
        timezone    varchar(200),
        unlocks     varchar(100)[]
    )
;