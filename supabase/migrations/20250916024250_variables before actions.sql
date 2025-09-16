set check_function_bodies = off;

CREATE OR REPLACE FUNCTION private.variables_before_actions()
 RETURNS trigger
 LANGUAGE plpgsql
 SET SEARCH_PATH = ''
AS $function$BEGIN

  IF TG_OP = 'INSERT' THEN

    -- for static, need to protect the secret by encrypting it
    IF NEW.generator_type = 'STATIC'::public.generator THEN

      -- force request type
      if not extensions.jsonb_matches_schema(
        schema := '{
          "type": "object",
          "properties": {
            "secret": {
              "type": "string"
            }
          },
          "required": [
            "secret"
          ],
          "additionalProperties": false
        }'::json,
        instance := NEW.generator_data
      ) then
        raise exception 'invalid format: must be an object with key "secret" of type "string"';
      end if;

      -- fix the secret to only store the encrypted id
      NEW.generator_data := jsonb_build_object(
        'secret-id', vault.create_secret(NEW.generator_data->>'secret')::text
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
