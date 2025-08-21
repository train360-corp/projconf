drop policy "select based on RLS on clients" on "public"."projects";

create policy "select based on RLS on clients"
on "public"."projects"
as permissive
for select
to anon
using ((( SELECT is_admin_client() AS is_admin_client) OR (EXISTS ( SELECT 1
   FROM clients cs
  WHERE (cs.project_id = projects.id)))));



