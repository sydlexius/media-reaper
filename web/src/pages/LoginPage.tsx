import { useState, type FormEvent } from "react";
import { useNavigate, useLocation } from "react-router";
import { useAuth } from "../queries/auth";

export function LoginPage() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const { login, loginError, isLoggingIn, isAuthenticated } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const from = (location.state as { from?: string })?.from || "/";

  if (isAuthenticated) {
    navigate(from, { replace: true });
    return null;
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    try {
      await login({ username, password });
      navigate(from, { replace: true });
    } catch {
      // Error is captured by loginError
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 dark:bg-gray-900">
      <div className="w-full max-w-sm rounded-lg bg-white p-8 shadow dark:bg-gray-800">
        <h1 className="mb-6 text-center text-2xl font-bold text-gray-900 dark:text-white">
          Media Reaper
        </h1>
        <form onSubmit={handleSubmit} className="space-y-4">
          {loginError && (
            <div className="rounded bg-red-100 p-3 text-sm text-red-700 dark:bg-red-900/50 dark:text-red-200">
              {loginError}
            </div>
          )}
          <div>
            <label
              htmlFor="username"
              className="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
            >
              Username
            </label>
            <input
              id="username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              autoFocus
              className="w-full rounded border border-gray-300 px-3 py-2 text-gray-900 focus:border-blue-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-white"
            />
          </div>
          <div>
            <label
              htmlFor="password"
              className="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
            >
              Password
            </label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              className="w-full rounded border border-gray-300 px-3 py-2 text-gray-900 focus:border-blue-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-white"
            />
          </div>
          <button
            type="submit"
            disabled={isLoggingIn}
            className="w-full rounded bg-blue-600 px-4 py-2 font-medium text-white hover:bg-blue-700 disabled:opacity-50"
          >
            {isLoggingIn ? "Signing in..." : "Sign in"}
          </button>
        </form>
      </div>
    </div>
  );
}
