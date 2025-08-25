set check_function_bodies = off;

CREATE OR REPLACE FUNCTION public.environments_after_actions()
 RETURNS trigger
 LANGUAGE plpgsql
 SECURITY DEFINER
AS $function$declare
    var public.variables%rowtype;
begin

    if tg_op = 'INSERT' then
        for var in (select * from public.variables where project_id = new.project_id)
            loop
                insert into public.secrets(variable_id, environment_id)
                values (var.id, new.id);
            end loop;
    end if;

    return coalesce(new, old);

end;$function$
;

CREATE OR REPLACE FUNCTION public.variables_after_actions()
 RETURNS trigger
 LANGUAGE plpgsql
 SECURITY DEFINER
AS $function$declare
    env public.environments%rowtype;
begin

    if tg_op = 'INSERT' then
        for env in (select * from public.environments where project_id = new.project_id)
            loop
                insert into public.secrets(variable_id, environment_id)
                values (new.id, env.id);
            end loop;
    end if;

    return coalesce(new, old);

end;$function$
;


