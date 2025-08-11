create policy "select based on RLS on clients_secrets"
on "public"."clients"
as permissive
for select
to anon
using ((EXISTS ( SELECT 1
   FROM clients_secrets cs
  WHERE (cs.client_id = clients.id))));



