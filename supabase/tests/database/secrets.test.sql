begin;

select extensions.plan(5);
select extensions.has_column('secrets', 'id');
select extensions.has_pk('secrets', 'id');
select extensions.has_column('secrets', 'created_at');
select extensions.has_column('secrets', 'variable_id');
select extensions.has_column('secrets', 'environment_id');

rollback;