"use client";

import { FormEvent, useEffect, useState } from "react";
import { API_URL, apiFetch } from "../lib/api";

type AuthResponse = {
  id: string;
  email: string;
  full_name: string;
  role: string;
  accessToken: string;
  refreshToken: string;
};

export default function Home() {
  const [mode, setMode] = useState<"login" | "register">("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [rememberMe, setRememberMe] = useState(false);

  useEffect(() => {
    const rememberedEmail = localStorage.getItem("rememberedEmail");
    const accessToken = localStorage.getItem("accessToken");
    const refreshToken = localStorage.getItem("refreshToken");
    const storedUser = localStorage.getItem("user");

    if (rememberedEmail) {
      setEmail(rememberedEmail);
      setRememberMe(true);
    }

    if (accessToken && refreshToken && storedUser) {
      try {
        const parsedUser: AuthResponse = JSON.parse(storedUser);

        if (
          parsedUser.id &&
          parsedUser.email &&
          parsedUser.full_name &&
          parsedUser.role
        ) {
          setUser(parsedUser);
        }
      } catch {
        localStorage.removeItem("accessToken");
        localStorage.removeItem("refreshToken");
        localStorage.removeItem("user");
      }
    }
  }, []);
  const [fullName, setFullName] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [user, setUser] = useState<AuthResponse | null>(null);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    setLoading(true);
    setError("");

    try {
      const endpoint =
        mode === "login"
          ? "/auth/login"
          : "/auth/register";

      const body =
        mode === "login"
          ? {
              email,
              password,
            }
          : {
              email,
              password,
              full_name: fullName,
            };

      const response = await apiFetch(`${API_URL}${endpoint}`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(body),
      });

      if (!response.ok) {
        const message = await response.text();
        throw new Error(
          message ||
            (mode === "login"
              ? "Giriş başarısız."
              : "Kayıt başarısız."),
        );
      }

      const data: AuthResponse = await response.json();

      if (mode === "register") {
        setMode("login");
        setPassword("");
        setError("Kayıt başarılı. Şimdi giriş yapabilirsiniz.");
        return;
      }

      if (rememberMe) {
        localStorage.setItem("rememberedEmail", email);
      } else {
        localStorage.removeItem("rememberedEmail");
      }

      localStorage.setItem("accessToken", data.accessToken);
      localStorage.setItem("refreshToken", data.refreshToken);
      localStorage.setItem("user", JSON.stringify(data));

      setUser(data);
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "İşlem sırasında bir hata oluştu.",
      );
    } finally {
      setLoading(false);
    }
  }

  async function handleLogout() {
    const refreshToken = localStorage.getItem("refreshToken");

    try {
      if (refreshToken) {
        await apiFetch(`${API_URL}/auth/logout`, {
          method: "POST",
          body: JSON.stringify({
            refreshToken,
          }),
        });
      }
    } finally {
      localStorage.removeItem("accessToken");
      localStorage.removeItem("refreshToken");
      localStorage.removeItem("user");

      setUser(null);
      setEmail("");
      setPassword("");
      setFullName("");
    }
  }

  if (user) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-slate-950 p-6 text-white">
        <div className="w-full max-w-lg rounded-2xl border border-slate-800 bg-slate-900 p-8 shadow-2xl">
          <p className="mb-2 text-sm font-medium text-blue-400">
            ENTERPRISE CORE
          </p>

          <h1 className="text-3xl font-bold">
            Hoş geldin, {user.full_name}
          </h1>

          <p className="mt-2 text-slate-400">
            Kimlik doğrulama başarılı.
          </p>

          <div className="mt-8 space-y-4 rounded-xl border border-slate-800 bg-slate-950 p-5">
            <div>
              <p className="text-xs text-slate-500">E-posta</p>
              <p className="mt-1 font-medium">{user.email}</p>
            </div>

            <div>
              <p className="text-xs text-slate-500">Rol</p>
              <p className="mt-1 font-medium">{user.role}</p>
            </div>

            <div>
              <p className="text-xs text-slate-500">Kullanıcı ID</p>
              <p className="mt-1 break-all font-mono text-sm">
                {user.id}
              </p>
            </div>
          </div>

          <button
            type="button"
            onClick={handleLogout}
            className="mt-6 w-full rounded-xl bg-red-600 px-4 py-3 font-semibold text-white transition hover:bg-red-500"
          >
            Çıkış Yap
          </button>
        </div>
      </main>
    );
  }

  return (
    <main className="flex min-h-screen items-center justify-center bg-slate-950 p-6 text-white">
      <div className="w-full max-w-md">
        <div className="mb-8 text-center">
          <p className="mb-3 text-sm font-semibold tracking-widest text-blue-400">
            ENTERPRISE CORE
          </p>

          <h1 className="text-4xl font-bold">
            {mode === "login"
              ? "Hesabına Giriş Yap"
              : "Yeni Hesap Oluştur"}
          </h1>

          <p className="mt-3 text-slate-400">
            {mode === "login"
              ? "Enterprise Core hesabına güvenli şekilde giriş yap."
              : "Enterprise Core hesabını oluştur."}
          </p>
        </div>

        <form
          onSubmit={handleSubmit}
          className="rounded-2xl border border-slate-800 bg-slate-900 p-8 shadow-2xl"
        >
          <div className="space-y-5">
            {mode === "register" && (
              <div>
                <label
                  htmlFor="fullName"
                  className="mb-2 block text-sm font-medium text-slate-300"
                >
                  Ad Soyad
                </label>

                <input
                  id="fullName"
                  type="text"
                  value={fullName}
                  onChange={(event) =>
                    setFullName(event.target.value)
                  }
                  placeholder="Ad Soyad"
                  required
                  autoComplete="name"
                  className="w-full rounded-xl border border-slate-700 bg-slate-950 px-4 py-3 text-white outline-none transition placeholder:text-slate-600 focus:border-blue-500"
                />
              </div>
            )}

            <div>
              <label
                htmlFor="email"
                className="mb-2 block text-sm font-medium text-slate-300"
              >
                E-posta
              </label>

              <input
                id="email"
                type="email"
                value={email}
                onChange={(event) =>
                  setEmail(event.target.value)
                }
                placeholder="ornek@email.com"
                required
                autoComplete="email"
                className="w-full rounded-xl border border-slate-700 bg-slate-950 px-4 py-3 text-white outline-none transition placeholder:text-slate-600 focus:border-blue-500"
              />
            </div>

            <div>
              <label
                htmlFor="password"
                className="mb-2 block text-sm font-medium text-slate-300"
              >
                Şifre
              </label>

              <input
                id="password"
                type="password"
                value={password}
                onChange={(event) =>
                  setPassword(event.target.value)
                }
                placeholder="••••••••"
                required
                autoComplete={
                  mode === "login"
                    ? "current-password"
                    : "new-password"
                }
                className="w-full rounded-xl border border-slate-700 bg-slate-950 px-4 py-3 text-white outline-none transition placeholder:text-slate-600 focus:border-blue-500"
              />
            </div>

            <label className="flex items-center gap-3 text-sm text-slate-300">
              <input
                type="checkbox"
                checked={rememberMe}
                onChange={(event) => setRememberMe(event.target.checked)}
                className="h-4 w-4 rounded border-slate-700 bg-slate-950 text-blue-600"
              />
              Beni Hatırla
            </label>
            {error && (
              <div className="rounded-xl border border-slate-700 bg-slate-950 px-4 py-3 text-sm text-slate-300">
                {error}
              </div>
            )}

            <button
              type="submit"
              disabled={loading}
              className="w-full rounded-xl bg-blue-600 px-4 py-3 font-semibold text-white transition hover:bg-blue-500 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {loading
                ? "İşlem yapılıyor..."
                : mode === "login"
                  ? "Giriş Yap"
                  : "Kayıt Ol"}
            </button>

            <button
              type="button"
              onClick={() => {
                setMode(
                  mode === "login"
                    ? "register"
                    : "login",
                );
                setError("");
                setPassword("");
              }}
              className="w-full rounded-xl border border-slate-700 px-4 py-3 font-semibold text-slate-300 transition hover:bg-slate-800"
            >
              {mode === "login"
                ? "Yeni Hesap Oluştur"
                : "Zaten Hesabım Var"}
            </button>
          </div>
        </form>
      </div>
    </main>
  );
}










