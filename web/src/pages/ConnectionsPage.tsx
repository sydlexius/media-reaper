import { useState } from "react";
import {
  useConnections,
  useCreateConnection,
  useUpdateConnection,
  useDeleteConnection,
  useTestSavedConnection,
  type Connection,
  type TestResult,
} from "../queries/connections";
import { ConnectionCard } from "../components/ConnectionCard";
import { ConnectionModal } from "../components/ConnectionModal";
import { ConfirmDialog } from "../components/ConfirmDialog";

export function ConnectionsPage() {
  const { data: connections, isLoading } = useConnections();
  const createMutation = useCreateConnection();
  const updateMutation = useUpdateConnection();
  const deleteMutation = useDeleteConnection();
  const testMutation = useTestSavedConnection();

  const [modalOpen, setModalOpen] = useState(false);
  const [editingConnection, setEditingConnection] =
    useState<Connection | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Connection | null>(null);
  const [testingId, setTestingId] = useState<string | null>(null);
  const [testResults, setTestResults] = useState<
    Record<string, TestResult>
  >({});

  function handleAdd() {
    setEditingConnection(null);
    setModalOpen(true);
  }

  function handleEdit(conn: Connection) {
    setEditingConnection(conn);
    setModalOpen(true);
  }

  async function handleSave(data: {
    name: string;
    type: string;
    url: string;
    apiKey: string;
    enabled?: boolean;
  }) {
    try {
      if (editingConnection) {
        await updateMutation.mutateAsync({
          id: editingConnection.id,
          input: {
            name: data.name,
            type: data.type,
            url: data.url,
            apiKey: data.apiKey || undefined,
            enabled: data.enabled,
          },
        });
      } else {
        await createMutation.mutateAsync({
          name: data.name,
          type: data.type,
          url: data.url,
          apiKey: data.apiKey,
        });
      }
      setModalOpen(false);
    } catch {
      // Error handled by mutation state
    }
  }

  async function handleTest(conn: Connection) {
    setTestingId(conn.id);
    try {
      const result = await testMutation.mutateAsync(conn.id);
      setTestResults((prev) => ({ ...prev, [conn.id]: result }));
    } catch {
      setTestResults((prev) => ({
        ...prev,
        [conn.id]: { success: false, message: "Test request failed" },
      }));
    } finally {
      setTestingId(null);
    }
  }

  async function handleDelete() {
    if (!deleteTarget) return;
    try {
      await deleteMutation.mutateAsync(deleteTarget.id);
      setDeleteTarget(null);
    } catch {
      // Error handled by mutation state
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-gray-500 dark:text-gray-400">
          Loading connections...
        </p>
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
          Connections
        </h1>
        <button
          onClick={handleAdd}
          className="rounded bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          Add Connection
        </button>
      </div>

      {connections && connections.length > 0 ? (
        <div className="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {connections.map((conn) => (
            <ConnectionCard
              key={conn.id}
              connection={conn}
              onEdit={() => handleEdit(conn)}
              onTest={() => handleTest(conn)}
              onDelete={() => setDeleteTarget(conn)}
              isTesting={testingId === conn.id}
              testResult={testResults[conn.id] || null}
            />
          ))}
        </div>
      ) : (
        <div className="mt-12 text-center">
          <p className="text-gray-500 dark:text-gray-400">
            No connections configured yet.
          </p>
          <p className="mt-1 text-sm text-gray-400 dark:text-gray-500">
            Add a Sonarr, Radarr, or Emby connection to get started.
          </p>
        </div>
      )}

      <ConnectionModal
        key={editingConnection?.id ?? "new"}
        open={modalOpen}
        connection={editingConnection}
        onSave={handleSave}
        onCancel={() => setModalOpen(false)}
        isSaving={createMutation.isPending || updateMutation.isPending}
      />

      <ConfirmDialog
        open={deleteTarget !== null}
        title="Delete Connection"
        message={`Are you sure you want to delete "${deleteTarget?.name}"? This action cannot be undone.`}
        onConfirm={handleDelete}
        onCancel={() => setDeleteTarget(null)}
      />
    </div>
  );
}
