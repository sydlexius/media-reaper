import { useState, type FormEvent } from "react";
import {
  useTestUnsavedConnection,
  type Connection,
  type TestResult,
} from "../queries/connections";

interface ConnectionModalProps {
  open: boolean;
  connection: Connection | null;
  onSave: (data: {
    name: string;
    type: string;
    url: string;
    apiKey: string;
    enabled?: boolean;
  }) => void;
  onCancel: () => void;
  isSaving: boolean;
}

export function ConnectionModal({
  open,
  connection,
  onSave,
  onCancel,
  isSaving,
}: ConnectionModalProps) {
  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <ConnectionForm
        connection={connection}
        onSave={onSave}
        onCancel={onCancel}
        isSaving={isSaving}
      />
    </div>
  );
}

interface ConnectionFormProps {
  connection: Connection | null;
  onSave: (data: {
    name: string;
    type: string;
    url: string;
    apiKey: string;
    enabled?: boolean;
  }) => void;
  onCancel: () => void;
  isSaving: boolean;
}

function ConnectionForm({
  connection,
  onSave,
  onCancel,
  isSaving,
}: ConnectionFormProps) {
  const [name, setName] = useState(connection?.name ?? "");
  const [type, setType] = useState<string>(connection?.type ?? "sonarr");
  const [url, setUrl] = useState(connection?.url ?? "");
  const [apiKey, setApiKey] = useState("");
  const [enabled, setEnabled] = useState(connection?.enabled ?? true);
  const [testResult, setTestResult] = useState<TestResult | null>(null);

  const testMutation = useTestUnsavedConnection();
  const isEditing = connection !== null;

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    onSave({
      name,
      type,
      url,
      apiKey,
      enabled: isEditing ? enabled : undefined,
    });
  }

  async function handleTest() {
    if (!url || !apiKey) return;
    setTestResult(null);
    try {
      const result = await testMutation.mutateAsync({ type, url, apiKey });
      setTestResult(result);
    } catch {
      setTestResult({ success: false, message: "Test request failed" });
    }
  }

  const placeholderUrls: Record<string, string> = {
    sonarr: "http://localhost:8989",
    radarr: "http://localhost:7878",
    emby: "http://localhost:8096",
  };

  return (
    <div className="w-full max-w-lg rounded-lg bg-white p-6 shadow-xl dark:bg-gray-800">
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
        {isEditing ? "Edit Connection" : "Add Connection"}
      </h3>

      <form onSubmit={handleSubmit} className="mt-4 space-y-4">
        <div>
          <label
            htmlFor="conn-name"
            className="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
          >
            Name
          </label>
          <input
            id="conn-name"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            className="w-full rounded border border-gray-300 px-3 py-2 text-gray-900 focus:border-blue-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-white"
            placeholder="My Sonarr"
          />
        </div>

        <div>
          <label
            htmlFor="conn-type"
            className="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
          >
            Type
          </label>
          <select
            id="conn-type"
            value={type}
            onChange={(e) => setType(e.target.value)}
            className="w-full rounded border border-gray-300 px-3 py-2 text-gray-900 focus:border-blue-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-white"
          >
            <option value="sonarr">Sonarr</option>
            <option value="radarr">Radarr</option>
            <option value="emby">Emby</option>
          </select>
        </div>

        <div>
          <label
            htmlFor="conn-url"
            className="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
          >
            URL
          </label>
          <input
            id="conn-url"
            type="url"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            required
            className="w-full rounded border border-gray-300 px-3 py-2 text-gray-900 focus:border-blue-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-white"
            placeholder={placeholderUrls[type] || "http://localhost:8080"}
          />
        </div>

        <div>
          <label
            htmlFor="conn-apikey"
            className="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
          >
            API Key
          </label>
          <input
            id="conn-apikey"
            type="password"
            value={apiKey}
            onChange={(e) => setApiKey(e.target.value)}
            required={!isEditing}
            className="w-full rounded border border-gray-300 px-3 py-2 text-gray-900 focus:border-blue-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-white"
            placeholder={
              isEditing ? "Leave blank to keep current" : "Enter API key"
            }
          />
        </div>

        {isEditing && (
          <div className="flex items-center gap-2">
            <input
              id="conn-enabled"
              type="checkbox"
              checked={enabled}
              onChange={(e) => setEnabled(e.target.checked)}
              className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
            />
            <label
              htmlFor="conn-enabled"
              className="text-sm font-medium text-gray-700 dark:text-gray-300"
            >
              Enabled
            </label>
          </div>
        )}

        {testResult && (
          <div
            className={`rounded p-3 text-sm ${
              testResult.success
                ? "bg-green-100 text-green-700 dark:bg-green-900/50 dark:text-green-200"
                : "bg-red-100 text-red-700 dark:bg-red-900/50 dark:text-red-200"
            }`}
          >
            {testResult.success
              ? `Connected: ${testResult.appName} v${testResult.version}`
              : `Failed: ${testResult.message}`}
          </div>
        )}

        <div className="flex justify-between pt-2">
          <button
            type="button"
            onClick={handleTest}
            disabled={!url || !apiKey || testMutation.isPending}
            className="rounded border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700"
          >
            {testMutation.isPending ? "Testing..." : "Test Connection"}
          </button>

          <div className="flex gap-3">
            <button
              type="button"
              onClick={onCancel}
              className="rounded border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isSaving}
              className="rounded bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
            >
              {isSaving ? "Saving..." : "Save"}
            </button>
          </div>
        </div>
      </form>
    </div>
  );
}
