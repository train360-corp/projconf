set check_function_bodies = off;

CREATE OR REPLACE FUNCTION public.variables_after_actions()
    RETURNS trigger
    security definer
    LANGUAGE plpgsql
AS
$function$
declare
    env public.environments%rowtype;
begin

    if tg_op = 'INSERT' then
        for env in (select * from public.environments where project_id = new.project_id)
            loop
                insert into public.secrets(project_id, variable_id, environment_id)
                values (new.project_id, new.id, env.id);
            end loop;
    end if;

    return coalesce(new, old);

end;
$function$
;

CREATE OR REPLACE FUNCTION public.environments_after_actions()
    RETURNS trigger
    security definer
    LANGUAGE plpgsql
AS
$function$
declare
    var public.variables%rowtype;
begin

    if tg_op = 'INSERT' then
        for var in (select * from public.variables where project_id = new.project_id)
            loop
                insert into public.secrets(project_id, variable_id, environment_id)
                values (new.project_id, var.id, new.id);
            end loop;
    end if;

    return coalesce(new, old);

end;
$function$;

CREATE TRIGGER variables_after_actions
    AFTER INSERT OR DELETE OR UPDATE
    ON public.variables
    FOR EACH ROW
EXECUTE FUNCTION variables_after_actions();
CREATE TRIGGER environments_after_actions
    AFTER INSERT OR DELETE OR UPDATE
    ON public.environments
    FOR EACH ROW
EXECUTE FUNCTION environments_after_actions();

CREATE OR REPLACE FUNCTION public.secrets_before_actions()
    RETURNS trigger
    security definer
    LANGUAGE plpgsql
AS
$function$
declare
    proj public.projects%rowtype;
    var  public.variables%rowtype;
    env  public.environments%rowtype;
begin
    if tg_op = 'INSERT' then
        select * from public.projects p where p.id = new.project_id limit 1 into proj;
        select * from public.variables v where v.id = new.variable_id limit 1 into var;
        select * from public.environments e where e.id = new.environment_id limit 1 into env;

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
        new.project_id := old.project_id;
        new.variable_id := old.variable_id;
        new.environment_id := old.environment_id;
    end if;

    return coalesce(new, old);

end;
$function$;

CREATE TRIGGER secrets_before_actions
    BEFORE INSERT OR DELETE OR UPDATE
    ON public.secrets
    FOR EACH ROW
EXECUTE FUNCTION secrets_before_actions();