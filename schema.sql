CREATE SCHEMA IF NOT EXISTS "public";

CREATE TABLE clinics
(
    id    TEXT PRIMARY KEY,
    name  TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE
);

CREATE TABLE doctors
(
    id        TEXT PRIMARY KEY,
    clinic_id TEXT NOT NULL,
    CONSTRAINT fk_clinic
        FOREIGN KEY (clinic_id)
            REFERENCES clinics (id)
);
