insert into clients (id, display, environment_id)
values ('ACAE624A-99A8-4208-9E28-19B3195A269E'::uuid, 'ProdServer01', '6B74DA6E-3690-401D-83A5-A8FE3C10FE94'::uuid),
       ('BC9EEC74-4B77-4592-8F14-C1E00B54E443'::uuid, 'Seeding Test 1', '8A306D91-B133-41B1-9AA2-45DF344CC501'::uuid);

-- secret: bmcjgYNQG3rFu935WfyoL5XMdi5RiSWR
insert into clients_secrets (id, hash, client_id)
values ('30ac0fe4-4207-4217-89b0-60e2ce9117ec'::uuid, '$2a$12$zf4fA05TBoJUn6QMWiLMcuEG4v5E1xKp4YolkNMmVHbZWyESyWzcO', 'ACAE624A-99A8-4208-9E28-19B3195A269E'::uuid);