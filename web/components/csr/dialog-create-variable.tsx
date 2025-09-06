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

// ---- Schema ----
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

const schema = z.discriminatedUnion("type", [ staticSchema, randomSchema ]);
type FormValues = z.infer<typeof schema>;

// ---- Props your app can use to receive the result ----
type NewVariable =
  | { type: "STATIC"; value?: string }
  | { type: "RANDOM"; length: number; letters: boolean; numbers: boolean; symbols: boolean };

export function DialogCreateVariable({
                                    onCreateAction,
                                    children,
                                  }: {
  onCreateAction?: (variable: NewVariable) => Promise<void> | void;
  children: ReactNode; // usually a button you click to open the dialog
}) {

  const [open, setOpen] = useState(false);

  const form = useForm<FormValues>({
    // @ts-expect-error - types exist when flips to RANDOM
    resolver: zodResolver(schema),
    defaultValues: {
      type: "STATIC",
      // random defaults (used when type is RANDOM)
      // @ts-expect-error – these fields exist when type flips to RANDOM
      length: 32,
      letters: true,
      numbers: true,
      symbols: false,
    },
  });

  const onSubmit = async (values: FormValues) => {

    if (onCreateAction) await onCreateAction(values as NewVariable);
    setOpen(false);

    // reset to defaults for next open
    form.reset({
      type: "STATIC",
      // @ts-expect-error – see note above
      length: 32,
      letters: true,
      numbers: true,
      symbols: false,
    });
  };

  const type = form.watch("type");

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>

      <DialogContent>
        <DialogHeader>
          <DialogTitle>Add Variable</DialogTitle>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-5">
            {/* Type switch (Static | Random) */}
            <FormField
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
                control={form.control}
                name="value" // only valid for STATIC branch
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Value</FormLabel>
                    <FormControl>
                      <Input placeholder="Enter static value…" {...field} />
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
                  control={form.control}
                  name="length"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Length</FormLabel>
                      <FormControl>
                        <Input type="number" min={1} placeholder="32" {...field} />
                      </FormControl>
                      <FormMessage/>
                    </FormItem>
                  )}
                />

                <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
                  <FormField
                    control={form.control}
                    name="letters"
                    render={({ field }) => (
                      <FormItem className="flex items-center gap-3 space-y-0 rounded-md border p-3">
                        <FormControl>
                          <Checkbox
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
                    control={form.control}
                    name="numbers"
                    render={({ field }) => (
                      <FormItem className="flex items-center gap-3 space-y-0 rounded-md border p-3">
                        <FormControl>
                          <Checkbox
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
                    control={form.control}
                    name="symbols"
                    render={({ field }) => (
                      <FormItem className="flex items-center gap-3 space-y-0 rounded-md border p-3">
                        <FormControl>
                          <Checkbox
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
                  {form.formState.errors?.letters?.message as React.ReactNode}
                </FormMessage>
              </div>
            )}

            <DialogFooter>
              <Button type="submit" disabled={form.formState.isSubmitting}>
                {form.formState.isSubmitting ? "Adding…" : "Add Variable"}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}