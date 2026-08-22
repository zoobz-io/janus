/** The hosted login flow's URL under an API base — navigate the browser here. */
export const loginUrl = (api: string): string =>
  `${api.replace(/\/+$/, "")}/auth/login`;

/** A response settled to its body, or null when the caller isn't signed in. */
export const settle = async <T>(res: Response): Promise<T | null> => {
  if (res.status === 401) {
    return null;
  }
  if (!res.ok) {
    throw new Error(`janus answered ${res.status} for ${res.url}`);
  }
  return (await res.json()) as T;
};
