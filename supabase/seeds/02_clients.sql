insert into clients (id, display, project_id)
values ('ACAE624A-99A8-4208-9E28-19B3195A269E'::uuid, 'ProdServer-01', '94AB1156-5B42-499F-B8AA-92CA45DFA180'::uuid);

-- secret: bmcjgYNQG3rFu935WfyoL5XMdi5RiSWR
insert into clients_secrets (id, hash, client_id)
values ('30ac0fe4-4207-4217-89b0-60e2ce9117ec'::uuid, '$2a$12$zf4fA05TBoJUn6QMWiLMcuEG4v5E1xKp4YolkNMmVHbZWyESyWzcO', 'ACAE624A-99A8-4208-9E28-19B3195A269E'::uuid);