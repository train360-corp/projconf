import { createClient } from "@/lib/supabase/server";
import { FullPageError } from "@/components/error-handling/error";
import ObjectPage from "@/components/object-page";
import { Card } from "@/components/ui/card";
import { TableData } from "@/components/table-data";
import { DialogCreateVariable } from "@/components/csr/dialog-create-variable";
import React from "react";
import { Button } from "@/components/ui/button";
import { getSupabase } from "@/lib/supabase/getSupabase";



export default async function Page(props: Readonly<{
  params: Promise<{
    ["project-id"]: string;
  }>;
}>) {

  const supabase = await createClient();
  const params = await props.params;
  const project = await supabase.from("projects").select()
    .eq("id", params["project-id"])
    .single();

  if (project.error) return (
    <FullPageError
      error={"Fail to Load Project"}
      details={project.error}
    />
  );

  return (
    <ObjectPage
      title={project.data.display}
      type={"Project"}
      id={project.data.id}
    >

      <div className={"flex flex-col gap-2"}>

        <div className={"flex flex-row justify-between items-center"}>
          <p className={"text-2xl font-bold"}>{"Variables"}</p>

          <DialogCreateVariable
            onCreate={"RELOAD"}
            project={{ id: params["project-id"] }}
            supabase={getSupabase()}
          >
            <Button size={"sm"} variant="outline">{"Add Variable"}</Button>
          </DialogCreateVariable>
        </div>

        <Card>
          <TableData
            table={"variables"}
            requestTransformer={r => r.eq("project_id", params["project-id"])}
            columns={{
              key: {
                header: "Key"
              },
              description: {
                header: "Description",
              },
              generator_type: {
                header: "Type"
              },
              generator_data: {
                header: "Config",
                formatter: (v) => JSON.stringify(v)
              }
            }}
          />
        </Card>
      </div>

    </ObjectPage>
  );
}