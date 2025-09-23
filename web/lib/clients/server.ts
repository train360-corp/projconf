import { createClient as Client } from "@train360-corp/projconf";



export const createServerClient = () => Client({
  adminApiKey: process.env.X_ADMIN_API_KEY,
  baseURL: process.env.PROJCONF_URL
});