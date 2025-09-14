import { ENV_KEYS } from "@/lib/supabase/types";



/**
 * called once on app start
 */
export function register() {

  // check environment variables
  for (const envKey of ENV_KEYS) {
    if(!process.env[envKey]?.trim()) {
      throw new Error(`runtime environment variable "${envKey}" is not defined!`);
    }
  }

}