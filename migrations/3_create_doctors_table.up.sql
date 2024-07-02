create table doctors
(
    id        text primary key,
    clinic_id text not null,
    constraint fk_clinic
        foreign key (clinic_id)
            references clinics (id)
);