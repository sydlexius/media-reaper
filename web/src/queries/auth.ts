import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";

export interface User {
  id: string;
  username: string;
  role: string;
}

interface LoginCredentials {
  username: string;
  password: string;
}

async function fetchMe(): Promise<User | null> {
  const res = await fetch("/api/auth/me");
  if (res.status === 401) {
    return null;
  }
  if (!res.ok) {
    throw new Error("Failed to fetch user");
  }
  return res.json();
}

async function login(credentials: LoginCredentials): Promise<User> {
  const res = await fetch("/api/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(credentials),
  });
  if (!res.ok) {
    const data = await res.json();
    throw new Error(data.error || "Login failed");
  }
  return res.json();
}

async function logout(): Promise<void> {
  const res = await fetch("/api/auth/logout", { method: "POST" });
  if (!res.ok) {
    throw new Error("Logout failed");
  }
}

export function useAuth() {
  const queryClient = useQueryClient();

  const userQuery = useQuery({
    queryKey: ["auth", "me"],
    queryFn: fetchMe,
    retry: false,
    staleTime: 5 * 60 * 1000,
  });

  const loginMutation = useMutation({
    mutationFn: login,
    onSuccess: (user) => {
      queryClient.setQueryData(["auth", "me"], user);
    },
  });

  const logoutMutation = useMutation({
    mutationFn: logout,
    onSuccess: () => {
      queryClient.setQueryData(["auth", "me"], null);
    },
  });

  return {
    user: userQuery.data ?? null,
    isLoading: userQuery.isLoading,
    isAuthenticated: userQuery.data != null,
    login: loginMutation.mutateAsync,
    logout: logoutMutation.mutateAsync,
    loginError: loginMutation.error?.message ?? null,
    isLoggingIn: loginMutation.isPending,
  };
}
