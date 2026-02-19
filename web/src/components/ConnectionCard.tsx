import { type Connection } from "../queries/connections";

interface ConnectionCardProps {
  connection: Connection;
  onEdit: () => void;
  onTest: () => void;
  onDelete: () => void;
  isTesting: boolean;
  testResult: { success: boolean; message?: string; appName?: string; version?: string } | null;
}

const typeBadges: Record<string, { label: string; className: string }> = {
  sonarr: {
    label: "S",
    className: "bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-200",
  },
  radarr: {
    label: "R",
    className: "bg-orange-100 text-orange-700 dark:bg-orange-900 dark:text-orange-200",
  },
  emby: {
    label: "E",
    className: "bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-200",
  },
};

const statusDots: Record<string, string> = {
  healthy: "bg-green-500",
  unhealthy: "bg-red-500",
  unknown: "bg-gray-400",
};

export function ConnectionCard({
  connection,
  onEdit,
  onTest,
  onDelete,
  isTesting,
  testResult,
}: ConnectionCardProps) {
  const badge = typeBadges[connection.type] || typeBadges.emby;
  const statusDot = statusDots[connection.status] || statusDots.unknown;

  return (
    <div className="rounded-lg border border-gray-200 bg-white p-4 shadow-sm dark:border-gray-700 dark:bg-gray-800">
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3">
          <span
            className={`flex h-8 w-8 items-center justify-center rounded-md text-sm font-bold ${badge.className}`}
          >
            {badge.label}
          </span>
          <div>
            <div className="flex items-center gap-2">
              <h3 className="font-medium text-gray-900 dark:text-white">
                {connection.name}
              </h3>
              <span
                className={`inline-block h-2.5 w-2.5 rounded-full ${statusDot}`}
                title={connection.status}
              />
              {!connection.enabled && (
                <span className="rounded bg-gray-100 px-1.5 py-0.5 text-xs text-gray-500 dark:bg-gray-700 dark:text-gray-400">
                  Disabled
                </span>
              )}
            </div>
            <p className="text-sm text-gray-500 dark:text-gray-400">
              {connection.url}
            </p>
          </div>
        </div>
      </div>

      <div className="mt-3 text-xs text-gray-400 dark:text-gray-500">
        API Key: {connection.maskedApiKey}
      </div>

      {testResult && (
        <div
          className={`mt-3 rounded p-2 text-xs ${
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

      <div className="mt-3 flex gap-2 border-t border-gray-100 pt-3 dark:border-gray-700">
        <button
          onClick={onEdit}
          className="rounded px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700"
        >
          Edit
        </button>
        <button
          onClick={onTest}
          disabled={isTesting}
          className="rounded px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100 disabled:opacity-50 dark:text-gray-300 dark:hover:bg-gray-700"
        >
          {isTesting ? "Testing..." : "Test"}
        </button>
        <button
          onClick={onDelete}
          className="rounded px-3 py-1.5 text-sm text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20"
        >
          Delete
        </button>
      </div>
    </div>
  );
}
