import { createBrowserClient } from "@supabase/ssr";
import { Database } from "@/lib/supabase/types.gen";
import { WithSupabaseEnv } from "@/lib/supabase/types";



export function createClient({ supabase: { SUPABASE_FRONTEND_URL, SUPABASE_PUBLISHABLE_OR_ANON_KEY, X_ADMIN_API_KEY } }: WithSupabaseEnv) {
  return createBrowserClient<Database>(
    SUPABASE_FRONTEND_URL,
    SUPABASE_PUBLISHABLE_OR_ANON_KEY,
    {
      global: {
        headers: {
          "x-admin-api-key": X_ADMIN_API_KEY,
        }
      },
    }
  );
}
