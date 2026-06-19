const BASE = "/api/v1";

async function request<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const token =
    typeof window !== "undefined" ? localStorage.getItem("access_token") : null;

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };
  if (token) headers["Authorization"] = `Bearer ${token}`;

  const res = await fetch(`${BASE}${path}`, { ...options, headers });

  if (res.status === 401) {
    if (typeof window !== "undefined") {
      localStorage.removeItem("access_token");
      localStorage.removeItem("refresh_token");
      window.location.href = "/login";
    }
    throw new Error("Unauthorized");
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body?.error?.message || `Request failed: ${res.status}`);
  }

  return res.json();
}

export const api = {
  // Auth
  signup: (email: string, password: string, full_name: string) =>
    request("/auth/signup", {
      method: "POST",
      body: JSON.stringify({ email, password, full_name }),
    }),
  login: (email: string, password: string) =>
    request<{ access_token: string; refresh_token: string }>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),
  logout: (refresh_token: string) =>
    request("/auth/logout", {
      method: "POST",
      body: JSON.stringify({ refresh_token }),
    }),

  // Users
  listUsers: (limit = 20, offset = 0) =>
    request<{ users: any[]; total: number }>(
      `/users?limit=${limit}&offset=${offset}`
    ),
  getUser: (id: string) => request<any>(`/users/${id}`),
  updateUser: (id: string, full_name: string) =>
    request<any>(`/users/${id}`, {
      method: "PATCH",
      body: JSON.stringify({ full_name }),
    }),

  // Roles
  listRoles: () => request<any[]>("/roles"),
  createRole: (name: string, description: string) =>
    request<any>("/roles", {
      method: "POST",
      body: JSON.stringify({ name, description }),
    }),
  deleteRole: (id: string) =>
    request(`/roles/${id}`, { method: "DELETE" }),
  assignRole: (userId: string, roleId: string) =>
    request(`/users/${userId}/roles`, {
      method: "POST",
      body: JSON.stringify({ role_id: roleId }),
    }),
  removeRole: (userId: string, roleId: string) =>
    request(`/users/${userId}/roles/${roleId}`, { method: "DELETE" }),
  attachPermission: (roleId: string, permissionId: string) =>
    request(`/roles/${roleId}/permissions`, {
      method: "POST",
      body: JSON.stringify({ permission_id: permissionId }),
    }),

  // Permissions
  listPermissions: () => request<any[]>("/permissions"),
  createPermission: (
    name: string,
    resource: string,
    action: string,
    description: string
  ) =>
    request<any>("/permissions", {
      method: "POST",
      body: JSON.stringify({ name, resource, action, description }),
    }),
  deletePermission: (id: string) =>
    request(`/permissions/${id}`, { method: "DELETE" }),

  // Authorize check
  authorize: (userId: string, permission: string) =>
    request<{ allowed: boolean; permission: string }>("/authorize", {
      method: "POST",
      body: JSON.stringify({ user_id: userId, permission }),
    }),

  // Audit
  listAudit: (limit = 50, offset = 0) =>
    request<{ logs: any[]; total: number }>(
      `/audit?limit=${limit}&offset=${offset}`
    ),
};
