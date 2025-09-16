insert into variables (project_id, key, description, generator_type, generator_data)
values ('008D797C-73E8-42BB-9B45-B5F7AD71D60A'::uuid, 'HOST', 'IP address to bind to', 'STATIC', '{"secret": "127.0.0.1"}'),
       ('94AB1156-5B42-499F-B8AA-92CA45DFA180'::uuid, 'HOST', 'IP address to bind to', 'STATIC', '{"secret": "127.0.0.1"}'),
       ('94AB1156-5B42-499F-B8AA-92CA45DFA180'::uuid, 'PASSWORD', 'password for basic auth', 'RANDOM', '{"length":32,"letters":true,"numbers":true,"symbols":false}');