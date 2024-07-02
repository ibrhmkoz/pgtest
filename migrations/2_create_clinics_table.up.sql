create table clinics
(
    id    text primary key,
    name  text not null,
    email text not null unique
);