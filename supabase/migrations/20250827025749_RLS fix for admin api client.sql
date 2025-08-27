drop policy "select based on RLS on clients_secrets" on "public"."clients";

drop policy "select based on RLS on projects" on "public"."environments";

drop policy "select based on RLS on environments" on "public"."secrets";

drop policy "select based on RLS on projects" on "public"."variables";

create policy "select based on RLS on clients_secrets"
on "public"."clients"
as permissive
for select
to anon
using ((( SELECT is_admin_client() AS is_admin_client) OR (EXISTS ( SELECT 1
   FROM clients_secrets cs
  WHERE (cs.client_id = clients.id)))));


create policy "select based on RLS on projects"
on "public"."environments"
as permissive
for select
to anon
using ((( SELECT is_admin_client() AS is_admin_client) OR (EXISTS ( SELECT 1
   FROM clients c
  WHERE (c.environment_id = environments.id)))));


create policy "select based on RLS on environments"
on "public"."secrets"
as permissive
for select
to anon
using ((( SELECT is_admin_client() AS is_admin_client) OR (EXISTS ( SELECT 1
   FROM environments e
  WHERE (e.id = secrets.environment_id)))));


create policy "select based on RLS on projects"
on "public"."variables"
as permissive
for select
to anon
using ((( SELECT is_admin_client() AS is_admin_client) OR (EXISTS ( SELECT 1
   FROM projects p
  WHERE (p.id = variables.project_id)))));



