create policy "select based on request headers"
on "public"."clients_secrets"
as permissive
for select
to anon
using ((((((current_setting('request.headers'::text, true))::json ->> 'x-client-secret-id'::text))::uuid = id) AND verify_client_secret((((current_setting('request.headers'::text, true))::json ->> 'x-client-secret-id'::text))::uuid, ((current_setting('request.headers'::text, true))::json ->> 'x-client-secret'::text))));



