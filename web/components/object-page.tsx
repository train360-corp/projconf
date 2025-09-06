import { Badge } from "@/components/ui/badge";
import { ReactNode } from "react";


export default async function ObjectPage(props: {
  title: string;
  type: string;
  id: string;
  children?: ReactNode;
}) {
  return (
    <div className="flex flex-col px-8 pb-8 w-full gap-8">
      <div className={"flex flex-col gap-2"}>
        <p className={"text-5xl font-extrabold"}>{props.title}</p>
        <div className={"flex flex-row items-center gap-4"}>
          <Badge
            variant={"outline"}
            className="bg-blue-500 text-white dark:bg-blue-600"
          >
            {props.type}
          </Badge>
          <p className={"text-muted text-xs"}>{props.id}</p>
        </div>
      </div>

      {props.children}
    </div>
  );
}