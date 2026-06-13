CREATE TABLE
    site (
        id UUID NOT NULL DEFAULT uuidv7 (),
        name TEXT NOT NULL,
        is_on BOOLEAN NOT NULL DEFAULT FALSE,
        PRIMARY KEY (id)
    );