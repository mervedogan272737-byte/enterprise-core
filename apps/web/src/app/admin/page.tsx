"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { API_URL, apiFetch } from "../../lib/api";

type User = {
  id: string;
  email: string;
  full_name: string;
  role: string;
};

export default function AdminPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [denied, setDenied] = useState(false);

  useEffect(() => {
    async function load() {
      const response = await apiFetch(`${API_URL}/auth/admin/me`);

      if (response.status === 401 || response.status === 403) {
        setDenied(true);
        return;
      }

      if (!response.ok) {
        router.replace("/");
        return;
      }

      setUser(await response.json());
    }

    load();
  }, [router]);

  if (denied) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-slate-950 p-6 text-white">
        <div className="rounded-2xl border border-red-900 bg-slate-900 p-8 text-center">
          <h1 className="text-2xl font-bold">
            Erişim Reddedildi
          </h1>

          <p className="mt-3 text-slate-400">
            Bu sayfaya erişmek için admin yetkisi gerekiyor.
          </p>

          <button
            onClick={() => router.push("/dashboard")}
            className="mt-6 rounded-xl bg-blue-600 px-5 py-3 font-semibold"
          >
            Dashboard'a Dön
          </button>
        </div>
      </main>
    );
  }

  if (!user) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-slate-950 text-white">
        Yükleniyor...
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-slate-950 p-6 text-white">
      <div className="mx-auto max-w-5xl">
        <button
          onClick={() => router.push("/dashboard")}
          className="mb-6 text-sm text-blue-400"
        >
          ← Dashboard
        </button>

        <div className="rounded-2xl border border-slate-800 bg-slate-900 p-8">
          <p className="text-sm font-semibold tracking-widest text-red-400">
            ADMIN PANEL
          </p>

          <h1 className="mt-2 text-3xl font-bold">
            Yönetici Merkezi
          </h1>

          <p className="mt-3 text-slate-400">
            Admin hesabı başarıyla doğrulandı.
          </p>

          <div className="mt-8 grid gap-6 md:grid-cols-3">
            <div className="rounded-xl border border-slate-800 bg-slate-950 p-6">
              <p className="text-sm text-slate-500">Admin</p>
              <p className="mt-2 font-semibold">{user.full_name}</p>
            </div>

            <div className="rounded-xl border border-slate-800 bg-slate-950 p-6">
              <p className="text-sm text-slate-500">Rol</p>
              <p className="mt-2 font-semibold">{user.role}</p>
            </div>

            <div className="rounded-xl border border-slate-800 bg-slate-950 p-6">
              <p className="text-sm text-slate-500">Sistem</p>
              <p className="mt-2 font-semibold text-green-400">
                Aktif
              </p>
            </div>
          </div>
        </div>
      </div>
    </main>
  );
}
