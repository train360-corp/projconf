drop policy "select based on RLS on projects" on "public"."secrets";

create policy "select based on RLS on environments"
on "public"."secrets"
as permissive
for select
to anon
using ((EXISTS ( SELECT 1
   FROM environments e
  WHERE (e.id = secrets.environment_id))));



