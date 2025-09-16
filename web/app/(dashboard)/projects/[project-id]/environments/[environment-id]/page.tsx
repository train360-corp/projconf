import { createClient } from "@/lib/supabase/server";
import { Error } from "@/components/error-handling/error";
import ObjectPage from "@/components/object-page";



export default async function Page(props: Readonly<{
  params: Promise<{
    ["project-id"]: string;
    ["environment-id"]: string;
  }>;
}>) {

  const supabase = await createClient();
  const params = await props.params;
  const environment = await supabase.from("environments").select()
    .eq("id", params["environment-id"])
    .eq("project_id", params["project-id"])
    .single();

  if (environment.error) return (
    <Error
      error={"Failed to Load Environment"}
      details={environment.error}
    />
  );

  const secrets = await supabase.from("secrets").select("*,variable:variables(*)")
    .eq("environment_id", environment.data.id);
  if (secrets.error) return (
    <Error
      error={"Failed to Load Secrets"}
      details={secrets.error}
    />
  );

  return (
    <ObjectPage
      title={environment.data.display}
      type={"Environment"}
      id={environment.data.id}
    >
      <div>
        <p className={"text-2xl font-bold"}>{"Clients"}</p>
      </div>
    </ObjectPage>
  );

}