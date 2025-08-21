create policy "insert based on admin-api-client"
    on "public"."clients"
    as permissive
    for insert
    to anon
    with check ((SELECT is_admin_client() AS is_admin_client));

create policy "insert based on admin-api-client"
    on "public"."clients_secrets"
    as permissive
    for insert
    to anon
    with check ((SELECT is_admin_client() AS is_admin_client));

create policy "insert based on admin-api-client"
    on "public"."environments"
    as permissive
    for insert
    to anon
    with check ((SELECT is_admin_client() AS is_admin_client));

create policy "insert based on admin-api-client"
    on "public"."projects"
    as permissive
    for insert
    to anon
    with check ((SELECT is_admin_client() AS is_admin_client));

create policy "insert based on admin-api-client"
    on "public"."secrets"
    as permissive
    for insert
    to anon
    with check ((SELECT is_admin_client() AS is_admin_client));

create policy "insert based on admin-api-client"
    on "public"."variables"
    as permissive
    for insert
    to anon
    with check ((SELECT is_admin_client() AS is_admin_client));



