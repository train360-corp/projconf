create policy "select based on RLS on projects"
on "public"."environments"
as permissive
for select
to anon
using ((EXISTS ( SELECT 1
   FROM projects p
  WHERE (p.id = environments.project_id))));


create policy "select based on RLS on projects"
on "public"."variables"
as permissive
for select
to anon
using ((EXISTS ( SELECT 1
   FROM projects p
  WHERE (p.id = variables.project_id))));



