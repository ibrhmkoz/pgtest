create schema if not exists "public";

create table clinics
(
    id    text primary key,
    name  text not null,
    email text not null unique
);

create table doctors
(
    id        text primary key,
    clinic_id text not null,
    constraint fk_clinic
        foreign key (clinic_id)
            references clinics (id)
);
