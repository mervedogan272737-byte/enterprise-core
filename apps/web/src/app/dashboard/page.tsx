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

export default function DashboardPage() {
  const router = useRouter();

  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function loadUser() {
      try {
        const response = await apiFetch(`${API_URL}/auth/me`);

        if (!response.ok) {
          router.replace("/");
          return;
        }

        const data: User = await response.json();

        setUser(data);
        localStorage.setItem("user", JSON.stringify(data));
      } catch {
        router.replace("/");
      } finally {
        setLoading(false);
      }
    }

    loadUser();
  }, [router]);

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

      router.replace("/");
    }
  }

  if (loading) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-slate-950 text-white">
        <p>Enterprise Core yükleniyor...</p>
      </main>
    );
  }

  if (!user) {
    return null;
  }

  return (
    <main className="min-h-screen bg-slate-950 text-white">
      <header className="border-b border-slate-800 bg-slate-900">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-6 py-5">
          <div>
            <p className="text-sm font-semibold tracking-widest text-blue-400">
              ENTERPRISE CORE
            </p>

            <h1 className="mt-1 text-xl font-bold">
              V9.9.9 Yönetim Paneli
            </h1>
          </div>

          <div className="flex items-center gap-5">
            <div className="text-right">
              <p className="font-medium">
                {user.full_name}
              </p>

              <p className="text-sm text-slate-400">
                {user.email}
              </p>
            </div>

            <button
              onClick={handleLogout}
              className="rounded-xl bg-red-600 px-4 py-2 text-sm font-semibold transition hover:bg-red-500"
            >
              Çıkış
            </button>
          </div>
        </div>
      </header>

      <section className="mx-auto max-w-7xl px-6 py-10">
        <div className="mb-8">
          <p className="text-sm font-semibold text-blue-400">
            ENTERPRISE CORE V9.9.9
          </p>

          <h2 className="mt-2 text-3xl font-bold">
            Hoş geldin, {user.full_name}
          </h2>

          <p className="mt-3 text-slate-400">
            Hesabın başarıyla doğrulandı. Yönetim merkezinden
            sistem özelliklerine erişebilirsin.
          </p>
        </div>

        <div className="grid gap-6 md:grid-cols-3">
          <div className="rounded-2xl border border-slate-800 bg-slate-900 p-6">
            <p className="text-sm text-slate-400">
              Hesap Durumu
            </p>

            <p className="mt-3 text-2xl font-bold text-green-400">
              Aktif
            </p>
          </div>

          <div className="rounded-2xl border border-slate-800 bg-slate-900 p-6">
            <p className="text-sm text-slate-400">
              Kullanıcı Rolü
            </p>

            <p className="mt-3 text-2xl font-bold">
              {user.role}
            </p>
          </div>

          <div className="rounded-2xl border border-slate-800 bg-slate-900 p-6">
            <p className="text-sm text-slate-400">
              Kimlik Doğrulama
            </p>

            <p className="mt-3 text-2xl font-bold text-blue-400">
              Güvenli
            </p>
          </div>
        </div>

        <div className="mt-8 grid gap-5 md:grid-cols-2 lg:grid-cols-4">
          <button
            onClick={() => router.push("/profile")}
            className="rounded-2xl border border-slate-800 bg-slate-900 p-6 text-left transition hover:border-blue-500 hover:bg-slate-800"
          >
            <p className="text-lg font-bold">
              Profilim
            </p>

            <p className="mt-2 text-sm text-slate-500">
              Hesap bilgilerini görüntüle
            </p>
          </button>

          <button
            onClick={() => router.push("/settings")}
            className="rounded-2xl border border-slate-800 bg-slate-900 p-6 text-left transition hover:border-blue-500 hover:bg-slate-800"
          >
            <p className="text-lg font-bold">
              Ayarlar
            </p>

            <p className="mt-2 text-sm text-slate-500">
              Sistem ve güvenlik ayarları
            </p>
          </button>

          <button
            onClick={() => router.push("/admin")}
            className="rounded-2xl border border-slate-800 bg-slate-900 p-6 text-left transition hover:border-red-500 hover:bg-slate-800"
          >
            <p className="text-lg font-bold">
              Admin Paneli
            </p>

            <p className="mt-2 text-sm text-slate-500">
              Yönetici merkezine eriş
            </p>
          </button>

          <button
            onClick={() => router.push("/")}
            className="rounded-2xl border border-slate-800 bg-slate-900 p-6 text-left transition hover:border-blue-500 hover:bg-slate-800"
          >
            <p className="text-lg font-bold">
              Ana Sayfa
            </p>

            <p className="mt-2 text-sm text-slate-500">
              Giriş ekranına dön
            </p>
          </button>
        </div>

        <div className="mt-8 rounded-2xl border border-slate-800 bg-slate-900 p-8">
          <h3 className="text-xl font-bold">
            Sistem Durumu
          </h3>

          <div className="mt-6 grid gap-4 md:grid-cols-4">
            <div className="rounded-xl bg-slate-950 p-5">
              <p className="text-sm text-slate-500">
                API
              </p>

              <p className="mt-2 font-semibold text-green-400">
                Online
              </p>
            </div>

            <div className="rounded-xl bg-slate-950 p-5">
              <p className="text-sm text-slate-500">
                Authentication
              </p>

              <p className="mt-2 font-semibold text-green-400">
                Aktif
              </p>
            </div>

            <div className="rounded-xl bg-slate-950 p-5">
              <p className="text-sm text-slate-500">
                Database
              </p>

              <p className="mt-2 font-semibold text-green-400">
                Bağlı
              </p>
            </div>

            <div className="rounded-xl bg-slate-950 p-5">
              <p className="text-sm text-slate-500">
                Redis
              </p>

              <p className="mt-2 font-semibold text-green-400">
                Aktif
              </p>
            </div>
          </div>
        </div>
      </section>
    </main>
  );
}
