begin;
select extensions.plan(3);

select extensions.has_column( 'projects', 'id' );
select extensions.col_is_pk( 'projects', 'id' );
select extensions.has_column( 'projects', 'display' );

select * from extensions.finish();
rollback;