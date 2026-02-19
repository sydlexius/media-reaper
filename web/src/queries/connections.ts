import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";

export interface Connection {
  id: string;
  name: string;
  type: "sonarr" | "radarr" | "emby";
  url: string;
  maskedApiKey: string;
  enabled: boolean;
  status: "healthy" | "unhealthy" | "unknown";
  lastCheckedAt: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface CreateConnectionInput {
  name: string;
  type: string;
  url: string;
  apiKey: string;
}

export interface UpdateConnectionInput {
  name: string;
  type: string;
  url: string;
  apiKey?: string;
  enabled?: boolean;
}

export interface TestResult {
  success: boolean;
  message?: string;
  appName?: string;
  version?: string;
}

async function fetchConnections(): Promise<Connection[]> {
  const res = await fetch("/api/connections");
  if (!res.ok) {
    throw new Error("Failed to fetch connections");
  }
  return res.json();
}

async function createConnection(
  input: CreateConnectionInput,
): Promise<Connection> {
  const res = await fetch("/api/connections", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!res.ok) {
    const data = await res.json();
    throw new Error(data.error || "Failed to create connection");
  }
  return res.json();
}

async function updateConnection({
  id,
  input,
}: {
  id: string;
  input: UpdateConnectionInput;
}): Promise<Connection> {
  const res = await fetch(`/api/connections/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!res.ok) {
    const data = await res.json();
    throw new Error(data.error || "Failed to update connection");
  }
  return res.json();
}

async function deleteConnection(id: string): Promise<void> {
  const res = await fetch(`/api/connections/${id}`, { method: "DELETE" });
  if (!res.ok) {
    throw new Error("Failed to delete connection");
  }
}

async function testSavedConnection(id: string): Promise<TestResult> {
  const res = await fetch(`/api/connections/${id}/test`, { method: "POST" });
  if (!res.ok) {
    const data = await res.json();
    throw new Error(data.error || "Failed to test connection");
  }
  return res.json();
}

async function testUnsavedConnection(input: {
  type: string;
  url: string;
  apiKey: string;
}): Promise<TestResult> {
  const res = await fetch("/api/connections/test", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!res.ok) {
    const data = await res.json();
    throw new Error(data.error || "Failed to test connection");
  }
  return res.json();
}

export function useConnections() {
  return useQuery({
    queryKey: ["connections"],
    queryFn: fetchConnections,
  });
}

export function useCreateConnection() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createConnection,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["connections"] });
    },
  });
}

export function useUpdateConnection() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: updateConnection,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["connections"] });
    },
  });
}

export function useDeleteConnection() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteConnection,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["connections"] });
    },
  });
}

export function useTestSavedConnection() {
  return useMutation({
    mutationFn: testSavedConnection,
  });
}

export function useTestUnsavedConnection() {
  return useMutation({
    mutationFn: testUnsavedConnection,
  });
}
