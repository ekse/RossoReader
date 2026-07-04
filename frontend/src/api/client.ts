import type {
  Feed,
  Item,
  ItemsResponse,
  DiscoveredFeed,
  User,
  Passkey,
  AdminSettings,
} from "@/types";

const BASE_URL = import.meta.env.VITE_API_URL || "";

class ApiError extends Error {
  response: { status: number; data: any };

  constructor(status: number, data: any) {
    super(`Request failed with status ${status}`);
    this.response = { status, data };
  }
}

async function request<T>(
  path: string,
  {
    method = "GET",
    body,
    params,
    responseType,
  }: {
    method?: string;
    body?: unknown;
    params?: Record<string, unknown>;
    responseType?: "json" | "blob";
  } = {},
): Promise<T> {
  let url = `${BASE_URL}${path}`;

  if (params) {
    const qs = new URLSearchParams();
    for (const [k, v] of Object.entries(params)) {
      if (v !== undefined) {
        qs.append(k, String(v));
      }
    }
    const qsStr = qs.toString();
    if (qsStr) {
      url += `?${qsStr}`;
    }
  }

  const headers: Record<string, string> = {};
  if (body !== undefined && !(body instanceof FormData)) {
    headers["Content-Type"] = "application/json";
  }

  const res = await fetch(url, {
    method,
    headers,
    body: body === undefined ? undefined : body instanceof FormData ? body : JSON.stringify(body),
    credentials: "include",
  });

  if (res.status === 401 && window.location.pathname !== "/login") {
    window.location.href = "/login";
  }

  if (!res.ok) {
    let data: any;
    try {
      data = await res.json();
    } catch {
      data = null;
    }
    throw new ApiError(res.status, data);
  }

  if (responseType === "blob") {
    return res.blob() as Promise<T>;
  }

  const text = await res.text();
  if (!text) return undefined as unknown as Promise<T>;
  return JSON.parse(text);
}

// Auth

export async function login(username: string, password: string): Promise<User> {
  return request<User>("/api/auth/login", {
    method: "POST",
    body: { username, password },
  });
}

export async function logout(): Promise<void> {
  await request<void>("/api/auth/logout", { method: "POST" });
}

export async function getMe(): Promise<User> {
  return request<User>("/api/auth/me");
}

export async function changePassword(currentPassword: string, newPassword: string): Promise<void> {
  await request<void>("/api/auth/password", {
    method: "PUT",
    body: { current_password: currentPassword, new_password: newPassword },
  });
}

export async function listUsers(): Promise<User[]> {
  return request<User[]>("/api/users");
}

export async function createUser(
  username: string,
  password: string,
  isAdmin: boolean,
): Promise<User> {
  return request<User>("/api/users", {
    method: "POST",
    body: { username, password, is_admin: isAdmin },
  });
}

export async function deleteUser(id: number): Promise<void> {
  await request<void>(`/api/users/${id}`, { method: "DELETE" });
}

// Feeds

export async function fetchFeeds(): Promise<Feed[]> {
  return request<Feed[]>("/api/feeds");
}

export async function addFeed(url: string): Promise<Feed> {
  return request<Feed>("/api/feeds", { method: "POST", body: { url } });
}

export async function discoverFeeds(url: string): Promise<DiscoveredFeed[]> {
  return request<DiscoveredFeed[]>("/api/feeds/discover", {
    method: "POST",
    body: { url },
  });
}

export async function deleteFeed(id: number): Promise<void> {
  await request<void>(`/api/feeds/${id}`, { method: "DELETE" });
}

export async function refreshFeed(id: number): Promise<void> {
  await request<void>(`/api/feeds/${id}/refresh`, { method: "POST" });
}

export async function markFeedRead(id: number): Promise<void> {
  await request<void>(`/api/feeds/${id}/read-all`, { method: "POST" });
}

export async function exportOpml(): Promise<Blob> {
  return request<Blob>("/api/feeds/opml/export", { responseType: "blob" });
}

export async function previewOpmlImport(file: File): Promise<DiscoveredFeed[]> {
  const form = new FormData();
  form.append("file", file);
  return request<DiscoveredFeed[]>("/api/feeds/opml/preview", {
    method: "POST",
    body: form,
  });
}

// Items

export async function fetchItems(params?: {
  page?: number;
  per_page?: number;
  feed_id?: number;
  read?: boolean;
  starred?: boolean;
}): Promise<ItemsResponse> {
  return request<ItemsResponse>("/api/items", { params });
}

export async function markAllItemsRead(): Promise<void> {
  await request<void>("/api/items/read-all", { method: "POST" });
}

export async function updateItem(
  id: number,
  data: { read?: boolean; starred?: boolean },
): Promise<Item> {
  return request<Item>(`/api/items/${id}`, {
    method: "PATCH",
    body: data,
  });
}

// Settings

export async function fetchSettings(): Promise<Record<string, string>> {
  return request<Record<string, string>>("/api/settings");
}

export async function updateSettings(
  data: Record<string, string>,
): Promise<Record<string, string>> {
  return request<Record<string, string>>("/api/settings", {
    method: "PATCH",
    body: data,
  });
}

// Admin Settings

export async function fetchAdminSettings(): Promise<AdminSettings> {
  return request<AdminSettings>("/api/admin/settings");
}

export async function updateAdminSettings(data: Partial<AdminSettings>): Promise<AdminSettings> {
  return request<AdminSettings>("/api/admin/settings", {
    method: "PATCH",
    body: data,
  });
}

// Passkeys

export async function passkeyRegisterBegin(): Promise<{ state_id: string; options: any }> {
  return request<{ state_id: string; options: any }>("/api/auth/passkey/register/begin", {
    method: "POST",
  });
}

export async function passkeyRegisterFinish(
  stateId: string,
  name: string,
  credential: any,
): Promise<Passkey> {
  return request<Passkey>("/api/auth/passkey/register/finish", {
    method: "POST",
    body: { state_id: stateId, name, credential },
  });
}

export async function passkeyLoginBegin(): Promise<{ state_id: string; options: any }> {
  return request<{ state_id: string; options: any }>("/api/auth/passkey/login/begin", {
    method: "POST",
  });
}

export async function passkeyLoginFinish(stateId: string, credential: any): Promise<User> {
  return request<User>("/api/auth/passkey/login/finish", {
    method: "POST",
    body: { state_id: stateId, credential },
  });
}

export async function listPasskeys(): Promise<Passkey[]> {
  return request<Passkey[]>("/api/auth/passkeys");
}

export async function deletePasskey(id: number): Promise<void> {
  await request<void>(`/api/auth/passkeys/${id}`, { method: "DELETE" });
}
