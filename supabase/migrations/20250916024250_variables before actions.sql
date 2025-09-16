set check_function_bodies = off;

CREATE OR REPLACE FUNCTION private.variables_before_actions()
 RETURNS trigger
 LANGUAGE plpgsql
AS $function$BEGIN

  IF TG_OP = 'INSERT' THEN

    -- for static, need to protect the secret by encrypting it
    IF NEW.generator_type = 'STATIC'::public.generator THEN
      NEW.generator_data := jsonb_set(
          jsonb_set(
            NEW.generator_data, 
            '{secret-id}', 
            to_jsonb(
              vault.create_secret(
                NEW.generator_data->>'secret'
              )::text  
            ), 
            true
          ),
          '{secret}',
          to_jsonb(''::text), 
          false
        );
    END IF;

  END IF;

  IF TG_OP = 'UPDATE' THEN

    NEW.generator_type = OLD.generator_type;
    NEW.generator_data = OLD.generator_data;

  END IF;

  RETURN COALESCE(NEW, OLD);

END;$function$
;


CREATE TRIGGER variables_before_actions BEFORE INSERT OR DELETE OR UPDATE ON public.variables FOR EACH ROW EXECUTE FUNCTION private.variables_before_actions();
