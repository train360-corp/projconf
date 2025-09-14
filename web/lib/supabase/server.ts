import { createServerClient } from "@supabase/ssr";
import { cookies } from "next/headers";
import { Database } from "@/lib/supabase/types.gen";



export async function createClient() {
  const cookieStore = await cookies();

  return createServerClient<Database>(
    process.env.SUPABASE_BACKEND_URL,
    process.env.SUPABASE_PUBLISHABLE_OR_ANON_KEY,
    {
      global: {
        headers: {
          "x-admin-api-key": process.env.X_ADMIN_API_KEY,
        }
      },
      cookies: {
        getAll() {
          return cookieStore.getAll();
        },
        setAll(cookiesToSet) {
          try {
            cookiesToSet.forEach(({ name, value, options }) =>
              cookieStore.set(name, value, options)
            );
          } catch {
            // The `setAll` method was called from a Server Component.
            // This can be ignored if you have middleware refreshing
            // user sessions.
          }
        },
      },
    }
  );
}
