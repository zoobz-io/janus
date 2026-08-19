import type {
  DataTableFetchParams,
  DataTableFetchResult,
} from "@zoobzio/foundation/types/data/table";

const compare = (a: unknown, b: unknown): number => {
  if (a === null || a === undefined) {
    return b === null || b === undefined ? 0 : -1;
  }
  if (b === null || b === undefined) {
    return 1;
  }
  if (typeof a === "number" && typeof b === "number") {
    return a - b;
  }
  return String(a).localeCompare(String(b));
};

/**
 * Client-side paging over an already-complete row set, for endpoints that
 * return everything in one response. Sort here is real: the full list is in
 * hand, so sortable columns behave correctly across pages.
 */
export const toPage = <T extends Record<string, unknown>>(
  rows: T[],
  { page, pageSize, sortField, sortDirection }: DataTableFetchParams,
): DataTableFetchResult<T> => {
  const sorted = [...rows];
  if (sortField !== null) {
    sorted.sort((x, y) => {
      const order = compare(x[sortField], y[sortField]);
      return sortDirection === "desc" ? -order : order;
    });
  }
  const start = (page - 1) * pageSize;
  return {
    data: sorted.slice(start, start + pageSize),
    total: rows.length,
    pageCount: Math.max(1, Math.ceil(rows.length / pageSize)),
  };
};
