create or replace function public.get_default_secret(
    variable_id uuid
)
    returns text
    security invoker
    language plpgsql
    set search_path to 'public'
as
$$
declare
    -- shared
    variable public.variables%rowtype;
    val      text := '';

    -- RANDOM
    len      int;
    charset  text := '';
begin

    select * from public.variables v where v.id = variable_id limit 1 into variable;
    if variable.id is null then
        raise exception 'variable (id=%) not found', variable_id;
    end if;

    case variable.generator_type
        when 'STATIC'::generator then val := variable.generator_data #>> '{}';
        when 'RANDOM'::generator then
            IF (variable.generator_data ->> 'letters')::boolean THEN
                charset := charset || 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ';
            END IF;
            IF (variable.generator_data ->> 'numbers')::boolean THEN
              charset := charset || '0123456789';
          END IF;
          IF (variable.generator_data ->> 'symbols')::boolean THEN
              charset := charset || '!@#$%^&*()-_=+[]{};:,.<>?';
          END IF;

          IF charset = '' THEN
              RAISE EXCEPTION 'charset empty!';
          END IF;

            len := (variable.generator_data->>'length')::int;

          FOR i IN 1..len LOOP
                  val := val || substr(charset, floor(random() * length(charset) + 1)::int, 1);
              END LOOP;


        else raise exception 'unhandled generator type (%)', variable.generator_type;
        end case;

    if val is null then
        raise exception 'val unexpectedly null (generator=%)', variable.generator_type;
    end if;

    return val;
end;
$$;