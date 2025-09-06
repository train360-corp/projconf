import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle
} from "@/components/ui/card";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { PostgrestError } from "@supabase/supabase-js";
import { RefreshButton } from "@/components/csr/refresh-button";
import * as React from "react";
import { cn } from "@/lib/utils";

type Props = {
  error: string;
  description?: string;
} | {
  error: PostgrestError
}

export const Error = ({ className, ...props }: Props & Pick<React.ComponentProps<"div">, "className">) => (
  <div
    className={cn(
      "w-full h-full flex items-center justify-center",
      className
    )}
  >
    <div className={"w-full md:w-[80%] lg:w-1/2 p-4"}>
      <Card className={"pt-6"}>
        <CardHeader>
          <CardTitle>
            {"Uh, oh..."}
          </CardTitle>
          <CardDescription>
            {"An unexpected error occurred."}
          </CardDescription>
          <CardAction>
            <RefreshButton/>
          </CardAction>
        </CardHeader>
        <CardContent>
          <p>{typeof props.error === "string" ? props.error : props.error.message}</p>
        </CardContent>
        <CardFooter>
          <Accordion className={"w-full"} type="single" collapsible>
            <AccordionItem value={"details"}>
              <AccordionTrigger>{"Details"}</AccordionTrigger>
              <AccordionContent className={"flex flex-col"}>
                {"description" in props ? props.description : (props.error as PostgrestError).details?.split("\n").map((str, i) => (
                  <p key={i}>
                    {str}
                  </p>
                ))}
              </AccordionContent>
            </AccordionItem>
          </Accordion>
        </CardFooter>
      </Card>
    </div>

  </div>
)

export const FullPageError = (props: Props) => (
  <Error {...props} className={"w-screen h-screen"} />
);