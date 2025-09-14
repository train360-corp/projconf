import { SupabaseRuntimeEnvironmentVariables } from "@/lib/supabase/types";



export const getSupabase = (): SupabaseRuntimeEnvironmentVariables => ({
  SUPABASE_FRONTEND_URL: process.env.SUPABASE_FRONTEND_URL,
  SUPABASE_BACKEND_URL: process.env.SUPABASE_BACKEND_URL,
  SUPABASE_PUBLISHABLE_OR_ANON_KEY: process.env.SUPABASE_PUBLISHABLE_OR_ANON_KEY,
  X_ADMIN_API_KEY: process.env.X_ADMIN_API_KEY,
})