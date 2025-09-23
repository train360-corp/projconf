import { SupabaseRuntimeEnvironmentVariables } from "@/lib/supabase/types";



declare global {
  namespace NodeJS {
    // Augment the built-in type
    interface ProcessEnv extends SupabaseRuntimeEnvironmentVariables {}
  }
}

// ensure this file is treated as a module
export {};