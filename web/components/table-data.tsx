import { Database, Tables } from "@/lib/supabase/types.gen";
import { Table as SupabaseTable } from "@/lib/supabase/types";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ComponentProps } from "react";
import { createClient } from "@/lib/supabase/server";
import { PostgrestFilterBuilder } from "@supabase/postgrest-js";
import { toast } from "sonner";
import { PostgrestResponse } from "@supabase/supabase-js";
import { cn } from "@/lib/utils";



type Filter<T extends SupabaseTable> = PostgrestFilterBuilder<object, Database["public"], Database["public"]["Tables"][T]["Row"], T, []>

type Column<T extends SupabaseTable, K extends keyof Tables<T>> = {
  header: string;
  className?: Pick<ComponentProps<"th">, "className"> & Pick<ComponentProps<"tr">, "className">;
  formatter?: (column: Tables<T>[K]) => string;
}

type Columns<T extends SupabaseTable> = {
  [K in keyof Tables<T>]?: Column<T, K>
};

type Props<T extends SupabaseTable> = {
  table: T;
  columns: Columns<T>;
  requestTransformer?: (filter: Filter<T>) => Filter<T>;
}

export const TableData = async <T extends SupabaseTable>(props: Props<T>) => {

  const supabase = await createClient();
  let query = supabase.from(props.table).select() as unknown as Filter<T>;
  if (props.requestTransformer) query = props.requestTransformer(query);
  const rows = await query as unknown as PostgrestResponse<Tables<T>>;

  if (rows.error) {
    console.error(rows.error);
    toast.error("An error occurred while fetching tables", {
      description: rows.error.message
    });
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          {Object.keys(props.columns).map((column, key) => {
            const col = props.columns[column as keyof Columns<T>]!;
            return (
              <TableHead
                key={key}
                className={cn(
                  "font-bold",
                  col.className
                )}
              >
                {col.header}
              </TableHead>
            );
          })}
        </TableRow>
      </TableHeader>
      <TableBody>

        {(rows.data ?? []).map((row, key) => (
          <TableRow key={key}>
            {Object.keys(props.columns).map((column, key) => {
              const col = props.columns[column as keyof Columns<T>]!;
              return (
                <TableCell
                  key={key}
                  className={col.className as string | undefined}
                >
                  {col.formatter ? col.formatter(row[column as keyof Tables<T>]) : row[column as keyof Tables<T>] as string}
                </TableCell>
              );
            })}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );

};