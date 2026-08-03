"use client";

import { useRouter } from "next/navigation";

export default function SettingsPage() {
  const router = useRouter();

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
            Ayarlar
          </h1>

          <div className="mt-8 space-y-4">
            <div className="rounded-xl border border-slate-800 bg-slate-950 p-5">
              <p className="font-semibold">Hesap Ayarları</p>
              <p className="mt-2 text-sm text-slate-500">
                Hesap ve güvenlik seçenekleri.
              </p>
            </div>

            <div className="rounded-xl border border-slate-800 bg-slate-950 p-5">
              <p className="font-semibold">Bildirimler</p>
              <p className="mt-2 text-sm text-slate-500">
                Bildirim tercihleri yakında burada olacak.
              </p>
            </div>

            <div className="rounded-xl border border-slate-800 bg-slate-950 p-5">
              <p className="font-semibold">Güvenlik</p>
              <p className="mt-2 text-sm text-slate-500">
                Oturum ve güvenlik yönetimi.
              </p>
            </div>
          </div>
        </div>
      </div>
    </main>
  );
}
