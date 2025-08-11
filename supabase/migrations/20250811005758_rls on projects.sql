create policy "select based on RLS on clients"
on "public"."projects"
as permissive
for select
to anon
using ((EXISTS ( SELECT 1
   FROM clients cs
  WHERE (cs.project_id = projects.id))));



