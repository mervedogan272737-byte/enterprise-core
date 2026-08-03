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

export default function ProfilePage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    async function load() {
      const response = await apiFetch(`${API_URL}/auth/me`);

      if (!response.ok) {
        router.replace("/");
        return;
      }

      setUser(await response.json());
    }

    load();
  }, [router]);

  if (!user) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-slate-950 text-white">
        Yükleniyor...
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-slate-950 p-6 text-white">
      <div className="mx-auto max-w-3xl">
        <button
          onClick={() => router.push("/dashboard")}
          className="mb-6 text-sm text-blue-400 hover:text-blue-300"
        >
          ← Dashboard
        </button>

        <div className="rounded-2xl border border-slate-800 bg-slate-900 p-8">
          <p className="text-sm font-semibold tracking-widest text-blue-400">
            ENTERPRISE CORE
          </p>

          <h1 className="mt-2 text-3xl font-bold">
            Profilim
          </h1>

          <div className="mt-8 space-y-5">
            <div>
              <p className="text-sm text-slate-500">Ad Soyad</p>
              <p className="mt-1 text-lg">{user.full_name}</p>
            </div>

            <div>
              <p className="text-sm text-slate-500">E-posta</p>
              <p className="mt-1 text-lg">{user.email}</p>
            </div>

            <div>
              <p className="text-sm text-slate-500">Rol</p>
              <p className="mt-1 text-lg">{user.role}</p>
            </div>

            <div>
              <p className="text-sm text-slate-500">Kullanıcı ID</p>
              <p className="mt-1 break-all font-mono text-sm text-slate-400">
                {user.id}
              </p>
            </div>
          </div>
        </div>
      </div>
    </main>
  );
}
