"use client";

import { z } from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";

import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle, DialogTrigger, } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage, } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Checkbox } from "@/components/ui/checkbox";
import { ReactNode, useState } from "react";
import { Tables, TablesInsert } from "@/lib/supabase/types.gen";
import { WithSupabaseEnv } from "@/lib/supabase/types";
import { createClient } from "@/lib/supabase/client";
import { toast } from "sonner";

// ---- Schema ----
const keySchema = z
  .string()
  .regex(/^[A-Z_][A-Z0-9_]*$/, {
    message: "Must start with a capital letter/underscore and contain only A–Z, 0–9, _",
  });

const randomSchema = z
  .object({
    type: z.literal("RANDOM"),
    length: z.coerce.number().int().min(1, "Length must be at least 1"),
    letters: z.boolean(),
    numbers: z.boolean(),
    symbols: z.boolean(),
  })
  .refine((v) => v.letters || v.numbers || v.symbols, {
    message: "Select at least one character set",
    path: [ "letters" ], // attach the message near the first checkbox
  });

const staticSchema = z.object({
  type: z.literal("STATIC"),
  value: z.string().optional(), // no validation
});


const schema = z
  .object({
    key: keySchema,
  })
  .and(z.discriminatedUnion("type", [staticSchema, randomSchema]));

type FormValues = z.infer<typeof schema>;

export function DialogCreateVariable({ children, ...props }: WithSupabaseEnv<{
  children: ReactNode;
  project: Pick<Tables<"projects">, "id">;
  onCreate: "RELOAD" | "CLOSE";
}>) {

  const [ open, setOpen ] = useState(false);

  const form = useForm<FormValues>({
    // @ts-expect-error - types exist when flips to RANDOM
    resolver: zodResolver(schema),
    defaultValues: {
      key: "",
      type: "STATIC" as "STATIC" | "RANDOM",
      value: "",
      length: 32,
      letters: true,
      numbers: true,
      symbols: false,
    },
  });

  const type = form.watch("type");

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>

      <DialogContent>
        <DialogHeader>
          <DialogTitle>Add Variable</DialogTitle>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(async (v) => {

            const values = v as unknown as FormValues;

            let genData: TablesInsert<"variables">["generator_data"];
            switch (values.type) {
              case "STATIC":
                genData = values.value ?? "";
                break;
              case "RANDOM":
                genData = {
                  "length": values.length,
                  "letters": values.letters,
                  "numbers": values.numbers,
                  "symbols": values.symbols
                };
                break;
              default:
                // @ts-expect-error - catch-all
                throw new Error(`type unhandled: ${values.type}`);
            }

            const r = await createClient(props).from("variables").insert({
              generator_type: values.type,
              generator_data: genData,
              key: values.key,
              project_id: props.project.id,
            }).select().single();

            if(r.error) toast.error("Unable to Create Variable", {
              description: r.error.message
            });
            else {

              switch (props.onCreate) {
                case "RELOAD":
                  window.location.reload();
                  break;
                case "CLOSE":
                  setOpen(false);
                  form.reset(); // reset to defaults for next open
                  break;
                default:
                  throw new Error(`action unhandled: ${props.onCreate}`)
              }
            }
          })} className="space-y-5">

            <FormField
              // @ts-expect-error – see note above
              control={form.control}
              name="key"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Key</FormLabel>
                  <FormControl>
                    <Input placeholder="MY_VARIABLE" {...field} />
                  </FormControl>
                  <FormDescription>Must match ^[A-Z_][A-Z0-9_]*$</FormDescription>
                  <FormMessage /> {/* will show regex / required errors */}
                </FormItem>
              )}
            />

            {/* Type switch (Static | Random) */}
            <FormField
              // @ts-expect-error – see note above
              control={form.control}
              name="type"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Type</FormLabel>
                  <Tabs
                    value={field.value}
                    onValueChange={(val) => field.onChange(val)}
                    className="w-full"
                  >
                    <TabsList className="grid w-full grid-cols-2">
                      <TabsTrigger value="STATIC">Static</TabsTrigger>
                      <TabsTrigger value="RANDOM">Random</TabsTrigger>
                    </TabsList>
                  </Tabs>
                </FormItem>
              )}
            />

            {/* Static form */}
            {type === "STATIC" && (
              <FormField
                // @ts-expect-error – see note above
                control={form.control}
                name="value" // only valid for STATIC branch
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Value</FormLabel>
                    <FormControl>
                      <Input disabled={form.formState.isSubmitting} placeholder="Enter static value…" {...field} />
                    </FormControl>
                    <FormDescription>No validation is applied for static values.</FormDescription>
                    <FormMessage/>
                  </FormItem>
                )}
              />
            )}

            {/* Random form */}
            {type === "RANDOM" && (
              <div className="space-y-4">
                <FormField
                  // @ts-expect-error – see note above
                  control={form.control}
                  name="length"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Length</FormLabel>
                      <FormControl>
                        <Input disabled={form.formState.isSubmitting} type="number" min={1} placeholder="32" {...field} />
                      </FormControl>
                      <FormMessage/>
                    </FormItem>
                  )}
                />

                <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
                  <FormField
                    // @ts-expect-error – see note above
                    control={form.control}
                    name="letters"
                    render={({ field }) => (
                      <FormItem className="flex items-center gap-3 space-y-0 rounded-md border p-3">
                        <FormControl>
                          <Checkbox
                            disabled={form.formState.isSubmitting}
                            checked={field.value}
                            onCheckedChange={(v) => field.onChange(Boolean(v))}
                          />
                        </FormControl>
                        <div className="leading-none">
                          <FormLabel className="text-sm font-medium">Letters</FormLabel>
                          <FormDescription className="text-xs">a–z, A–Z</FormDescription>
                        </div>
                      </FormItem>
                    )}
                  />

                  <FormField
                    // @ts-expect-error – see note above
                    control={form.control}
                    name="numbers"
                    render={({ field }) => (
                      <FormItem className="flex items-center gap-3 space-y-0 rounded-md border p-3">
                        <FormControl>
                          <Checkbox
                            disabled={form.formState.isSubmitting}
                            checked={field.value}
                            onCheckedChange={(v) => field.onChange(Boolean(v))}
                          />
                        </FormControl>
                        <div className="leading-none">
                          <FormLabel className="text-sm font-medium">Numbers</FormLabel>
                          <FormDescription className="text-xs">0–9</FormDescription>
                        </div>
                      </FormItem>
                    )}
                  />

                  <FormField
                    // @ts-expect-error – see note above
                    control={form.control}
                    name="symbols"
                    render={({ field }) => (
                      <FormItem className="flex items-center gap-3 space-y-0 rounded-md border p-3">
                        <FormControl>
                          <Checkbox
                            disabled={form.formState.isSubmitting}
                            checked={field.value}
                            onCheckedChange={(v) => field.onChange(Boolean(v))}
                          />
                        </FormControl>
                        <div className="leading-none">
                          <FormLabel className="text-sm font-medium">Symbols</FormLabel>
                          <FormDescription className="text-xs">!@#$…</FormDescription>
                        </div>
                      </FormItem>
                    )}
                  />
                </div>

                {/* Show union-level error (e.g., "Select at least one") */}
                <FormMessage>
                  {/* @ts-expect-error – see note above*/}
                  {form.formState.errors?.letters?.message as ReactNode}
                </FormMessage>
              </div>
            )}

            <DialogFooter>
              <Button type="submit" disabled={form.formState.isSubmitting || !form.formState.isValid}>
                {form.formState.isSubmitting ? "Adding…" : "Add Variable"}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}