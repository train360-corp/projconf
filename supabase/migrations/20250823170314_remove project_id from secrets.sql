set check_function_bodies = off;

CREATE OR REPLACE FUNCTION public.secrets_before_actions()
 RETURNS trigger
 LANGUAGE plpgsql
 SECURITY DEFINER
AS $function$declare
    proj public.projects%rowtype;
    var  public.variables%rowtype;
    env  public.environments%rowtype;
begin
    if tg_op = 'INSERT' then
        select * from public.variables v where v.id = new.variable_id limit 1 into var;
        select * from public.environments e where e.id = new.environment_id limit 1 into env;
        select * from public.projects p where p.id = env.project_id limit 1 into proj;

        if proj.id is null or var.id is null or env.id is null then
            raise exception 'data-loading error';
        end if;

        if (proj.id <> var.project_id) OR (proj.id <> env.project_id) then
            raise exception 'project mismatch';
        end if;

        -- generate default value
        select public.get_default_secret(variable_id := var.id) into new.value;
    elsif tg_op = 'UPDATE' then
        new.id := old.id;
        new.variable_id := old.variable_id;
        new.environment_id := old.environment_id;
    end if;

    return coalesce(new, old);

end;$function$
;


