drop policy "insert based on admin-api-client" on "public"."clients";

drop policy "insert based on admin-api-client" on "public"."clients_secrets";

drop policy "insert based on admin-api-client" on "public"."environments";

drop policy "insert based on admin-api-client" on "public"."projects";

drop policy "insert based on admin-api-client" on "public"."secrets";

drop policy "insert based on admin-api-client" on "public"."variables";

-- clients
create policy "x-admin-api-key"
    on public.clients
    as permissive
    for all
    to anon
    using ((select is_admin_client() as is_admin_client));

-- clients_secrets
create policy "x-admin-api-key"
    on public.clients_secrets
    as permissive
    for all
    to anon
    using ((select is_admin_client() as is_admin_client));

-- environments
create policy "x-admin-api-key"
    on public.environments
    as permissive
    for all
    to anon
    using ((select is_admin_client() as is_admin_client));

-- projects
create policy "x-admin-api-key"
    on public.projects
    as permissive
    for all
    to anon
    using ((select is_admin_client() as is_admin_client));

-- secrets
create policy "x-admin-api-key"
    on public.secrets
    as permissive
    for all
    to anon
    using ((select is_admin_client() as is_admin_client));

-- variables
create policy "x-admin-api-key"
    on public.variables
    as permissive
    for all
    to anon
    using ((select is_admin_client() as is_admin_client));



